package component

import (
	"encoding/json"

	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type ComponentTrigger[T any] struct {
	ComponentName string
	TriggerName   string
}

func NewComponentTrigger[T any](m ComponentManifest, name string) ComponentTrigger[T] {
	return ComponentTrigger[T]{
		ComponentName: name,
		TriggerName:   m.Name,
	}
}

func (c *ComponentTrigger[T]) Send(cc *ComponentContainer, payload T) {
	enc, _ := json.Marshal(payload)
	cc.channel.Send(
		componentipc.ComponentMessageTypeSendTrigger,
		componentipc.ComponentMessageSendTrigger{
			Name:    c.ComponentName,
			Payload: enc,
		},
	)
}
