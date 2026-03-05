package strata

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type RequestInfo struct {
	Body    []byte
	HasBody bool
	Headers map[string][]string
	Query   map[string][]string
}

type AppServer struct {
	state           *AppState
	srv             *http.Server
	listener        *http.ServeMux
	approvedActions []core.ApprovedComponentPermission
}

func NewAppServer(tasks []Task, deps []core.ComponentImport, cfg ...*ConfigurationModification) AppServer {
	var approvedActions []core.ApprovedComponentPermission
	for _, op := range cfg {
		if op.Permissions != nil {
			approvedActions = append(approvedActions, op.Permissions...)
		}
	}

	appState := newAppState()
	mux := http.NewServeMux()

	for idx := range tasks {
		url := fmt.Sprintf("/tasks/%s", tasks[idx].Name)
		mux.HandleFunc(url, appState.handler(tasks[idx]))
		appState.host.Emit(hostio.HostMessageTypeTaskRegistered, hostio.EventTaskRegisterPayload{
			Name: tasks[idx].Name,
			Url:  url,
		})
	}

	for idx := range deps {
		cmd, err := deps[idx].Setup()
		if err != nil {
			appState.host.Info("Failed to register component at index %d: '%s'", idx, err.Error())
			continue
		}
		name := cmd.CanonicalName
		cnt := appState.buildContainer(cmd.CanonicalName)

		runner, err := RegisterComponent(cmd, cnt)
		if err != nil {
			appState.host.Info("Failed to register component '%s': '%s'", name, err.Error())
			continue
		}

		ev := componentipc.ReceiveOnce[componentipc.ComponentMessageHello](runner.transport, 5*time.Second, componentipc.ComponentMessageTypeHello)
		if ev.Error {
			appState.host.Emit(hostio.HostMessageTypeComponentRegistered, hostio.HostMessageComponentRegistered{
				Suceeded: false,
				Name:     name,
				Path:     runner.path,
				Error:    new("Component didn't connect."),
			})
			continue
		}
		hello := ev.Payload
		appState.host.Emit(hostio.HostMessageTypeComponentRegistered, hostio.HostMessageComponentRegistered{
			Suceeded: true,
			Name:     hello.Name,
			Version:  hello.Version,
			Path:     runner.path,
		})
		name = hello.Name
		appState.components[name] = runner

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
			appState.components[name].available = true
			continue
		}

		appState.host.Emit(hostio.HostMessageTypeComponentRegistered, hostio.HostMessageComponentRegistered{
			Suceeded: false,
			Name:     name,
			Error:    errMsgPtr,
		})
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	ns := os.Getenv("ADDRESS")

	addr := fmt.Sprintf("%s:%s", ns, port)

	as := AppServer{
		state: &appState,
		srv: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		listener:        mux,
		approvedActions: approvedActions,
	}
	return as
}

func (as *AppServer) Start() error {
	select {
	case <-as.state.host.Done():
		return fmt.Errorf("host rpc connection unavailable")
	default:
	}

	as.state.host.Info("Listening on %s", as.srv.Addr)
	return as.srv.ListenAndServe()
}
