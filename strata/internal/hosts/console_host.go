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
		incoming: make(chan core.HostReceivedEvent),
	}
}

type ConsoleHost struct {
	ctx    context.Context
	cancel context.CancelFunc

	incoming chan core.HostReceivedEvent
}

func (cl *ConsoleHost) Initialize(_ core.PersistenceProvider) {}

func (cl *ConsoleHost) Send(msg core.HostMessage) bool {
	select {
	case <-cl.ctx.Done():
		return false
	default:
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

func (cl *ConsoleHost) Incoming() <-chan core.HostReceivedEvent {
	return cl.incoming
}

func (cl *ConsoleHost) Done() <-chan struct{} {
	return cl.ctx.Done()
}
