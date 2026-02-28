package hosts

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jacksonzamorano/tasks/strata/core"
)

type WebHost struct {
	persistence core.PersistenceProvider
	channel     *webHostChannel
	enableUI    bool

	once sync.Once

	upgrader websocket.Upgrader
	server   *http.Server

	clientLock sync.RWMutex
	clients    map[*webHostClient]struct{}

	logLock sync.RWMutex
	logs    []core.HostMessageEventReceived
	maxLogs int

	permissionLock     sync.Mutex
	pendingPermissions map[string]chan bool

	messageID atomic.Uint64
}

type webHostClient struct {
	conn       *websocket.Conn
	send       chan any
	done       chan struct{}
	closeOnce  sync.Once
	subscribed atomic.Bool
}

type webHostPermissionRequestPayload struct {
	RequestPermission *core.HostMessageRequestPermission `json:"request_permission"`
}

type webHostPermissionRequestMessage struct {
	Id      string                          `json:"id"`
	Type    core.HostMessageType            `json:"type"`
	Payload webHostPermissionRequestPayload `json:"payload"`
}

func (c *webHostClient) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		_ = c.conn.Close()
	})
}

func NewWebHost(enableUI bool) core.HostBus {
	host := &WebHost{
		enableUI: enableUI,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:            map[*webHostClient]struct{}{},
		maxLogs:            200,
		pendingPermissions: map[string]chan bool{},
	}
	host.channel = &webHostChannel{host: host}
	return host
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

func (wh *WebHost) Channel() core.HostBusChannel {
	return wh.channel
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

	client := &webHostClient{
		conn: conn,
		send: make(chan any, 128),
		done: make(chan struct{}),
	}

	wh.addClient(client)
	defer wh.removeClient(client)

	go wh.writePump(client)
	wh.readPump(client)
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
	client.close()
}

func (wh *WebHost) writePump(client *webHostClient) {
	for {
		select {
		case msg := <-client.send:
			if err := client.conn.WriteJSON(msg); err != nil {
				wh.removeClient(client)
				return
			}
		case <-client.done:
			return
		}
	}
}

func (wh *WebHost) readPump(client *webHostClient) {
	client.conn.SetReadLimit(1 << 20)
	for {
		_, raw, err := client.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg core.HostMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			wh.sendToClient(client, wh.errorMessage("", "bad_message", "Could not decode message."))
			continue
		}

		wh.handleHostMessage(client, msg)
	}
}

func (wh *WebHost) handleHostMessage(client *webHostClient, msg core.HostMessage) {
	switch msg.Type {
	case core.HostMessageTypeHello:
		wh.handleHello(client, msg)
	case core.HostMessageTypeSubscribeLogs:
		wh.handleSubscribeLogs(client, msg)
	case core.HostMessageTypeAuthorizationCreate:
		wh.handleAuthorizationCreate(client, msg)
	case core.HostMessageTypePermissionReplied:
		wh.handlePermissionReplied(client, msg)
	default:
		wh.sendToClient(client, wh.errorMessage(msg.Id, "unsupported_type", "Unsupported message type."))
	}
}

func (wh *WebHost) handleHello(client *webHostClient, msg core.HostMessage) {
	if msg.Payload.Hello == nil {
		wh.sendToClient(client, wh.errorMessage(msg.Id, "invalid_payload", "Missing hello payload."))
		return
	}

	resp := core.HostMessage{
		Id:   msg.Id,
		Type: core.HostMessageTypeHelloAck,
		Payload: core.HostMessagePayload{
			HelloAck: &core.HostMessageHelloAck{
				ProtocolVersion: "1",
				ServerName:      "Strata Web Host",
			},
		},
	}
	wh.sendToClient(client, resp)
	wh.replayAuthorizations(client)
}

func (wh *WebHost) replayAuthorizations(client *webHostClient) {
	authorizations := wh.persistence.Authorization.GetAuthorizations()
	for i := range authorizations {
		auth := authorizations[i]
		wh.sendToClient(client, core.HostMessage{
			Id:   wh.nextMessageID(),
			Type: core.HostMessageTypeAuthorizationCreated,
			Payload: core.HostMessagePayload{
				AuthorizationCreated: &core.HostMessageAuthorizationCreated{
					Nickname:    auth.Nickname,
					Secret:      auth.Secret,
					Source:      auth.Source,
					CreatedDate: auth.CreatedDate,
				},
			},
		})
	}
}

func (wh *WebHost) handleSubscribeLogs(client *webHostClient, msg core.HostMessage) {
	if msg.Payload.SubscribeLogs == nil {
		wh.sendToClient(client, wh.errorMessage(msg.Id, "invalid_payload", "Missing subscribe payload."))
		return
	}

	tail := msg.Payload.SubscribeLogs.Tail
	if tail <= 0 {
		tail = int64(wh.maxLogs)
	}

	client.subscribed.Store(true)
	events := wh.readTailLogs(tail)

	for i := range events {
		event := events[i]
		wh.sendToClient(client, core.HostMessage{
			Id:   wh.nextMessageID(),
			Type: core.HostMessageTypeEventReceived,
			Payload: core.HostMessagePayload{
				EventReceived: &event,
			},
		})
	}

	wh.sendToClient(client, core.HostMessage{
		Id:   msg.Id,
		Type: core.HostMessageTypeSubscribeLogsAck,
		Payload: core.HostMessagePayload{
			SubscribeLogsAck: &core.HostMessageSubscribeLogsAck{
				ReplayCount: int64(len(events)),
			},
		},
	})
}

func (wh *WebHost) handleAuthorizationCreate(client *webHostClient, msg core.HostMessage) {
	if msg.Payload.AuthorizationCreate == nil {
		wh.sendToClient(client, wh.errorMessage(msg.Id, "invalid_payload", "Missing authorization payload."))
		return
	}

	nickname := strings.TrimSpace(msg.Payload.AuthorizationCreate.Nickname)
	if len(nickname) == 0 {
		wh.sendToClient(client, wh.errorMessage(msg.Id, "invalid_nickname", "Nickname is required."))
		return
	}

	auth := wh.persistence.Authorization.NewAuthorization("Web Host", nickname)
	if auth == nil {
		wh.sendToClient(client, wh.errorMessage(msg.Id, "auth_failed", "Could not create token."))
		return
	}

	wh.sendToClient(client, core.HostMessage{
		Id:   msg.Id,
		Type: core.HostMessageTypeAuthorizationCreated,
		Payload: core.HostMessagePayload{
			AuthorizationCreated: &core.HostMessageAuthorizationCreated{
				Nickname:    auth.Nickname,
				Secret:      auth.Secret,
				Source:      auth.Source,
				CreatedDate: auth.CreatedDate,
			},
		},
	})
}

func (wh *WebHost) handlePermissionReplied(client *webHostClient, msg core.HostMessage) {
	response := msg.Payload.PermissionResponse
	if response == nil {
		// Accept request_permission as a fallback to tolerate protocol/schema drift.
		response = msg.Payload.RequestPermission
	}
	if response == nil {
		wh.sendToClient(client, wh.errorMessage(msg.Id, "invalid_payload", "Missing permission response payload."))
		return
	}

	wh.permissionLock.Lock()
	waiter, ok := wh.pendingPermissions[msg.Id]
	if ok {
		delete(wh.pendingPermissions, msg.Id)
	}
	wh.permissionLock.Unlock()

	if !ok {
		wh.sendToClient(client, wh.errorMessage(msg.Id, "unknown_permission_request", "Unknown permission request id."))
		return
	}

	select {
	case waiter <- response.Approve:
	default:
	}
}

func (wh *WebHost) nextMessageID() string {
	return strconv.FormatUint(wh.messageID.Add(1), 10)
}

func (wh *WebHost) errorMessage(id, code, message string) core.HostMessage {
	return core.HostMessage{
		Id:   id,
		Type: core.HostMessageTypeError,
		Payload: core.HostMessagePayload{
			Error: &core.HostMessageError{
				Code:    code,
				Message: message,
			},
		},
	}
}

func (wh *WebHost) sendToClient(client *webHostClient, msg any) bool {
	select {
	case client.send <- msg:
		return true
	default:
		wh.removeClient(client)
		return false
	}
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

func (wh *WebHost) broadcastEvent(ev core.HostMessageEventReceived) {
	wh.writeEvent(ev)

	msg := core.HostMessage{
		Id:   wh.nextMessageID(),
		Type: core.HostMessageTypeEventReceived,
		Payload: core.HostMessagePayload{
			EventReceived: &ev,
		},
	}

	wh.clientLock.RLock()
	clients := make([]*webHostClient, 0, len(wh.clients))
	for client := range wh.clients {
		if client.subscribed.Load() {
			clients = append(clients, client)
		}
	}
	wh.clientLock.RUnlock()

	for _, client := range clients {
		wh.sendToClient(client, msg)
	}
}

func (wh *WebHost) writeEvent(ev core.HostMessageEventReceived) {
	wh.logLock.Lock()
	if len(wh.logs) >= wh.maxLogs {
		wh.logs = wh.logs[1:]
	}
	wh.logs = append(wh.logs, ev)
	wh.logLock.Unlock()
}

func (wh *WebHost) readTailLogs(tail int64) []core.HostMessageEventReceived {
	wh.logLock.RLock()
	defer wh.logLock.RUnlock()

	if len(wh.logs) == 0 {
		return []core.HostMessageEventReceived{}
	}

	count := int(tail)
	if count > len(wh.logs) {
		count = len(wh.logs)
	}
	if count < 0 {
		count = 0
	}

	start := len(wh.logs) - count
	out := make([]core.HostMessageEventReceived, count)
	copy(out, wh.logs[start:])
	return out
}

func (wh *WebHost) emitLog(channel, kind string, namespace *string, message string, payload *string) {
	wh.broadcastEvent(core.HostMessageEventReceived{
		Date:      time.Now(),
		Channel:   channel,
		Kind:      kind,
		Namespace: namespace,
		Message:   message,
		Payload:   payload,
	})
}

type webHostChannel struct {
	host *WebHost
}

func (whc *webHostChannel) Info(v string, args ...any) {
	whc.host.emitLog("log", "info", nil, fmt.Sprintf(v, args...), nil)
}

func (whc *webHostChannel) Event(ev core.EventKind, payload any) {
	encoded, _ := json.Marshal(payload)
	p := string(encoded)
	kind := string(ev)
	msg := fmt.Sprintf("Event(%s)", kind)
	whc.host.emitLog("event", kind, nil, msg, &p)
}

func (whc *webHostChannel) Container(namespace string) core.Logger {
	return &webHostContainerLogger{
		host:      whc.host,
		namespace: namespace,
	}
}

func (whc *webHostChannel) RequestPermission(p core.Permission) bool {
	encoded, _ := json.Marshal(p)
	payload := string(encoded)
	requestID := whc.host.nextMessageID()
	waiter := make(chan bool, 1)

	whc.host.permissionLock.Lock()
	whc.host.pendingPermissions[requestID] = waiter
	whc.host.permissionLock.Unlock()

	clients := whc.host.readConnectedClients()
	if len(clients) == 0 {
		whc.host.permissionLock.Lock()
		delete(whc.host.pendingPermissions, requestID)
		whc.host.permissionLock.Unlock()

		msg := fmt.Sprintf(
			"[%s] requested permission '%s' but no web host clients are connected.",
			p.Container,
			p.Action,
		)
		whc.host.emitLog("permission", "requestRejected", nil, msg, &payload)
		return false
	}

	msg := webHostPermissionRequestMessage{
		Id:   requestID,
		Type: core.HostMessageTypePermissionRequest,
		Payload: webHostPermissionRequestPayload{
			RequestPermission: &core.HostMessageRequestPermission{
				Permission: p,
			},
		},
	}

	delivered := 0
	for _, client := range clients {
		if whc.host.sendToClient(client, msg) {
			delivered += 1
		}
	}
	if delivered == 0 {
		whc.host.permissionLock.Lock()
		delete(whc.host.pendingPermissions, requestID)
		whc.host.permissionLock.Unlock()

		msg := fmt.Sprintf(
			"[%s] requested permission '%s' but request could not be delivered to a web host client.",
			p.Container,
			p.Action,
		)
		whc.host.emitLog("permission", "requestRejected", nil, msg, &payload)
		return false
	}

	startMsg := fmt.Sprintf("[%s] requested permission '%s'. Awaiting approval.", p.Container, p.Action)
	whc.host.emitLog("permission", "requestPending", nil, startMsg, &payload)

	approved := <-waiter
	decision := "denied"
	if approved {
		decision = "approved"
	}
	endMsg := fmt.Sprintf("[%s] permission '%s' %s.", p.Container, p.Action, decision)
	whc.host.emitLog("permission", "requestResolved", nil, endMsg, &payload)
	return approved
}

type webHostContainerLogger struct {
	host      *WebHost
	namespace string
}

func (whc *webHostContainerLogger) Log(v string, args ...any) {
	ns := whc.namespace
	whc.host.emitLog("log", "container", &ns, fmt.Sprintf(v, args...), nil)
}
