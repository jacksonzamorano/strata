package componentipc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"sync"
	"time"
)

type IO struct {
	transport *stdioTransport

	mu sync.Mutex

	threadWaiters map[string]*Thread
	globalWaiters map[MessageType]chan ComponentMessage
}

func NewIO(ctx context.Context, cancel context.CancelFunc, read io.ReadCloser, write io.Writer) *IO {
	cio := &IO{
		transport:     startStdioTransport(ctx, cancel, read, write),
		threadWaiters: map[string]*Thread{},
		globalWaiters: map[MessageType]chan ComponentMessage{},
	}
	go cio.readLoop()
	return cio
}

func makeMessageID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (c *IO) readLoop() {
	for event := range c.transport.read {
		c.mu.Lock()
		threadWaiter := c.threadWaiters[event.Id]
		globalWaiter := c.globalWaiters[event.Type]
		c.mu.Unlock()

		if threadWaiter != nil {
			threadWaiter.incoming <- event
		}
		if globalWaiter != nil {
			globalWaiter <- event
		}
	}
}

func (c *IO) NewThread() *Thread {
	t := NewThread(makeMessageID(), c)
	c.mu.Lock()
	c.threadWaiters[t.id] = t
	c.mu.Unlock()
	return t
}

func (c *IO) newThreadWithID(id string) *Thread {
	t := NewThread(id, c)
	c.mu.Lock()
	c.threadWaiters[t.id] = t
	c.mu.Unlock()
	return t
}

type ReceivedEvent[T any] struct {
	Payload T
	Message ComponentMessage
	Thread  *Thread
	Error   bool
}

func Receive[T any](c *IO, recvType MessageType) chan ReceivedEvent[T] {
	c.mu.Lock()
	cn := make(chan ComponentMessage)
	c.globalWaiters[recvType] = cn
	c.mu.Unlock()

	output := make(chan ReceivedEvent[T])

	go func() {
		for {
			select {
			case ev := <-cn:
				t := c.newThreadWithID(ev.Id)

				var msg T
				if err := json.Unmarshal(ev.Payload, &msg); err != nil {
					continue
				}

				output <- ReceivedEvent[T]{
					Payload: msg,
					Message: ev,
					Thread:  t,
				}
			case <-c.transport.ctx.Done():
				output <- ReceivedEvent[T]{Error: true}
				return
			}
		}
	}()

	return output
}

func ReceiveOnce[T any](c *IO, timeout time.Duration, recvType MessageType) ReceivedEvent[T] {
	child, cancel := context.WithTimeout(c.transport.ctx, timeout)
	defer cancel()

	c.mu.Lock()
	cn := make(chan ComponentMessage)
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
			t := c.newThreadWithID(ev.Id)

			var msg T
			if err := json.Unmarshal(ev.Payload, &msg); err != nil {
				continue
			}

			return ReceivedEvent[T]{
				Payload: msg,
				Message: ev,
				Thread:  t,
			}
		case <-child.Done():
			return ReceivedEvent[T]{Error: true}
		}
	}
}
