package tasklib

import (
	"encoding/json"
	"os/exec"

	"github.com/jacksonzamorano/tasks/tasklib/component"
)

type ComponentRunner struct {
	transport component.StdioTransport
}

func RegisterComponent(path string) (*ComponentRunner, error) {
	cmd := exec.Command(path)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	transport := component.StartStdioTransport(out, in)

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &ComponentRunner{
		transport,
	}, nil
}

func (cr *ComponentRunner) Send(ev component.ComponentMessageType, nm string, args any) {
	encoded, _ := json.Marshal(args)
	v := component.ComponentMessage{
		Type: ev,
		Name: nm,
		Payload: encoded,
	}
	cr.transport.Send(v)
} 

func (cr *ComponentRunner) Execute(fname string, args any) *component.ComponentResultPayload {
	cr.Send(component.ComponentMessageTypeExecute, fname, args)
	msg := cr.transport.Read()

	var dt *component.ComponentResultPayload
	err := json.Unmarshal(msg.Payload, &dt)
	if err != nil {
		return nil
	}
	return dt
}
