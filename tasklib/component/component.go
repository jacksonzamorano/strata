package component

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
)

type Component struct {
	Name      string
	Version   string
	Functions map[string]ComponentFunctionFn
	transport StdioTransport
}

type ComponentMessage struct {
	Type    ComponentMessageType `json:"type"`
	Name    string               `json:"name"`
	Payload json.RawMessage      `json:"payload"`
}

type ComponentResultPayload struct {
	Success  bool
	Response json.RawMessage
	Error    string
}

func Result(r any) *ComponentResultPayload {
	b, _ := json.Marshal(r)
	return &ComponentResultPayload{
		Success:  true,
		Response: b,
	}
}
func Error(e string) *ComponentResultPayload {
	return &ComponentResultPayload{
		Success: true,
		Error:   e,
	}
}

type ComponentResultError struct {
	ErrorMessage string
}

type ComponentFunction struct {
	Name    string
	Execute ComponentFunctionFn
}

type ComponentFunctionFn = func(body []byte, ctx *ComponentContext) *ComponentResultPayload
type ComponentFunctionTypedFn[T any] = func(body T, ctx *ComponentContext) *ComponentResultPayload

type ComponentContext struct {
}

func CreateComponent(name string, version string, fns ...ComponentFunction) *Component {
	cmp := &Component{
		Name:      name,
		Version:   version,
		Functions: map[string]ComponentFunctionFn{},
	}

	for i := range fns {
		cmp.Functions[fns[i].Name] = fns[i].Execute
	}

	return cmp
}

func CreateFunction[T any](name string, fn ComponentFunctionTypedFn[T]) ComponentFunction {
	return ComponentFunction{
		Name: name,
		Execute: func(body []byte, ctx *ComponentContext) *ComponentResultPayload {
			var v T
			err := json.Unmarshal(body, &v)
			if err != nil {
				return Error(err.Error())
			}

			return fn(v, ctx)
		},
	}
}

func (c *Component) DispatchEvent(ev ComponentMessage) {
	switch ev.Type {
	case ComponentMessageTypeExecute:
		if handler, ok := c.Functions[ev.Name]; ok {
			ret := handler([]byte(ev.Payload), &ComponentContext{})
			c.Send(ComponentMessageTypeRet, ret)
		} else {
			c.SendError("Function not found.")
		}
	}
}

func (c *Component) Send(typ ComponentMessageType, payload any) {
	var data []byte

	if s, ok := payload.(string); ok {
		data = []byte(s)
	} else if p, ok := payload.([]byte); ok {
		data = p
	} else {
		var err error
		data, err = json.Marshal(payload)
		if err != nil {
			return
		}
	}

	c.transport.Send(ComponentMessage{Type: typ, Payload: data})
}

func (c *Component) SendError(err string) {
	c.Send(ComponentMessageTypeError, err)
}

func (c *Component) Start() {
	c.transport = StartStdioTransport(os.Stdin, os.Stdout)

	go func() {
		for event := range c.transport.read {
			c.DispatchEvent(event)
		}
	}()

	c.Send(ComponentMessageTypeReady, ComponentReadyMessage{
		Name:    c.Name,
		Version: c.Version,
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
