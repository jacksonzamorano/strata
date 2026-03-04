package component

import (
	"encoding/json"
	"fmt"

	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type ComponentLogger struct {
	io *componentipc.IO
}

func newComponentLogger(io *componentipc.IO) *ComponentLogger {
	return &ComponentLogger{io: io}
}

func (cl *ComponentLogger) Log(v string, args ...any) {
	cl.io.NewThread().Send(componentipc.MessageTypeLog, componentipc.ComponentMessageLog{
		Message: fmt.Sprintf(v, args...),
	})
}

func (cl *ComponentLogger) Event(ev string, payload any) {
	encoded, _ := json.Marshal(payload)
	cl.Log("(%s): %s", ev, string(encoded))
}
