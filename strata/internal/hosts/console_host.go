package hosts

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jacksonzamorano/tasks/strata/core"
)

func NewConsoleHost() core.HostBus {
	return &ConsoleHost{}
}

type ConsoleHost struct{}

func (cl *ConsoleHost) Initialize(data core.PersistenceProvider) {}
func (cl *ConsoleHost) Channel() core.HostBusChannel {
	return &ConsoleLoggerTransport{}
}

type ConsoleLoggerTransport struct{}

func (cl *ConsoleLoggerTransport) Info(v string, args ...any) {
	log.Printf(v, args...)
}
func (cl *ConsoleLoggerTransport) Container(name string) core.Logger {
	return &ContainerConsoleLogger{
		namespace: name,
	}
}
func (cl *ConsoleLoggerTransport) Event(ev core.EventKind, payload any) {
	encoded, _ := json.Marshal(payload)
	log.Printf("(%s): %s", ev, string(encoded))
}
func (cl *ConsoleLoggerTransport) RequestPermission(p core.Permission) bool {
	log.Printf("[%s] requested permission '%s'. Permissions are rejected in the console host.", p.Container, p.Action)
	return false
}

type ContainerConsoleLogger struct {
	namespace string
}

func (cl *ContainerConsoleLogger) Log(v string, args ...any) {
	log.Printf("[%s]: %s", cl.namespace, fmt.Sprintf(v, args...))
}
