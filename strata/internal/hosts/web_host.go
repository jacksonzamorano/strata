package hosts

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jacksonzamorano/tasks/strata/core"
	"github.com/jacksonzamorano/tasks/strata/internal/websocketipc"
)

type WebHost struct {
	persistence core.PersistenceProvider
	enableUI    bool

	ctx    context.Context
	cancel context.CancelFunc

	incoming chan core.HostMessage
	once     sync.Once

	upgrader websocket.Upgrader
	server   *http.Server

	clientLock sync.RWMutex
	clients    map[*webHostClient]struct{}
}

type webHostClient struct {
	io        *websocketipc.IO
	cancel    context.CancelFunc
	closeOnce sync.Once
}

func NewWebHost(enableUI bool) core.HostBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebHost{
		enableUI: enableUI,
		ctx:      ctx,
		cancel:   cancel,
		incoming: make(chan core.HostMessage, 256),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: map[*webHostClient]struct{}{},
	}
}

func (wh *WebHost) Initialize(data core.PersistenceProvider) {
	wh.persistence = data
	wh.once.Do(func() {
		mux := http.NewServeMux()
		mux.Handle("/", wh.basicAuth(http.HandlerFunc(wh.handleIndex)))
		mux.Handle("/ws", wh.basicAuth(http.HandlerFunc(wh.handleWebsocket)))

		wh.server = &http.Server{
			Addr:              "127.0.0.1:9090",
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}

		go func() {
			err := wh.server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Printf("WebHost failed: %s", err.Error())
			}
		}()
	})
}

func (wh *WebHost) Send(msg core.HostMessage) bool {
	select {
	case <-wh.ctx.Done():
		return false
	default:
	}

	clients := wh.readConnectedClients()
	for _, client := range clients {
		if client.io.Send(msg) {
			continue
		}
		wh.removeClient(client)
	}
	return true
}

func (wh *WebHost) Incoming() <-chan core.HostMessage {
	return wh.incoming
}

func (wh *WebHost) Done() <-chan struct{} {
	return wh.ctx.Done()
}

func (wh *WebHost) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !wh.enableUI {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Strata web UI disabled."))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(webHostIndexHTML)
}

func (wh *WebHost) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, password, ok := r.BasicAuth()
		if !ok || len(password) == 0 {
			wh.authChallenge(w)
			return
		}
		auth := wh.persistence.Authorization.GetAuthorization(password)
		if auth == nil || !auth.Active {
			wh.authChallenge(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (wh *WebHost) authChallenge(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Strata Host"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (wh *WebHost) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(1 << 20)

	ctx, cancel := context.WithCancel(context.Background())
	io := websocketipc.NewIO(ctx, cancel, conn)
	client := &webHostClient{
		io:     io,
		cancel: cancel,
	}

	wh.addClient(client)
	defer wh.removeClient(client)

	wh.handleClientMessages(client)
}

func (wh *WebHost) handleClientMessages(client *webHostClient) {
	incoming := client.io.Incoming()

	for {
		select {
		case msg := <-incoming:
			wh.pushIncoming(msg)
		case <-client.io.Done():
			return
		case <-wh.ctx.Done():
			return
		}
	}
}

func (wh *WebHost) pushIncoming(msg core.HostMessage) bool {
	select {
	case <-wh.ctx.Done():
		return false
	case wh.incoming <- msg:
		return true
	}
}

func (wh *WebHost) addClient(client *webHostClient) {
	wh.clientLock.Lock()
	wh.clients[client] = struct{}{}
	wh.clientLock.Unlock()
}

func (wh *WebHost) removeClient(client *webHostClient) {
	wh.clientLock.Lock()
	delete(wh.clients, client)
	wh.clientLock.Unlock()

	client.closeOnce.Do(func() {
		client.cancel()
		client.io.Close()
	})
}

func (wh *WebHost) readConnectedClients() []*webHostClient {
	wh.clientLock.RLock()
	defer wh.clientLock.RUnlock()

	clients := make([]*webHostClient, 0, len(wh.clients))
	for client := range wh.clients {
		clients = append(clients, client)
	}
	return clients
}
