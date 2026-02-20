package tasklib

import (
	"fmt"
	"net/http"
	"os"
	"time"
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

func NewAppServer(tasks []Task) AppServer {
	appState := newAppState()
	mux := http.NewServeMux()

	for idx := range tasks {
		url := fmt.Sprintf("/tasks/%s", tasks[idx].Name)
		mux.HandleFunc(url, appState.handler(tasks[idx]))
		appState.Logger.Info("Registered task '%s'", url)
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
		listener: mux,
	}
	return as
}
func (as *AppServer) Start() error {
	return as.srv.ListenAndServe()
}
