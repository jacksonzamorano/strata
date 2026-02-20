package tasklib

import (
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
	state      *AppState
	srv        *http.Server
	listener   *http.ServeMux
}

func NewAppServer(tasks []Task, cNames []string) AppServer {
	appState := newAppState()
	mux := http.NewServeMux()

	for idx := range tasks {
		url := fmt.Sprintf("/tasks/%s", tasks[idx].Name)
		mux.HandleFunc(url, appState.handler(tasks[idx]))
		appState.Logger.Info("Registered task '%s'", url)
	}

	for idx := range cNames {
		nameIdx := strings.LastIndex(cNames[idx], "/")
		name := cNames[idx][nameIdx+1:]
		runner, err := RegisterComponent(cNames[idx])
		if err != nil {
			appState.Logger.Info("Failed to register component '%s': '%s'", name, err.Error())
			continue
		}
		appState.components[name] = runner
		msg := runner.transport.Read()
		if msg.Type == component.ComponentMessageTypeReady {
			appState.Logger.Info("Registed '%s': '%s'", name, msg.Payload)
		} else {
			appState.Logger.Info("Failed to register component '%s', component sent invalid message.", name)
		}
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	addr := os.Getenv("ADDRESS")

	as := AppServer{
		state: &appState,
		srv: &http.Server{
			Addr:              fmt.Sprintf("%s:%s", addr, port),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		listener:   mux,
	}
	return as
}
func (as *AppServer) Start() error {
	return as.srv.ListenAndServe()
}
