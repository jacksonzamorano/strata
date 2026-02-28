package hosts

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/jacksonzamorano/tasks/strata/core"
)

type HostBusCoordinator struct {
	bus core.HostBus

	mu sync.Mutex

	threadWaiters map[string]*Thread
	globalWaiters map[core.HostMessageType]chan core.HostMessage
}

func NewHostBusCoordinator(bus core.HostBus) *HostBusCoordinator {
	c := &HostBusCoordinator{
		bus:           bus,
		threadWaiters: map[string]*Thread{},
		globalWaiters: map[core.HostMessageType]chan core.HostMessage{},
	}
	go c.readLoop()
	return c
}

func makeHostMessageID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (c *HostBusCoordinator) readLoop() {
	incoming := c.bus.Incoming()
	for {
		select {
		case <-c.bus.Done():
			return
		case ev := <-incoming:
			msg := ev

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

func (c *HostBusCoordinator) NewThread() *Thread {
	return c.loadOrCreateThread(makeHostMessageID())
}

func (c *HostBusCoordinator) Send(msg core.HostMessage) bool {
	return c.bus.Send(msg)
}

func (c *HostBusCoordinator) loadOrCreateThread(id string) *Thread {
	c.mu.Lock()
	defer c.mu.Unlock()
	if t, ok := c.threadWaiters[id]; ok {
		return t
	}
	t := NewThread(id, c)
	c.threadWaiters[t.id] = t
	return t
}

type ReceivedEvent[T any] struct {
	Payload T
	Message core.HostMessage
	Thread  *Thread
	Error   bool
}

func Receive[T any](c *HostBusCoordinator, recvType core.HostMessageType) chan ReceivedEvent[T] {
	c.mu.Lock()
	cn := make(chan core.HostMessage)
	c.globalWaiters[recvType] = cn
	c.mu.Unlock()

	output := make(chan ReceivedEvent[T])

	go func() {
		for {
			select {
			case msg := <-cn:
				thread := c.loadOrCreateThread(msg.Id)

				var payload T
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					continue
				}

				output <- ReceivedEvent[T]{
					Payload: payload,
					Message: msg,
					Thread:  thread,
				}
			case <-c.bus.Done():
				output <- ReceivedEvent[T]{Error: true}
				return
			}
		}
	}()

	return output
}

func ReceiveOnce[T any](c *HostBusCoordinator, timeout time.Duration, recvType core.HostMessageType) ReceivedEvent[T] {
	child, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.mu.Lock()
	cn := make(chan core.HostMessage)
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

			return ReceivedEvent[T]{
				Payload: payload,
				Message: msg,
				Thread:  thread,
			}
		case <-c.bus.Done():
			return ReceivedEvent[T]{Error: true}
		case <-child.Done():
			return ReceivedEvent[T]{Error: true}
		}
	}
}
