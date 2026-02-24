package tasklib

import (
	"encoding/json"
	"os/exec"

	"github.com/jacksonzamorano/tasks/tasklib/component"
)

type ComponentRunner struct {
	transport *component.ComponentIO
	container *Container
	available bool
}

func RegisterComponent(path string, container *Container) (*ComponentRunner, error) {
	cmd := exec.Command(path)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	transport := component.NewComponentIO(out, in)

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	runner := &ComponentRunner{
		transport: transport,
		container: container,
		available: false,
	}

	runner.ListenForStorage()

	return runner, nil
}

func (cr *ComponentRunner) Send(ev component.ComponentMessageType, nm string, args any) {
	cr.transport.NewThread().Send(ev, args)
}

func (cr *ComponentRunner) Execute(fname string, args any) *component.ComponentResultPayload {
	thread := cr.transport.NewThread()
	enc, _ := json.Marshal(args)
	payload, _ := component.SendAndReceive[component.ComponentResultPayload](thread, component.ComponentMessageTypeExecute, component.ComponentMessageExecute{
		Name:      fname,
		Arguments: enc,
	}, component.ComponentMessageTypeRet)

	return &payload
}

func (cr *ComponentRunner) ListenForStorage() {
	go func() {
		getVal := component.Recieve[component.ComponentMessageGetValueRequest](cr.transport, component.ComponentMessageTypeGetValueRequest)
		setVal := component.Recieve[component.ComponentMessageSetValueRequest](cr.transport, component.ComponentMessageTypeGetValueRequest)
		for {
			select {
			case ev := <-getVal:
				ev.Thread.Send(component.ComponentMessageTypeGetValueResponse, component.ComponentMessageGetValueResponse{
					Value: cr.container.Storage.GetString(ev.Payload.Key),
				})
			case ev := <-setVal:
				cr.container.Storage.SetString(ev.Payload.Key, ev.Payload.Value)
			}
		}
	}()
}
