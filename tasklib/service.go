package tasklib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jacksonzamorano/tasks/tasklib/component"
)

type RequestInfo struct {
	Body    []byte
	HasBody bool
	Headers map[string][]string
	Query   map[string][]string
}

type AppServer struct {
	state    *AppState
	srv      *http.Server
	listener *http.ServeMux
}

func NewAppServer(tasks []Task, cNames []string) AppServer {
	appState := newAppState()
	mux := http.NewServeMux()

	for idx := range tasks {
		url := fmt.Sprintf("/tasks/%s", tasks[idx].Name)
		mux.HandleFunc(url, appState.handler(tasks[idx]))
		appState.Logger.Event(EventKindTaskRegistered, EventTaskRegisterPayload{
			Name: tasks[idx].Name,
			Url:  url,
		})
	}

	for idx := range cNames {
		nameIdx := strings.LastIndex(cNames[idx], "/")
		name := cNames[idx][nameIdx+1:]
		cnt := appState.buildContainer(name)
		runner, err := RegisterComponent(cNames[idx], cnt)
		if err != nil {
			appState.Logger.Info("Failed to register component '%s': '%s'", name, err.Error())
			continue
		}
		appState.components[name] = runner
		msg := runner.transport.Read()
		if msg.Type == component.ComponentMessageTypeReady {
			var rdy component.ComponentMessageReady
			_ = json.Unmarshal(msg.Payload, &rdy)
			appState.Logger.Event(EventKindComponentRegistered, EventComponentRegisteredPayload{
				Suceeded: true,
				Name:     rdy.Name,
				Version:  rdy.Version,
			})
		} else {
			appState.Logger.Info("Failed to register component '%s', component sent invalid message.", name)
		}
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
	}
	return as
}
func (as *AppServer) Start() error {
	as.state.Logger.Info("Listening on %s", as.srv.Addr)
	return as.srv.ListenAndServe()
}
