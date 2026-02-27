package strata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/jacksonzamorano/tasks/strata/component"
	"github.com/jacksonzamorano/tasks/strata/internal/componentipc"
)

type ComponentRunner struct {
	transport *componentipc.IO
	container *Container
	available bool
	path      string
	context   context.Context
	cancel    context.CancelFunc
}

func runGit(p string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = p
	txt, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(txt))
	}
	return nil
}

func checkoutGit(url, ref, subdir string) (string, error) {
	tmp, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	checkout := path.Join(tmp, "com.jacksonzamorano.tasks", path.Base(url))

	_, err = os.Stat(checkout)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := runGit(tmp, "clone", url, checkout)
			if err != nil {
				return checkout, err
			}
		} else {
			return checkout, err
		}
	}

	runGit(checkout, "pull")

	if len(ref) == 0 {
		err = runGit(checkout, "switch", ref)
		if err != nil {
			return checkout, err
		}
	}

	return path.Join(checkout, subdir), nil
}

func RegisterComponent(dep AppDependancy, container *Container) (*ComponentRunner, error) {
	var cmd_path string
	var args []string
	var cwd_path string
	var display_path string

	switch dep.dep_type {
	case AppDependancyTypeBinary:
		cmd_path = dep.url
		display_path = dep.url
	case AppDependancyTypeLocalProject:
		cmd_path = "go"
		cwd_path = dep.url
		args = []string{"run", "."}
		display_path = cwd_path
	case AppDependancyTypeGit:
		p, err := checkoutGit(dep.url, dep.branch, dep.subdir)
		if err != nil {
			return nil, err
		}
		display_path = p
		cmd_path = "go"
		cwd_path = p
		args = []string{"run", "."}
	}

	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, cmd_path, args...)
	if len(cwd_path) > 0 {
		cmd.Dir = cwd_path
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

	runner := &ComponentRunner{
		transport: transport,
		container: container,
		available: false,
		path:      display_path,
		context:   ctx,
		cancel:    cancel,
	}

	runner.HandleAPIRequests()

	return runner, nil
}

func (cr *ComponentRunner) Execute(fname string, args any) *component.ComponentResultPayload {
	thread := cr.transport.NewThread()
	enc, _ := json.Marshal(args)
	payload, _ := componentipc.SendAndReceive[component.ComponentResultPayload](
		thread,
		componentipc.MessageTypeExecute,
		componentipc.ComponentMessageExecute{Name: fname, Arguments: enc},
		componentipc.MessageTypeRet,
	)

	return &payload
}

func (cr *ComponentRunner) HandleAPIRequests() {
	go func() {
		getVal := componentipc.Receive[componentipc.ComponentMessageGetValueRequest](cr.transport, componentipc.MessageTypeGetValueRequest)
		setVal := componentipc.Receive[componentipc.ComponentMessageSetValueRequest](cr.transport, componentipc.MessageTypeStoreValueRequest)
		getKeychain := componentipc.Receive[componentipc.ComponentMessageGetKeychainRequest](cr.transport, componentipc.MessageTypeGetKeychainRequest)
		setKeychain := componentipc.Receive[componentipc.ComponentMessageSetKeychainRequest](cr.transport, componentipc.MessageTypeStoreKeychainRequest)
		log := componentipc.Receive[componentipc.ComponentMessageLog](cr.transport, componentipc.MessageTypeLog)
		for {
			select {
			case ev := <-getVal:
				if ev.Error {
					return
				}
				ev.Thread.Send(componentipc.MessageTypeGetValueResponse, componentipc.ComponentMessageGetValueResponse{
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
				ev.Thread.Send(componentipc.MessageTypeGetKeychainResponse, componentipc.ComponentMessageGetKeychainResponse{
					Value: cr.container.Keychain.Get(ev.Payload.Key),
				})
			case ev := <-setKeychain:
				if ev.Error {
					return
				}
				cr.container.Keychain.Set(ev.Payload.Key, ev.Payload.Value)
			case ev := <-log:
				if ev.Error {
					return
				}
				cr.container.Logger.Log("Component: '%s'", ev.Payload.Message)
			case <-cr.context.Done():
				return
			}
		}
	}()
}
