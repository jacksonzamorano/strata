package strata

import (
	"encoding/json"

	"github.com/jacksonzamorano/strata/component"
	"github.com/jacksonzamorano/strata/core"
)

type TriggeredTask struct {
	namespace string
	trigger   string
	execute   func([]byte, *Container)
}

type TriggerTaskFn[T any] = func(input T, container *Container)

func NewTriggerTask[T any](m core.ComponentManifest, trigger component.ComponentTrigger[T], fn TriggerTaskFn[T]) Task {
	return NewTask(fn, &TriggeredTask{
		namespace: m.Name,
		trigger:   trigger.Name,
		execute: func(b []byte, container *Container) {
			var input T
			json.Unmarshal(b, &input)
			fn(input, container)
		},
	})
}

func (tt *TriggeredTask) Attach(ctx *TaskAttachContext) {
	ctx.Trigger(tt.namespace, tt.trigger, func(b []byte) {
		tt.execute(b, ctx.Container)
	})
}
