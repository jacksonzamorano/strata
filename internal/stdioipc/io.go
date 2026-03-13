package stdioipc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"sync"
	"time"
)

type Message[MT ~string] struct {
	Id      string          `json:"id"`
	Type    MT              `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type IO[MT ~string] struct {
	transport *Transport[MT]

	mu sync.Mutex

	threadWaiters map[string]*Thread[MT]
	globalWaiters map[MT]chan Message[MT]
}

func NewIO[MT ~string](ctx context.Context, cancel context.CancelFunc, read io.ReadCloser, write io.Writer) *IO[MT] {
	io := &IO[MT]{
		transport:     NewTransport[MT](ctx, cancel, read, write),
		threadWaiters: map[string]*Thread[MT]{},
		globalWaiters: map[MT]chan Message[MT]{},
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

func (c *IO[MT]) readLoop() {
	for {
		select {
		case <-c.transport.Done():
			return
		case event := <-c.transport.Read():
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
}

func (c *IO[MT]) Done() <-chan struct{} {
	return c.transport.Done()
}

func (c *IO[MT]) Send(typ MT, payload any) bool {
	enc, _ := json.Marshal(payload)
	return c.transport.Send(Message[MT]{
		Id:      makeMessageID(),
		Type:    typ,
		Payload: enc,
	})
}
func (c *IO[MT]) SendId(id string, typ MT, payload any) bool {
	enc, _ := json.Marshal(payload)
	return c.transport.Send(Message[MT]{
		Id:      id,
		Type:    typ,
		Payload: enc,
	})
}

func (c *IO[MT]) NewThread() *Thread[MT] {
	return c.loadOrCreateThread(makeMessageID())
}

func (c *IO[MT]) loadOrCreateThread(id string) *Thread[MT] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if t, ok := c.threadWaiters[id]; ok {
		return t
	}
	t := NewThread(id, c)
	c.threadWaiters[t.id] = t
	return t
}

type ReceivedEvent[MT ~string, T any] struct {
	Payload T
	Message Message[MT]
	Thread  *Thread[MT]
	Error   bool
}

func Receive[T any, MT ~string](c *IO[MT], recvType MT) chan ReceivedEvent[MT, T] {
	c.mu.Lock()
	cn := make(chan Message[MT])
	c.globalWaiters[recvType] = cn
	c.mu.Unlock()

	output := make(chan ReceivedEvent[MT, T])

	go func() {
		for {
			select {
			case msg := <-cn:
				thread := c.loadOrCreateThread(msg.Id)

				var payload T
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					continue
				}

				output <- ReceivedEvent[MT, T]{
					Payload: payload,
					Message: msg,
					Thread:  thread,
				}
			case <-c.transport.Done():
				output <- ReceivedEvent[MT, T]{Error: true}
				return
			}
		}
	}()

	return output
}

func ReceiveOnce[T any, MT ~string](c *IO[MT], timeout time.Duration, recvType MT) ReceivedEvent[MT, T] {
	child, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.mu.Lock()
	cn := make(chan Message[MT])
	c.globalWaiters[recvType] = cn
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.globalWaiters, recvType)
		c.mu.Unlock()
	}()

	for {
		select {
		case msg := <-cn:
			thread := c.loadOrCreateThread(msg.Id)

			var payload T
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}

			return ReceivedEvent[MT, T]{
				Payload: payload,
				Message: msg,
				Thread:  thread,
			}
		case <-c.transport.Done():
			return ReceivedEvent[MT, T]{Error: true}
		case <-child.Done():
			return ReceivedEvent[MT, T]{Error: true}
		}
	}
}

// Thread is intentionally not goroutine-safe and is expected to have a single owner.
type Thread[MT ~string] struct {
	id      string
	io      *IO[MT]

	lock     sync.Mutex
	incoming chan Message[MT]
	outgoing map[MT]chan Message[MT]
}

func NewThread[MT ~string](id string, io *IO[MT]) *Thread[MT] {
	t := &Thread[MT]{
		id:       id,
		io:       io,
		incoming: make(chan Message[MT]),
		outgoing: map[MT]chan Message[MT]{},
	}

	go func() {
		for incoming := range t.incoming {
			t.lock.Lock()
			recv := t.outgoing[incoming.Type]
			t.lock.Unlock()
			if recv != nil {
				recv <- incoming
			}
		}
	}()

	return t
}

func (t *Thread[MT]) ID() string {
	return t.id
}

func (t *Thread[MT]) loadChannel(recvType MT) chan Message[MT] {
	t.lock.Lock()
	defer t.lock.Unlock()
	if c, ok := t.outgoing[recvType]; ok {
		return c
	}
	cn := make(chan Message[MT])
	t.outgoing[recvType] = cn
	return cn
}

func (t *Thread[MT]) dropChannel(recvType MT) {
	t.lock.Lock()
	delete(t.outgoing, recvType)
	t.lock.Unlock()
}

func (t *Thread[MT]) Send(typ MT, payload any) bool {
	return t.io.SendId(t.id, typ, payload)
}

func WaitFor[T any, MT ~string](t *Thread[MT], recvType MT) (T, Message[MT]) {
	cn := t.loadChannel(recvType)
	defer t.dropChannel(recvType)

	for {
		select {
		case msg := <-cn:
			var payload T
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}
			return payload, msg
		case <-t.io.Done():
			var payload T
			return payload, Message[MT]{}
		}
	}
}

func SendAndReceive[T any, MT ~string](t *Thread[MT], sendType MT, sendPayload any, recvType MT) (T, Message[MT]) {
	cn := t.loadChannel(recvType)
	defer t.dropChannel(recvType)

	if !t.Send(sendType, sendPayload) {
		var payload T
		return payload, Message[MT]{}
	}

	for {
		select {
		case msg := <-cn:
			var payload T
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}
			return payload, msg
		case <-t.io.Done():
			var payload T
			return payload, Message[MT]{}
		}
	}
}
