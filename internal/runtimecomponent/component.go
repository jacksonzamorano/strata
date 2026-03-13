package runtimecomponent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jacksonzamorano/strata/component"
	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type ContainerAccess interface {
	GetStorage() core.Storage
	GetKeychain() core.Keychain
	HasPermission(core.PermissionAction, string) bool
	ReadFile(string) ([]byte, bool)
	TemporaryFile() string
	Namespace() string
}

type Runner struct {
	transport     *componentipc.IO
	hostTransport *hostio.IO
	container     ContainerAccess
	available     bool
	path          string
	context       context.Context
	cancel        context.CancelFunc
	terminal      terminalProvider
	triggers      chan componentipc.ComponentMessageSendTrigger
	logger        core.Logger
}

func Register(dep *core.ComponentExecuteCommand, storageDir, tempDir string) (*Runner, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd, err := core.PlatformSandboxProvider().Execute(ctx, storageDir, tempDir, dep)
	if err != nil {
		cancel()
		return nil, err
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	transport := componentipc.NewIO(ctx, cancel, out, in)

	err = cmd.Start()
	if err != nil {
		cancel()
		return nil, err
	}

	runner := &Runner{
		transport: transport,
		available: false,
		path:      dep.CanonicalName,
		context:   ctx,
		cancel:    cancel,
		triggers:  make(chan componentipc.ComponentMessageSendTrigger, 64),
		terminal:  newTerminalProvider(),
	}

	return runner, nil
}

func (cr *Runner) Transport() *componentipc.IO {
	return cr.transport
}

func (cr *Runner) Path() string {
	return cr.path
}

func (cr *Runner) Available() bool {
	return cr.available
}

func (cr *Runner) SetAvailable(v bool) {
	cr.available = v
}

func (cr *Runner) TriggerChannel() <-chan componentipc.ComponentMessageSendTrigger {
	return cr.triggers
}

func (cr *Runner) Begin(cnt ContainerAccess, hostTransport *hostio.IO, logger core.Logger) {
	cr.container = cnt
	cr.hostTransport = hostTransport
	cr.logger = logger
	cr.handleAPIRequests()
}

func (cr *Runner) Execute(fname string, args any) *component.ComponentResultPayload {
	thread := cr.transport.NewThread()
	enc, _ := json.Marshal(args)
	payload, _ := componentipc.SendAndReceive[component.ComponentResultPayload](
		thread,
		componentipc.ComponentMessageTypeExecute,
		componentipc.ComponentMessageExecute{Name: fname, Arguments: enc},
		componentipc.ComponentMessageTypeRet,
	)

	return &payload
}

func (cr *Runner) requestSecretAuth(ev componentipc.ReceivedEvent[componentipc.ComponentMessageRequestSecretAuthentication]) {
	hostThread := cr.hostTransport.NewThread()
	hostRes, _ := hostio.SendAndReceive[hostio.HostMessageCompleteSecret](
		hostThread,
		hostio.HostMessageTypeRequestSecret,
		hostio.HostMessageRequestSecret{
			Namespace: cr.container.Namespace(),
			Prompt:    ev.Payload.Prompt,
		},
		hostio.HostMessageTypeCompleteSecret,
	)
	ev.Thread.Send(
		componentipc.ComponentMessageTypeCompleteSecretAuthentication,
		componentipc.ComponentMessageCompleteSecretAuthentication{
			Secret: hostRes.Secret,
		},
	)
}

func (cr *Runner) requestOauthAuth(ev componentipc.ReceivedEvent[componentipc.ComponentMessageRequestOauthAuthentication]) {
	hostThread := cr.hostTransport.NewThread()
	hostRes, _ := hostio.SendAndReceive[hostio.HostMessageCompleteOauth](
		hostThread,
		hostio.HostMessageTypeRequestSecret,
		hostio.HostMessageRequestOauth{
			Namespace:   cr.container.Namespace(),
			Url:         ev.Payload.Url,
			Destination: ev.Payload.Callback,
		},
		hostio.HostMessageTypeCompleteSecret,
	)
	ev.Thread.Send(
		componentipc.ComponentMessageTypeCompleteOauthAuthentication,
		componentipc.ComponentMessageCompleteOauthAuthentication{
			Url: hostRes.Url,
		},
	)
}

func (cr *Runner) executeCommandRequest(ev componentipc.ReceivedEvent[componentipc.ComponentMessageExecuteProgramRequest]) {
	if ev.Payload.Background {
		cr.executeBackgroundRequest(ev)
		return
	}
	pm := fmt.Sprintf("%s %s", ev.Payload.Program, strings.Join(ev.Payload.Arguments, " "))

	if !cr.container.HasPermission(core.PermissionActionExecuteCommandLine, pm) {
		ev.Thread.Send(componentipc.ComponentMessageTypeExecuteProgramResponse, componentipc.ComponentMessageExecuteProgramResponse{
			Error: "Permission denied.",
			Ok:    false,
		})
		return
	}

	result := cr.terminal.RunInDirectory(
		time.Minute*2,
		ev.Payload.WorkingDirectory,
		ev.Payload.Program,
		ev.Payload.Arguments...,
	)
	if result.Ok {
		ev.Thread.Send(componentipc.ComponentMessageTypeExecuteProgramResponse, componentipc.ComponentMessageExecuteProgramResponse{
			Output: result.Output,
			Code:   result.Code,
			Ok:     true,
		})
	} else {
		ev.Thread.Send(componentipc.ComponentMessageTypeExecuteProgramResponse, componentipc.ComponentMessageExecuteProgramResponse{
			Error: result.Error + ": " + result.Output,
			Code:  result.Code,
			Ok:    false,
		})
	}
}
func (cr *Runner) executeBackgroundRequest(ev componentipc.ReceivedEvent[componentipc.ComponentMessageExecuteProgramRequest]) {
	pm := fmt.Sprintf("%s %s", ev.Payload.Program, strings.Join(ev.Payload.Arguments, " "))

	if !cr.container.HasPermission(core.PermissionActionExecuteDaemon, pm) {
		ev.Thread.Send(componentipc.ComponentMessageTypeExecuteProgramStartedResponse, componentipc.ComponentMessageExecuteProgramStartedResponse{
			Error: "Permission denied.",
			Ok:    false,
		})
		return
	}
	ev.Thread.Send(componentipc.ComponentMessageTypeExecuteProgramStartedResponse, componentipc.ComponentMessageExecuteProgramStartedResponse{
		Ok: true,
	})

	result := cr.terminal.RunInDirectoryWithContext(
		cr.context,
		ev.Payload.WorkingDirectory,
		ev.Payload.Program,
		ev.Payload.Arguments...,
	)
	if result.Ok {
		ev.Thread.Send(componentipc.ComponentMessageTypeExecuteProgramResponse, componentipc.ComponentMessageExecuteProgramResponse{
			Output: result.Output,
			Code:   result.Code,
			Ok:     true,
		})
	} else {
		ev.Thread.Send(componentipc.ComponentMessageTypeExecuteProgramResponse, componentipc.ComponentMessageExecuteProgramResponse{
			Error: result.Error + ": " + result.Output,
			Code:  result.Code,
			Ok:    false,
		})
	}
}

func (cr *Runner) launchUrlRequest(ev componentipc.ReceivedEvent[componentipc.ComponentMessageLaunchUrlRequest]) {
	if !cr.container.HasPermission(core.PermissionActionLaunchUrl, ev.Payload.Url) {
		ev.Thread.Send(componentipc.ComponentMessageTypeLaunchUrlResponse, componentipc.ComponentMessageLaunchUrlResponse{
			Completed: false,
		})
		return
	}

	res := cr.terminal.OpenUrl(ev.Payload.Url)
	ev.Thread.Send(componentipc.ComponentMessageTypeLaunchUrlResponse, componentipc.ComponentMessageLaunchUrlResponse{
		Completed: res,
	})
}

func (cr *Runner) readFile(ev componentipc.ReceivedEvent[componentipc.ComponentMessageReadFileRequest]) {
	buf, ok := cr.container.ReadFile(ev.Payload.Path)
	if !ok {
		ev.Thread.Send(componentipc.ComponentMessageTypeReadFileResponse, componentipc.ComponentMessageReadFileResponse{
			Succeeded: false,
		})
		return
	}

	if len(buf) > core.IPC_MAX_BUF_SIZE {
		bufF := cr.container.TemporaryFile()
		os.WriteFile(bufF, buf, 0755)
		ev.Thread.Send(componentipc.ComponentMessageTypeReadFileResponse, componentipc.ComponentMessageReadFileResponse{
			Succeeded: true,
			Path:      bufF,
		})
		return
	}

	ev.Thread.Send(componentipc.ComponentMessageTypeReadFileResponse, componentipc.ComponentMessageReadFileResponse{
		Succeeded: true,
		Contents:  buf,
	})
}

func (cr *Runner) handleAPIRequests() {
	go func() {
		getVal := componentipc.Receive[componentipc.ComponentMessageGetValueRequest](cr.transport, componentipc.ComponentMessageTypeGetValueRequest)
		setVal := componentipc.Receive[componentipc.ComponentMessageSetValueRequest](cr.transport, componentipc.ComponentMessageTypeStoreValueRequest)
		getKeychain := componentipc.Receive[componentipc.ComponentMessageGetKeychainRequest](cr.transport, componentipc.ComponentMessageTypeGetKeychainRequest)
		setKeychain := componentipc.Receive[componentipc.ComponentMessageSetKeychainRequest](cr.transport, componentipc.ComponentMessageTypeStoreKeychainRequest)
		log := componentipc.Receive[componentipc.ComponentMessageLog](cr.transport, componentipc.ComponentMessageTypeLog)
		trigger := componentipc.Receive[componentipc.ComponentMessageSendTrigger](cr.transport, componentipc.ComponentMessageTypeSendTrigger)
		secretRequest := componentipc.Receive[componentipc.ComponentMessageRequestSecretAuthentication](cr.transport, componentipc.ComponentMessageTypeRequestSecretAuthentication)
		oauthRequest := componentipc.Receive[componentipc.ComponentMessageRequestOauthAuthentication](cr.transport, componentipc.ComponentMessageTypeRequestOauthAuthentication)
		executeCommandRequest := componentipc.Receive[componentipc.ComponentMessageExecuteProgramRequest](cr.transport, componentipc.ComponentMessageTypeExecuteProgramRequest)
		launchUrlRequest := componentipc.Receive[componentipc.ComponentMessageLaunchUrlRequest](cr.transport, componentipc.ComponentMessageTypeLaunchUrlRequest)
		readFile := componentipc.Receive[componentipc.ComponentMessageReadFileRequest](cr.transport, componentipc.ComponentMessageTypeReadFileRequest)
		for {
			select {
			case ev := <-getVal:
				if ev.Error {
					return
				}
				ev.Thread.Send(componentipc.ComponentMessageTypeGetValueResponse, componentipc.ComponentMessageGetValueResponse{
					Value: cr.container.GetStorage().GetString(ev.Payload.Key),
				})
			case ev := <-setVal:
				if ev.Error {
					return
				}
				cr.container.GetStorage().SetString(ev.Payload.Key, ev.Payload.Value)
			case ev := <-getKeychain:
				if ev.Error {
					return
				}
				ev.Thread.Send(componentipc.ComponentMessageTypeGetKeychainResponse, componentipc.ComponentMessageGetKeychainResponse{
					Value: cr.container.GetKeychain().Get(ev.Payload.Key),
				})
			case ev := <-setKeychain:
				if ev.Error {
					return
				}
				cr.container.GetKeychain().Set(ev.Payload.Key, ev.Payload.Value)
			case ev := <-secretRequest:
				go cr.requestSecretAuth(ev)
			case ev := <-oauthRequest:
				go cr.requestOauthAuth(ev)
			case ev := <-log:
				if ev.Error {
					return
				}
				cr.logger.Log("%s", ev.Payload.Message)
			case ev := <-trigger:
				cr.triggers <- ev.Payload
			case ev := <-executeCommandRequest:
				go cr.executeCommandRequest(ev)
			case ev := <-launchUrlRequest:
				go cr.launchUrlRequest(ev)
			case ev := <-readFile:
				go cr.readFile(ev)
			case <-cr.context.Done():
				return
			}
		}
	}()
}
