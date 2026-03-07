package strata

import (
	"encoding/json"

	"github.com/jacksonzamorano/strata/component"
)

type TriggeredTask struct {
	namespace string
	trigger   string
	execute   func([]byte, *TaskContext)
}

type TriggerTaskFn[T any] = func(input T, container *TaskContext)

func NewTriggerTask[T any](trigger component.ComponentTrigger[T], fn TriggerTaskFn[T]) Task {
	return NewTask(fn, &TriggeredTask{
		namespace: trigger.ComponentName,
		trigger:   trigger.TriggerName,
		execute: func(b []byte, ctx *TaskContext) {
			var input T
			json.Unmarshal(b, &input)
			fn(input, ctx)
		},
	})
}

func (tt *TriggeredTask) Attach(ctx *TaskAttachContext) {
	ctx.Trigger(tt.namespace, tt.trigger, func(b []byte) {
		tt.execute(b, ctx.TaskContextGlobal())
	})
}
