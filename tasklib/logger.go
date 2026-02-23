package tasklib

import (
	"encoding/json"
	"fmt"
	"log"
)

type ConsoleLogger struct{}

func (cl *ConsoleLogger) Container(name string) ContainerLogger {
	return &ContainerConsoleLogger{
		namespace: name,
	}
}
func (cl *ConsoleLogger) Event(ev EventKind, payload any) {
	encoded, _ := json.Marshal(payload)
	log.Printf("(%s): %s", ev, string(encoded))
}

func (cl *ConsoleLogger) Info(v string, args ...any) {
	log.Printf(v, args...)
}

type ContainerConsoleLogger struct {
	namespace string
}

func (cl *ContainerConsoleLogger) Info(v string, args ...any) {
	log.Printf("[%s]: %s", cl.namespace, fmt.Sprintf(v, args...))
}
func (cl *ContainerConsoleLogger) Event(ev EventKind, payload any) {
	encoded, _ := json.Marshal(payload)
	log.Printf("[%s] (%s): %s", cl.namespace, ev, string(encoded))
}

type ContainerLogger interface {
	Info(v string, args ...any)
	Event(ev EventKind, payload any)
}

type AppServerLogger interface {
	Info(v string, args ...any)
	Event(ev EventKind, payload any)
	Container(name string) ContainerLogger
}
