package component

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
)

type Component struct {
	name      string
	version   string
	functions map[string]ComponentFunctionFn
	setupFn   ComponentSetupFn
	ioChannel *ComponentIO
}

type ComponentMessage struct {
	Id      string               `json:"id"`
	Type    ComponentMessageType `json:"type"`
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
		Success: false,
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
type ComponentSetupFn = func(ctx *ComponentContext) string

type ComponentContext struct {
	Storage *ComponentStorage
}

func CreateComponent(name string, version string, fns ...ComponentFunction) *Component {
	cmp := &Component{
		name:      name,
		version:   version,
		functions: map[string]ComponentFunctionFn{},
	}

	for i := range fns {
		cmp.functions[fns[i].Name] = fns[i].Execute
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

func (c *Component) buildContext() *ComponentContext {
	return &ComponentContext{
		Storage: newComponentStorage(c.ioChannel),
	}
}

func (c *Component) Start() {
	c.ioChannel = NewComponentIO(os.Stdin, os.Stdout)

	thread := c.ioChannel.NewThread()
	_, _ = SendAndReceive[struct{}](thread, ComponentMessageTypeHello, ComponentMessageHello{
		Name:    c.name,
		Version: c.version,
	}, ComponentMessageTypeSetup)

	var err string
	if c.setupFn != nil {
		err = c.setupFn(c.buildContext())
	}
	thread.Send(ComponentMessageTypeReady, ComponentMessageReady{
		Error: err,
	})

	go func() {
		cn := Recieve[ComponentMessageExecute](c.ioChannel, ComponentMessageTypeExecute)
		for ev := range cn {
			d := ev.Payload
			if handler, ok := c.functions[d.Name]; ok {
				ret := handler([]byte(d.Arguments), c.buildContext())
				ev.Thread.Send(ComponentMessageTypeRet, ret)
			} else {
				ev.Thread.Send(ComponentMessageTypeRet, ComponentResultPayload{
					Error: "Function not found.",
				})
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}

func (c *Component) StartWithSetup(setup ComponentSetupFn) {
	c.setupFn = setup
	c.Start()
}

func DecodePayload[T any](msg *ComponentMessage) T {
	var v T
	err := json.Unmarshal(msg.Payload, &v)
	if err != nil {
		panic(err)
	}
	return v
}
