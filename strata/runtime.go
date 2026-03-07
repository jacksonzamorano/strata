package strata

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type Runtime struct {
	state      *AppState
	httpServer *http.Server

	triggers *RuntimeTriggers

	ctx    *context.Context
	cancel context.CancelFunc
}

func NewRuntime(tasks []Task, deps []core.ComponentImport, cfg ...*ConfigurationModification) *Runtime {
	appState := newAppState()
	mux := http.NewServeMux()

	triggers := &RuntimeTriggers{}
	runtimeContext, runtimeCancel := context.WithCancel(context.Background())

	for idx := range deps {
		cmd, err := deps[idx].Setup()
		if err != nil {
			appState.host.Log("Failed to register component at index %d: '%s'", idx, err.Error())
			continue
		}

		runner, err := RegisterComponent(cmd)
		if err != nil {
			appState.host.Log("Failed to register component '%s': '%s'", cmd.CanonicalName, err.Error())
			continue
		}

		ev := componentipc.ReceiveOnce[componentipc.ComponentMessageHello](runner.transport, 10*time.Second, componentipc.ComponentMessageTypeHello)
		if ev.Error {
			appState.host.Emit(hostio.HostMessageTypeComponentRegistered, hostio.HostMessageComponentRegistered{
				Suceeded: false,
				Name:     cmd.CanonicalName,
				Path:     runner.path,
				Error:    new("Component didn't connect."),
			})
			continue
		}

		cnt := appState.buildContainer(ev.Payload.Name)
		runner.Begin(cnt)
		appState.components[ev.Payload.Name] = runner

		rdy, _ := componentipc.SendAndReceive[componentipc.ComponentMessageReady](
			ev.Thread,
			componentipc.ComponentMessageTypeSetup,
			struct{}{},
			componentipc.ComponentMessageTypeReady,
		)
		var errMsgPtr *string
		if len(rdy.Error) > 0 {
			errMsgPtr = &rdy.Error
		}
		if errMsgPtr == nil {
			appState.components[ev.Payload.Name].available = true
			go func() {
				for trigger := range appState.components[ev.Payload.Name].triggers {
					triggers.Execute(ev.Payload.Name, &trigger, appState)
				}
			}()
			appState.host.Emit(hostio.HostMessageTypeComponentRegistered, hostio.HostMessageComponentRegistered{
				Suceeded: true,
				Name:     ev.Payload.Name,
				Version:  ev.Payload.Version,
				Path:     runner.path,
			})
		} else {
			appState.host.Emit(hostio.HostMessageTypeComponentRegistered, hostio.HostMessageComponentRegistered{
				Suceeded: false,
				Name:     ev.Payload.Name,
				Error:    errMsgPtr,
			})
		}
	}

	taskContainer := appState.buildContainer("tasks")
	taskContext := TaskAttachContext{
		mux:          mux,
		authorizaton: appState.persistence.Authorization,
		triggers:     triggers,
		Context:      runtimeContext,
		Container:    taskContainer,
	}
	for idx := range tasks {
		tasks[idx].Implementation.Attach(&taskContext)
		appState.host.Emit(hostio.HostMessageTypeTaskRegistered, hostio.EventTaskRegisterPayload{
			Name: tasks[idx].Name,
		})
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "7700"
	}
	ns := os.Getenv("ADDRESS")

	addr := fmt.Sprintf("%s:%s", ns, port)

	as := Runtime{
		state: appState,
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		ctx:      &runtimeContext,
		cancel:   runtimeCancel,
		triggers: triggers,
	}
	return &as
}

func (as *Runtime) Start() error {
	select {
	case <-as.state.host.Done():
		return fmt.Errorf("host rpc connection unavailable")
	default:
	}

	as.state.host.Log("Listening on %s", as.httpServer.Addr)
	return as.httpServer.ListenAndServe()
}
