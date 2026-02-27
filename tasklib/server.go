package tasklib

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jacksonzamorano/tasks/tasklib/core"
	"github.com/jacksonzamorano/tasks/tasklib/internal/componentipc"
)

type RequestInfo struct {
	Body    []byte
	HasBody bool
	Headers map[string][]string
	Query   map[string][]string
}

type AppServer struct {
	bus      core.HostBus
	channel  core.HostBusChannel
	state    *AppState
	srv      *http.Server
	listener *http.ServeMux
}

func NewAppServer(tasks []Task, deps []AppDependancy) AppServer {
	bus := &ConsoleHost{}
	appState := newAppState(bus)
	channel := bus.Channel()
	mux := http.NewServeMux()

	for idx := range tasks {
		url := fmt.Sprintf("/tasks/%s", tasks[idx].Name)
		mux.HandleFunc(url, appState.handler(tasks[idx]))
		channel.Event(core.EventKindTaskRegistered, core.EventTaskRegisterPayload{
			Name: tasks[idx].Name,
			Url:  url,
		})
	}

	for idx := range deps {
		nameIdx := strings.LastIndex(deps[idx].url, "/")
		name := deps[idx].url[nameIdx+1:]
		cnt := appState.buildContainer(name)

		runner, err := RegisterComponent(deps[idx], cnt)
		if err != nil {
			channel.Info("Failed to register component '%s': '%s'", name, err.Error())
			continue
		}

		ev := componentipc.ReceiveOnce[componentipc.ComponentMessageHello](runner.transport, 5*time.Second, componentipc.MessageTypeHello)
		if ev.Error {
			channel.Event(core.EventKindComponentRegistered, core.EventComponentRegisteredPayload{
				Suceeded: false,
				Name:     name,
				Path:     runner.path,
				Error:    new("Component didn't connect."),
			})
			continue
		}
		hello := ev.Payload
		channel.Event(core.EventKindComponentRegistered, core.EventComponentRegisteredPayload{
			Suceeded: true,
			Name:     hello.Name,
			Version:  hello.Version,
			Path:     runner.path,
		})
		name = hello.Name
		appState.components[name] = runner

		rdy, _ := componentipc.SendAndReceive[componentipc.ComponentMessageReady](ev.Thread, componentipc.MessageTypeSetup, struct{}{}, componentipc.MessageTypeReady)
		var err_msg_ptr *string
		if len(rdy.Error) > 0 {
			err_msg_ptr = &rdy.Error
		}

		channel.Event(core.EventKindComponentReady, core.EventComponentReadyPayload{
			Name:      hello.Name,
			Succeeded: err_msg_ptr == nil,
			Error:     err_msg_ptr,
		})
		if err_msg_ptr == nil {
			appState.components[name].available = true
			continue
		}

		channel.Event(core.EventKindComponentRegistered, core.EventComponentRegisteredPayload{
			Suceeded: false,
			Name:     name,
			Error:    new("Component sent invalid message."),
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
		listener: mux,
		bus:      bus,
		channel:  channel,
	}
	return as
}

func (as *AppServer) Start() error {
	as.channel.Info("Listening on %s", as.srv.Addr)
	return as.srv.ListenAndServe()
}
