package websocketipc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type IO struct {
	transport *websocketTransport

	mu sync.Mutex

	threadWaiters map[string]*Thread
	globalWaiters map[MessageType]chan Message
}

func NewIO(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) *IO {
	io := &IO{
		transport:     startWebsocketTransport(ctx, cancel, conn),
		threadWaiters: map[string]*Thread{},
		globalWaiters: map[MessageType]chan Message{},
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
			globalWaiter := c.globalWaiters[msg.Type]
			c.mu.Unlock()

			if threadWaiter != nil {
				threadWaiter.incoming <- msg
			}
			if globalWaiter != nil {
				globalWaiter <- msg
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

type ReceivedEvent struct {
	Message Message
	Thread  *Thread
	Error   bool
}

func Receive(c *IO, recvType MessageType) chan ReceivedEvent {
	c.mu.Lock()
	cn := make(chan Message)
	c.globalWaiters[recvType] = cn
	c.mu.Unlock()

	output := make(chan ReceivedEvent)

	go func() {
		for {
			select {
			case ev := <-cn:
				t := c.loadOrCreateThread(ev.Id)
				output <- ReceivedEvent{Message: ev, Thread: t}
			case <-c.transport.ctx.Done():
				output <- ReceivedEvent{Error: true}
				return
			}
		}
	}()

	return output
}

func ReceiveOnce(c *IO, timeout time.Duration, recvType MessageType) ReceivedEvent {
	child, cancel := context.WithTimeout(c.transport.ctx, timeout)
	defer cancel()

	c.mu.Lock()
	cn := make(chan Message)
	c.globalWaiters[recvType] = cn
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.globalWaiters, recvType)
		c.mu.Unlock()
	}()

	for {
		select {
		case ev := <-cn:
			t := c.loadOrCreateThread(ev.Id)
			return ReceivedEvent{Message: ev, Thread: t}
		case <-child.Done():
			return ReceivedEvent{Error: true}
		}
	}
}
