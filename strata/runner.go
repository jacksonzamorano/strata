package strata

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/jacksonzamorano/tasks/strata/component"
	"github.com/jacksonzamorano/tasks/strata/core"
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


func RegisterComponent(dep *core.ComponentExecuteCommand, container *Container) (*ComponentRunner, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, dep.Command, dep.Args...)
	if len(dep.WorkingDirectory) > 0 {
		cmd.Dir = dep.WorkingDirectory
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
		path:      dep.CanonicalName,
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
