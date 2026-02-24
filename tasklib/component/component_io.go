package component

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"sync"
)

var ErrConcurrentInboundWait = errors.New("concurrent inbound wait is not supported")
var ErrConcurrentThreadWait = errors.New("concurrent wait on a single thread is not supported")

type DecodedMessage[T any] struct {
	ID      string
	Type    ComponentMessageType
	Payload T
}

type ComponentIO struct {
	transport StdioTransport

	mu sync.Mutex

	threadWaiters map[string]*Thread
	globalWaiters map[ComponentMessageType]chan ComponentMessage
}

func NewComponentIO(read io.ReadCloser, write io.Writer) *ComponentIO {
	cio := ComponentIO{
		transport:     StartStdioTransport(read, write),
		threadWaiters: map[string]*Thread{},
		globalWaiters: map[ComponentMessageType]chan ComponentMessage{},
	}
	go cio.readLoop()
	return &cio
}

func MakeComponentMessageId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (c *ComponentIO) readLoop() {
	for event := range c.transport.Read {
		c.mu.Lock()
		if wait, ok := c.threadWaiters[event.Id]; ok {
			wait.incoming <- event
		}
		if typ, ok := c.globalWaiters[event.Type]; ok {
			typ <- event
		}
		c.mu.Unlock()
	}
}

func (c *ComponentIO) NewThread() *Thread {
	t := NewThread(MakeComponentMessageId(), c)
	c.mu.Lock()
	c.threadWaiters[t.id] = t
	c.mu.Unlock()
	return t
}

type RecievedEvent[T any] struct {
	Payload T
	Message ComponentMessage
	Thread  *Thread
}

func Recieve[T any](c *ComponentIO, rcvTyp ComponentMessageType) chan RecievedEvent[T] {
	c.mu.Lock()
	cn := make(chan ComponentMessage)
	c.globalWaiters[rcvTyp] = cn
	c.mu.Unlock()

	output := make(chan RecievedEvent[T])

	go func() {
		for ev := range cn {
			c.mu.Lock()

			t := NewThread(ev.Id, c)
			c.threadWaiters[t.id] = t

			c.mu.Unlock()

			var msg T
			err := json.Unmarshal(ev.Payload, &msg)
			if err != nil {
				continue
			}

			output <- RecievedEvent[T]{
				Payload: msg,
				Message: ev,
				Thread:  t,
			}
		}
	}()

	return output
}

func RecieveOnce[T any](c *ComponentIO, rcvTyp ComponentMessageType) RecievedEvent[T] {
	c.mu.Lock()
	cn := make(chan ComponentMessage)
	c.globalWaiters[rcvTyp] = cn
	c.mu.Unlock()

	for ev := range cn {
		c.mu.Lock()

		t := NewThread(ev.Id, c)
		c.threadWaiters[t.id] = t

		c.mu.Unlock()

		var msg T
		err := json.Unmarshal(ev.Payload, &msg)
		if err != nil {
			continue
		}

		return RecievedEvent[T]{
			Payload: msg,
			Message: ev,
			Thread:  t,
		}
	}

	return RecievedEvent[T]{}
}
