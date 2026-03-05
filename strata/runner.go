package strata

import (
	"context"
	"encoding/json"

	"github.com/jacksonzamorano/strata/component"
	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/componentipc"
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
		componentipc.ComponentMessageTypeExecute,
		componentipc.ComponentMessageExecute{Name: fname, Arguments: enc},
		componentipc.ComponentMessageTypeRet,
	)

	return &payload
}

func (cr *ComponentRunner) HandleAPIRequests() {
	go func() {
		getVal := componentipc.Receive[componentipc.ComponentMessageGetValueRequest](cr.transport, componentipc.ComponentMessageTypeGetValueRequest)
		setVal := componentipc.Receive[componentipc.ComponentMessageSetValueRequest](cr.transport, componentipc.ComponentMessageTypeStoreValueRequest)
		getKeychain := componentipc.Receive[componentipc.ComponentMessageGetKeychainRequest](cr.transport, componentipc.ComponentMessageTypeGetKeychainRequest)
		setKeychain := componentipc.Receive[componentipc.ComponentMessageSetKeychainRequest](cr.transport, componentipc.ComponentMessageTypeStoreKeychainRequest)
		log := componentipc.Receive[componentipc.ComponentMessageLog](cr.transport, componentipc.ComponentMessageTypeLog)
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
