package strata

import (
	"sync"

	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type RuntimeTriggers struct {
	triggers    []ComponentTrigger
	triggerLock sync.RWMutex
}

func (rt *RuntimeTriggers) Add(ns, name string, body func([]byte)) {
	rt.triggerLock.Lock()
	rt.triggers = append(rt.triggers, ComponentTrigger{
		Namespace: ns,
		Name:      name,
		Trigger:   body,
	})
	rt.triggerLock.Unlock()
}

func (rt *RuntimeTriggers) Execute(ns string, r *componentipc.ComponentMessageSendTrigger, as *AppState) {
	rt.triggerLock.RLock()

	for ix := range rt.triggers {
		if rt.triggers[ix].Namespace == ns && rt.triggers[ix].Name == r.Name {
			go rt.triggers[ix].Trigger(r.Payload)
		}
	}

	rt.triggerLock.RUnlock()
}
