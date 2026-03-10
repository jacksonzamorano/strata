package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/jacksonzamorano/strata/hostio"
)

type wsMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

type wsClient struct {
	conn   *websocket.Conn
	cancel context.CancelFunc
}

type WebHost struct {
	validTokens sync.Map
	port        int

	mu     sync.RWMutex
	client *wsClient

	pending sync.Map // map[string]chan json.RawMessage

	stateMu    sync.RWMutex
	logs       []json.RawMessage
	tasks      []hostio.HostMessageTaskRegistered
	components []hostio.HostMessageComponentRegistered
	triggers   []hostio.HostMessageTaskTriggered
	auths      []hostio.HostMessageAuthorizationCreated

	io *hostio.IO
}

func NewWebHost() *WebHost {
	wh := &WebHost{
		port: 7800,
	}
	go wh.serve()
	return wh
}

func (wh *WebHost) serve() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ws":
			wh.handleWS(w, r)
		default:
			wh.handleIndex(w, r)
		}
	})

	fmt.Printf("Web UI: http://localhost:%d\n", wh.port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", wh.port), handler); err != nil {
		fmt.Printf("Web server error: %s\n", err)
	}
}

func (wh *WebHost) checkAuth(r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token == "" {
		_, token, _ = r.BasicAuth()
	}
	if token == "" {
		return false
	}
	_, ok := wh.validTokens.Load(token)
	return ok
}

func (wh *WebHost) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(webUIHTML)
}

func (wh *WebHost) handleWS(w http.ResponseWriter, r *http.Request) {
	if !wh.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := &wsClient{conn: c, cancel: cancel}

	// Kick existing client
	wh.mu.Lock()
	old := wh.client
	wh.client = client
	wh.mu.Unlock()

	if old != nil {
		wh.sendToConn(old, wsMessage{Type: "kicked", Payload: json.RawMessage("{}")})
		old.cancel()
		old.conn.Close(websocket.StatusGoingAway, "new client connected")
	}

	// Send sync
	wh.sendSync(client)

	// Read loop
	wh.readLoop(ctx, client)
}

func (wh *WebHost) readLoop(ctx context.Context, client *wsClient) {
	for {
		_, data, err := client.conn.Read(ctx)
		if err != nil {
			wh.mu.Lock()
			if wh.client == client {
				wh.client = nil
			}
			wh.mu.Unlock()
			client.cancel()
			return
		}

		var msg wsMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "permissionResponse", "secretResponse", "oauthResponse":
			if ch, ok := wh.pending.LoadAndDelete(msg.ID); ok {
				ch.(chan json.RawMessage) <- msg.Payload
			}
		case "createAuthorization":
			var payload struct {
				Nickname string `json:"nickname"`
			}
			if json.Unmarshal(msg.Payload, &payload) == nil && wh.io != nil {
				wh.io.Send(hostio.HostMessageTypeCreateAuthorization, hostio.HostMessageCreateAuthorization{
					Nickname: payload.Nickname,
				})
			}
		case "deleteAuthorization":
			var payload struct {
				Secret string `json:"secret"`
			}
			if json.Unmarshal(msg.Payload, &payload) == nil && wh.io != nil {
				wh.io.Send(hostio.HostMessageTypeDeleteAuthorization, hostio.HostMessageDeleteAuthorization{
					Secret: payload.Secret,
				})
			}
		}
	}
}

func (wh *WebHost) sendSync(client *wsClient) {
	wh.stateMu.RLock()
	defer wh.stateMu.RUnlock()

	type syncPayload struct {
		Logs       []json.RawMessage                        `json:"logs"`
		Tasks      []hostio.HostMessageTaskRegistered       `json:"tasks"`
		Components []hostio.HostMessageComponentRegistered  `json:"components"`
		Triggers   []hostio.HostMessageTaskTriggered        `json:"triggers"`
		Auths      []hostio.HostMessageAuthorizationCreated `json:"authorizations"`
	}

	payload := syncPayload{
		Logs:       wh.logs,
		Tasks:      wh.tasks,
		Components: wh.components,
		Triggers:   wh.triggers,
		Auths:      wh.auths,
	}

	if payload.Logs == nil {
		payload.Logs = []json.RawMessage{}
	}
	if payload.Tasks == nil {
		payload.Tasks = []hostio.HostMessageTaskRegistered{}
	}
	if payload.Components == nil {
		payload.Components = []hostio.HostMessageComponentRegistered{}
	}
	if payload.Triggers == nil {
		payload.Triggers = []hostio.HostMessageTaskTriggered{}
	}
	if payload.Auths == nil {
		payload.Auths = []hostio.HostMessageAuthorizationCreated{}
	}

	wh.sendToConn(client, wsMessage{Type: "sync", Payload: mustMarshal(payload)})
}

func (wh *WebHost) broadcast(msg wsMessage) {
	wh.mu.RLock()
	client := wh.client
	wh.mu.RUnlock()
	if client != nil {
		wh.sendToConn(client, msg)
	}
}

func (wh *WebHost) sendToConn(client *wsClient, msg wsMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.conn.Write(ctx, websocket.MessageText, data)
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func (wh *WebHost) sendRequest(typ string, payload any) (json.RawMessage, bool) {
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	ch := make(chan json.RawMessage, 1)
	wh.pending.Store(id, ch)
	defer wh.pending.Delete(id)

	wh.broadcast(wsMessage{
		Type:    typ,
		ID:      id,
		Payload: mustMarshal(payload),
	})

	select {
	case resp := <-ch:
		return resp, true
	case <-time.After(10 * time.Minute):
		return nil, false
	}
}

// Host interface implementation

func (wh *WebHost) Log(ev hostio.ReceivedEvent[hostio.HostMessageLogEvent]) {
	type logEntry struct {
		Kind      string `json:"kind"`
		Namespace string `json:"namespace"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}

	entry := logEntry{
		Kind:      ev.Payload.Kind,
		Namespace: ev.Payload.Namespace,
		Message:   ev.Payload.Message,
		Timestamp: time.Now().Format(time.RFC3339Nano),
	}

	raw := mustMarshal(entry)

	wh.stateMu.Lock()
	wh.logs = append(wh.logs, raw)
	wh.stateMu.Unlock()

	wh.broadcast(wsMessage{Type: "log", Payload: raw})
}

func (wh *WebHost) TaskRegistered(ev hostio.ReceivedEvent[hostio.HostMessageTaskRegistered]) {
	wh.stateMu.Lock()
	wh.tasks = append(wh.tasks, ev.Payload)
	wh.stateMu.Unlock()

	wh.broadcast(wsMessage{Type: "taskRegistered", Payload: mustMarshal(ev.Payload)})
}

func (wh *WebHost) ComponentRegistered(ev hostio.ReceivedEvent[hostio.HostMessageComponentRegistered]) {
	wh.stateMu.Lock()
	wh.components = append(wh.components, ev.Payload)
	wh.stateMu.Unlock()

	wh.broadcast(wsMessage{Type: "componentRegistered", Payload: mustMarshal(ev.Payload)})
}

func (wh *WebHost) TaskTriggered(ev hostio.ReceivedEvent[hostio.HostMessageTaskTriggered]) {
	wh.stateMu.Lock()
	wh.triggers = append(wh.triggers, ev.Payload)
	wh.stateMu.Unlock()

	wh.broadcast(wsMessage{Type: "taskTriggered", Payload: mustMarshal(ev.Payload)})
}

func (wh *WebHost) AuthorizationsUpdated(ev hostio.ReceivedEvent[hostio.HostMessageAuthorizationsList]) {
	wh.stateMu.Lock()
	wh.auths = ev.Payload.Authorizations
	wh.stateMu.Unlock()

	// Reset valid tokens
	wh.validTokens.Range(func(key, value any) bool {
		wh.validTokens.Delete(key)
		return true
	})
	for _, auth := range ev.Payload.Authorizations {
		wh.validTokens.Store(auth.Secret, struct{}{})
	}

	wh.broadcast(wsMessage{Type: "authorizationsUpdated", Payload: mustMarshal(ev.Payload)})
}

func (wh *WebHost) PermissionRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestPermission]) bool {
	type permReq struct {
		Permission hostio.HostMessageRequestPermission `json:"permission"`
	}

	resp, ok := wh.sendRequest("permissionRequest", permReq{Permission: ev.Payload})
	if !ok {
		return false
	}

	var result struct {
		Approve bool `json:"approve"`
	}
	if json.Unmarshal(resp, &result) != nil {
		return false
	}
	return result.Approve
}

func (wh *WebHost) SecretRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestSecret]) string {
	type secReq struct {
		Namespace string `json:"namespace"`
		Prompt    string `json:"prompt"`
	}

	resp, ok := wh.sendRequest("secretRequest", secReq{
		Namespace: ev.Payload.Namespace,
		Prompt:    ev.Payload.Prompt,
	})
	if !ok {
		return ""
	}

	var result struct {
		Secret string `json:"secret"`
	}
	json.Unmarshal(resp, &result)
	return result.Secret
}

func (wh *WebHost) OauthRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestOauth]) string {
	type oauthReq struct {
		Namespace   string `json:"namespace"`
		Url         string `json:"url"`
		Destination string `json:"destination"`
	}

	resp, ok := wh.sendRequest("oauthRequest", oauthReq{
		Namespace:   ev.Payload.Namespace,
		Url:         ev.Payload.Url,
		Destination: ev.Payload.Destination,
	})
	if !ok {
		return ""
	}

	var result struct {
		Url string `json:"url"`
	}
	json.Unmarshal(resp, &result)
	return result.Url
}
