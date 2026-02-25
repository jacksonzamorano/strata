package componentipc

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Thread is intentionally not goroutine-safe and is expected to have a single owner.
type Thread struct {
	id      string
	io      *IO
	timeout time.Duration

	lock     sync.Mutex
	incoming chan ComponentMessage
	outgoing map[MessageType]chan ComponentMessage
}

func NewThread(id string, io *IO) *Thread {
	t := &Thread{
		id:       id,
		io:       io,
		timeout:  2 * time.Minute,
		incoming: make(chan ComponentMessage),
		outgoing: map[MessageType]chan ComponentMessage{},
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

func (t *Thread) loadChannel(recvType MessageType) chan ComponentMessage {
	t.lock.Lock()
	defer t.lock.Unlock()
	if c, ok := t.outgoing[recvType]; ok {
		return c
	}
	cm := make(chan ComponentMessage)
	t.outgoing[recvType] = cm
	return cm
}

func (t *Thread) dropChannel(recvType MessageType) {
	t.lock.Lock()
	delete(t.outgoing, recvType)
	t.lock.Unlock()
}

func (t *Thread) Send(typ MessageType, payload any) {
	data, err := encodePayload(payload)
	if err != nil {
		return
	}
	msg := ComponentMessage{Id: t.id, Type: typ, Payload: data}
	t.io.transport.send(msg)
}

func SendAndReceive[T any](t *Thread, sendType MessageType, sendPayload any, recvType MessageType) (T, ComponentMessage) {
	child, cancel := context.WithTimeout(t.io.transport.ctx, t.timeout)
	defer cancel()
	c := t.loadChannel(recvType)
	defer t.dropChannel(recvType)

	t.Send(sendType, sendPayload)

	for {
		select {
		case msg := <-c:
			var payload T
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}
			return payload, msg
		case <-child.Done():
			var payload T
			return payload, ComponentMessage{}
		}
	}
}

func encodePayload(payload any) ([]byte, error) {
	if s, ok := payload.(string); ok {
		return []byte(s), nil
	}
	if p, ok := payload.([]byte); ok {
		return p, nil
	}
	return json.Marshal(payload)
}
