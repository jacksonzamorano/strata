package component

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Thread is intentionally not goroutine-safe and is expected to have a single owner.
type Thread struct {
	id      string
	io      *ComponentIO
	timeout time.Duration

	lock     sync.Mutex
	incoming chan ComponentMessage
	outgoing map[ComponentMessageType]chan ComponentMessage
}

func NewThread(id string, io *ComponentIO) *Thread {
	t := &Thread{
		id:       id,
		io:       io,
		timeout:  time.Minute * 2,
		incoming: make(chan ComponentMessage),
		outgoing: map[ComponentMessageType]chan ComponentMessage{},
	}

	go func() {
		for incoming := range t.incoming {
			t.lock.Lock()
			if recv, ok := t.outgoing[incoming.Type]; ok {
				recv <- incoming
			}
			t.lock.Unlock()
		}
	}()

	return t
}

func (t *Thread) ID() string {
	return t.id
}

func (t *Thread) loadChannel(recvType ComponentMessageType) chan ComponentMessage {
	if c, ok := t.outgoing[recvType]; ok {
		return c
	}
	cm := make(chan ComponentMessage)
	t.outgoing[recvType] = cm
	return cm
}

func (t *Thread) Send(typ ComponentMessageType, payload any) {
	data, err := encodePayload(payload)
	if err != nil {
		return
	}
	msg := ComponentMessage{Id: t.id, Type: typ, Payload: data}
	t.io.transport.Send(msg)
}

func SendAndReceive[T any](t *Thread, sendType ComponentMessageType, sendPayload any, recvType ComponentMessageType) (T, ComponentMessage) {
	child, cancel := context.WithTimeout(t.io.transport.ctx, t.timeout)
	defer cancel()
	c := t.loadChannel(recvType)

	t.Send(sendType, sendPayload)

	for {
		select {
		case msg := <-c:

			t.lock.Lock()
			delete(t.outgoing, recvType)
			t.lock.Unlock()

			var payload T
			e := json.Unmarshal(msg.Payload, &payload)
			if e != nil {
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
