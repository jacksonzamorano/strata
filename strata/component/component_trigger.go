package component

import (
	"encoding/json"

	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type ComponentTrigger[T any] struct {
	Name string
}

func NewComponentTrigger[T any](name string) ComponentTrigger[T] {
	return ComponentTrigger[T]{
		Name: name,
	}
}

func (c *ComponentTrigger[T]) Send(cc *ComponentContainer, payload T) {
	enc, _ := json.Marshal(payload)
	cc.channel.Send(
		componentipc.ComponentMessageTypeSendTrigger,
		componentipc.ComponentMessageSendTrigger{
			Name:    c.Name,
			Payload: enc,
		},
	)
}
