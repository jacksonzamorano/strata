package websocketipc

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/gorilla/websocket"
)

type websocketTransport struct {
	read   chan Message
	write  chan []byte
	ctx    context.Context
	cancel context.CancelFunc
	conn   *websocket.Conn
}

func startWebsocketTransport(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) *websocketTransport {
	wt := &websocketTransport{
		read:   make(chan Message, 64),
		write:  make(chan []byte, 64),
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
	}

	go wt.reader()
	go wt.writer()

	return wt
}

func (wt *websocketTransport) writer() {
	for {
		select {
		case <-wt.ctx.Done():
			_ = wt.conn.Close()
			return
		case payload := <-wt.write:
			if err := wt.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				wt.cancel()
				return
			}
		}
	}
}

func (wt *websocketTransport) reader() {
	for {
		_, payload, err := wt.conn.ReadMessage()
		if errors.Is(err, io.EOF) {
			wt.cancel()
			return
		}
		if err != nil {
			wt.cancel()
			return
		}

		var msg Message
		if err := json.Unmarshal(payload, &msg); err != nil {
			continue
		}

		select {
		case <-wt.ctx.Done():
			return
		case wt.read <- msg:
		}
	}
}

func (wt *websocketTransport) send(payload []byte) bool {
	select {
	case <-wt.ctx.Done():
		return false
	case wt.write <- payload:
		return true
	}
}
