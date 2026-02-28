package hosts

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jacksonzamorano/tasks/strata/core"
)

func NewConsoleHost() core.HostBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConsoleHost{
		ctx:      ctx,
		cancel:   cancel,
		incoming: make(chan core.HostMessage),
	}
}

type ConsoleHost struct {
	ctx    context.Context
	cancel context.CancelFunc

	incoming chan core.HostMessage
}

func (cl *ConsoleHost) Initialize(_ core.PersistenceProvider) {}

func (cl *ConsoleHost) reply(id string, typ core.HostMessageType, payload any) {
	d, _ := json.Marshal(payload)
	msg := core.HostMessage{
		Id:      id,
		Type:    typ,
		Payload: d,
	}
	cl.incoming <- msg
}

func (cl *ConsoleHost) Send(msg core.HostMessage) bool {
	select {
	case <-cl.ctx.Done():
		return false
	default:
	}

	if msg.Type == core.HostMessageTypePermissionRequest {
		cl.reply(msg.Id, core.HostMessageTypeRespondPermission, core.HostMessageRespondPermission{
			Approve: true,
		})
	}
	payload := "{}"
	if len(msg.Payload) > 0 {
		payload = string(msg.Payload)
	}

	decoded := map[string]any{}
	if jsonErr := json.Unmarshal(msg.Payload, &decoded); jsonErr == nil {
		if encoded, encodeErr := json.Marshal(decoded); encodeErr == nil {
			payload = string(encoded)
		}
	}
	log.Printf("HostMessage(type=%s id=%s payload=%s)", msg.Type, msg.Id, payload)
	return true
}

func (cl *ConsoleHost) Incoming() <-chan core.HostMessage {
	return cl.incoming
}

func (cl *ConsoleHost) Done() <-chan struct{} {
	return cl.ctx.Done()
}
