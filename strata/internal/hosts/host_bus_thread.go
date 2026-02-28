package hosts

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jacksonzamorano/tasks/strata/core"
)

// Thread is intentionally not goroutine-safe and is expected to have a single owner.
type Thread struct {
	id          string
	coordinator *HostBusCoordinator
	timeout     time.Duration

	lock     sync.Mutex
	incoming chan core.HostMessage
	outgoing map[core.HostMessageType]chan core.HostMessage
}

func NewThread(id string, coordinator *HostBusCoordinator) *Thread {
	t := &Thread{
		id:          id,
		coordinator: coordinator,
		timeout:     2 * time.Minute,
		incoming:    make(chan core.HostMessage),
		outgoing:    map[core.HostMessageType]chan core.HostMessage{},
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

func (t *Thread) loadChannel(recvType core.HostMessageType) chan core.HostMessage {
	t.lock.Lock()
	defer t.lock.Unlock()
	if c, ok := t.outgoing[recvType]; ok {
		return c
	}
	cn := make(chan core.HostMessage)
	t.outgoing[recvType] = cn
	return cn
}

func (t *Thread) dropChannel(recvType core.HostMessageType) {
	t.lock.Lock()
	delete(t.outgoing, recvType)
	t.lock.Unlock()
}

func (t *Thread) Send(typ core.HostMessageType, payload any) bool {
	msg, err := core.NewHostMessage(t.id, typ, payload)
	if err != nil {
		return false
	}
	return t.coordinator.bus.Send(msg)
}

func SendAndReceive[T any](t *Thread, sendType core.HostMessageType, sendPayload any, recvType core.HostMessageType) (T, core.HostMessage) {
	child, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	cn := t.loadChannel(recvType)
	defer t.dropChannel(recvType)

	if !t.Send(sendType, sendPayload) {
		var payload T
		return payload, core.HostMessage{}
	}

	for {
		select {
		case msg := <-cn:
			var payload T
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}
			return payload, msg
		case <-t.coordinator.bus.Done():
			var payload T
			return payload, core.HostMessage{}
		case <-child.Done():
			var payload T
			return payload, core.HostMessage{}
		}
	}
}
