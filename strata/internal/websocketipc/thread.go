package websocketipc

import (
	"context"
	"sync"
	"time"

	"github.com/jacksonzamorano/tasks/strata/core"
)

// Thread is intentionally not goroutine-safe and is expected to have a single owner.
type Thread struct {
	id      string
	io      *IO
	timeout time.Duration

	lock     sync.Mutex
	incoming chan Message
	outgoing map[MessageType]chan Message
}

func NewThread(id string, io *IO) *Thread {
	t := &Thread{
		id:       id,
		io:       io,
		timeout:  2 * time.Minute,
		incoming: make(chan Message),
		outgoing: map[MessageType]chan Message{},
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

func (t *Thread) ID() string {
	return t.id
}

func (t *Thread) loadChannel(recvType MessageType) chan Message {
	t.lock.Lock()
	defer t.lock.Unlock()
	if c, ok := t.outgoing[recvType]; ok {
		return c
	}
	cm := make(chan Message)
	t.outgoing[recvType] = cm
	return cm
}

func (t *Thread) dropChannel(recvType MessageType) {
	t.lock.Lock()
	delete(t.outgoing, recvType)
	t.lock.Unlock()
}

func (t *Thread) Send(typ MessageType, payload core.HostMessagePayload) bool {
	msg := Message{Id: t.id, Type: typ, Payload: payload}
	return t.io.Send(msg)
}

func SendAndReceive(t *Thread, sendType MessageType, sendPayload core.HostMessagePayload, recvType MessageType) (Message, bool) {
	child, cancel := context.WithTimeout(t.io.transport.ctx, t.timeout)
	defer cancel()
	c := t.loadChannel(recvType)
	defer t.dropChannel(recvType)

	if !t.Send(sendType, sendPayload) {
		return Message{}, false
	}

	for {
		select {
		case msg := <-c:
			return msg, true
		case <-child.Done():
			return Message{}, false
		}
	}
}
