package strata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jacksonzamorano/strata/component"
	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type ComponentTrigger struct {
	Namespace string
	Name      string
	Trigger   func(b []byte)
}

type ComponentIO struct {
	transport   *componentipc.IO
	hostService *hostio.IO
	container   *Container
	available   bool
	path        string
	context     context.Context
	cancel      context.CancelFunc
	terminal    TerminalProvider
	triggers    chan componentipc.ComponentMessageSendTrigger
	logger      core.Logger
}

func RegisterComponent(dep *core.ComponentExecuteCommand) (*ComponentIO, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd, err := core.PlatformSandboxProvider().Execute(ctx, dep)
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

	runner := &ComponentIO{
		transport: transport,
		available: false,
		path:      dep.CanonicalName,
		context:   ctx,
		cancel:    cancel,
		triggers:  make(chan componentipc.ComponentMessageSendTrigger, 64),
		terminal:  TerminalProvider{terminal: &NativeTerminal{}},
	}

	return runner, nil
}

func (cr *ComponentIO) Begin(cnt *Container, logger core.Logger) {
	cr.container = cnt
	cr.hostService = cnt.hostService.host
	cr.logger = logger
	cr.HandleAPIRequests()
}

func (cr *ComponentIO) Execute(fname string, args any) *component.ComponentResultPayload {
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

func (cr *ComponentIO) requestSecretAuth(ev componentipc.ReceivedEvent[componentipc.ComponentMessageRequestSecretAuthentication]) {
	hostThread := cr.hostService.NewThread()
	hostRes, _ := hostio.SendAndReceive[hostio.HostMessageCompleteSecret](
		hostThread,
		hostio.HostMessageTypeRequestSecret,
		hostio.HostMessageRequestSecret{
			Namespace: cr.container.namespace,
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

func (cr *ComponentIO) requestOauthAuth(ev componentipc.ReceivedEvent[componentipc.ComponentMessageRequestOauthAuthentication]) {
	hostThread := cr.hostService.NewThread()
	hostRes, _ := hostio.SendAndReceive[hostio.HostMessageCompleteOauth](
		hostThread,
		hostio.HostMessageTypeRequestSecret,
		hostio.HostMessageRequestOauth{
			Namespace:   cr.container.namespace,
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

func (cr *ComponentIO) executeCommandRequest(ev componentipc.ReceivedEvent[componentipc.ComponentMessageExecuteProgramRequest]) {
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

func (cr *ComponentIO) launchUrlRequest(ev componentipc.ReceivedEvent[componentipc.ComponentMessageLaunchUrlRequest]) {
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

func (cr *ComponentIO) HandleAPIRequests() {
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
		for {
			select {
			case ev := <-getVal:
				if ev.Error {
					return
				}
				ev.Thread.Send(componentipc.ComponentMessageTypeGetValueResponse, componentipc.ComponentMessageGetValueResponse{
					Value: cr.container.Storage.GetString(ev.Payload.Key),
				})
			case ev := <-setVal:
				if ev.Error {
					return
				}
				cr.container.Storage.SetString(ev.Payload.Key, ev.Payload.Value)
			case ev := <-getKeychain:
				if ev.Error {
					return
				}
				ev.Thread.Send(componentipc.ComponentMessageTypeGetKeychainResponse, componentipc.ComponentMessageGetKeychainResponse{
					Value: cr.container.Keychain.Get(ev.Payload.Key),
				})
			case ev := <-setKeychain:
				if ev.Error {
					return
				}
				cr.container.Keychain.Set(ev.Payload.Key, ev.Payload.Value)
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
			case <-cr.context.Done():
				return
			}
		}
	}()
}
