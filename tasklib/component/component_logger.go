package component

import "fmt"

type ComponentLogger struct {
	io *ComponentIO
}

func newComponentLogger(io *ComponentIO) *ComponentLogger {
	return &ComponentLogger{
		io,
	}
}

func (cl *ComponentLogger) Log(v string, args ...any) {
	cl.io.NewThread().Send(ComponentMessageTypeLog, ComponentMessageLog{
		Message: fmt.Sprintf(v, args...),
	})
}
