package runtimecomponent

import (
	"sync"

	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type Trigger struct {
	Namespace string
	Name      string
	Fn        func([]byte)
}

type Triggers struct {
	triggers    []Trigger
	triggerLock sync.RWMutex
}

func (rt *Triggers) Add(ns, name string, body func([]byte)) {
	rt.triggerLock.Lock()
	rt.triggers = append(rt.triggers, Trigger{
		Namespace: ns,
		Name:      name,
		Fn:        body,
	})
	rt.triggerLock.Unlock()
}

func (rt *Triggers) Execute(ns string, r *componentipc.ComponentMessageSendTrigger) {
	rt.triggerLock.RLock()

	for ix := range rt.triggers {
		if rt.triggers[ix].Namespace == ns && rt.triggers[ix].Name == r.Name {
			go rt.triggers[ix].Fn(r.Payload)
		}
	}

	rt.triggerLock.RUnlock()
}
