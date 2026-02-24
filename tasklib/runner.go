package tasklib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/jacksonzamorano/tasks/tasklib/component"
)

type ComponentRunner struct {
	transport *component.ComponentIO
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

	transport := component.NewComponentIO(ctx, cancel, out, in)

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
		setVal := component.Recieve[component.ComponentMessageSetValueRequest](cr.transport, component.ComponentMessageTypeStoreValueRequest)
		for {
			select {
			case ev := <-getVal:
				ev.Thread.Send(component.ComponentMessageTypeGetValueResponse, component.ComponentMessageGetValueResponse{
					Value: cr.container.Storage.GetString(ev.Payload.Key),
				})
			case ev := <-setVal:
				cr.container.Storage.SetString(ev.Payload.Key, ev.Payload.Value)
			case <-cr.context.Done():
				return
			}
		}
	}()
}
