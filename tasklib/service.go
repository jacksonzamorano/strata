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
	state    *AppState
	srv      *http.Server
	listener *http.ServeMux
}

type AppDependancyType int

const (
	AppDependancyTypeBinary AppDependancyType = iota
	AppDependancyTypeLocalProject
	AppDependancyTypeGit
)

type AppDependancy struct {
	url      string
	branch   string
	subdir   string
	dep_type AppDependancyType
}

func Binary(url string) AppDependancy {
	return AppDependancy{
		url:      url,
		dep_type: AppDependancyTypeBinary,
	}
}
func LocalProject(url string) AppDependancy {
	return AppDependancy{
		url:      url,
		dep_type: AppDependancyTypeLocalProject,
	}
}
func Git(url string, branch string) AppDependancy {
	return AppDependancy{
		url:      url,
		branch:   branch,
		dep_type: AppDependancyTypeGit,
	}
}
func GitSubdirectory(url string, branch string, subdir string) AppDependancy {
	return AppDependancy{
		url:      url,
		branch:   branch,
		subdir:   subdir,
		dep_type: AppDependancyTypeGit,
	}
}
func Import(deps ...AppDependancy) []AppDependancy {
	return deps
}

func NewAppServer(tasks []Task, deps []AppDependancy) AppServer {
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

	for idx := range deps {
		nameIdx := strings.LastIndex(deps[idx].url, "/")
		name := deps[idx].url[nameIdx+1:]
		cnt := appState.buildContainer(name)

		runner, err := RegisterComponent(deps[idx], cnt)
		if err != nil {
			appState.Logger.Info("Failed to register component '%s': '%s'", name, err.Error())
			continue
		}

		ev := component.RecieveOnce[component.ComponentMessageHello](runner.transport, 5*time.Second, component.ComponentMessageTypeHello)
		if ev.Error == true {
			appState.Logger.Event(EventKindComponentRegistered, EventComponentRegisteredPayload{
				Suceeded: false,
				Name:     name,
				Path:     runner.path,
				Error:    new("Component didn't connect."),
			})
			continue
		}
		hello := ev.Payload
		appState.Logger.Event(EventKindComponentRegistered, EventComponentRegisteredPayload{
			Suceeded: true,
			Name:     hello.Name,
			Version:  hello.Version,
			Path:     runner.path,
		})
		name = hello.Name
		appState.components[name] = runner

		rdy, _ := component.SendAndReceive[component.ComponentMessageReady](ev.Thread, component.ComponentMessageTypeSetup, struct{}{}, component.ComponentMessageTypeReady)
		var err_msg_ptr *string = nil
		if len(rdy.Error) > 0 {
			err_msg_ptr = &rdy.Error
		}

		appState.Logger.Event(EventKindComponentReady, EventComponentReadyPayload{
			Name:      hello.Name,
			Succeeded: err_msg_ptr == nil,
			Error:     err_msg_ptr,
		})
		if err_msg_ptr == nil {
			appState.components[name].available = true
			continue
		}

		appState.Logger.Event(EventKindComponentRegistered, EventComponentRegisteredPayload{
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
	}
	return as
}
func (as *AppServer) Start() error {
	as.state.Logger.Info("Listening on %s", as.srv.Addr)
	return as.srv.ListenAndServe()
}
