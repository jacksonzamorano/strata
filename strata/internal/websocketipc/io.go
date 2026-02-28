package websocketipc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type IO struct {
	transport *websocketTransport

	mu sync.Mutex

	threadWaiters map[string]*Thread
	incoming      chan Message
}

func NewIO(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) *IO {
	io := &IO{
		transport:     startWebsocketTransport(ctx, cancel, conn),
		threadWaiters: map[string]*Thread{},
		incoming:      make(chan Message, 64),
	}
	go io.readLoop()
	return io
}

func makeMessageID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (c *IO) readLoop() {
	for {
		select {
		case <-c.transport.ctx.Done():
			return
		case msg := <-c.transport.read:
			c.mu.Lock()
			threadWaiter := c.threadWaiters[msg.Id]
			c.mu.Unlock()

			if threadWaiter != nil {
				select {
				case <-c.transport.ctx.Done():
					return
				case threadWaiter.incoming <- msg:
				}
			}

			select {
			case <-c.transport.ctx.Done():
				return
			case c.incoming <- msg:
			}
		}
	}
}

func (c *IO) Send(msg Message) bool {
	payload, err := json.Marshal(msg)
	if err != nil {
		return false
	}
	return c.transport.send(payload)
}

func (c *IO) Close() {
	c.transport.cancel()
	_ = c.transport.conn.Close()
}

func (c *IO) Done() <-chan struct{} {
	return c.transport.ctx.Done()
}

func (c *IO) Incoming() <-chan Message {
	return c.incoming
}

func (c *IO) NewThread() *Thread {
	return c.loadOrCreateThread(makeMessageID())
}

func (c *IO) loadOrCreateThread(id string) *Thread {
	c.mu.Lock()
	defer c.mu.Unlock()
	if t, ok := c.threadWaiters[id]; ok {
		return t
	}
	t := NewThread(id, c)
	c.threadWaiters[t.id] = t
	return t
}
