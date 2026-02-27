package component

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacksonzamorano/tasks/strata/core"
	"github.com/jacksonzamorano/tasks/strata/internal/componentipc"
)

type Component struct {
	name      string
	version   string
	functions map[string]ComponentMountable
	ctx       context.Context
	cancel    context.CancelFunc
	setupFn   ComponentSetupFn
	ioChannel *componentipc.IO
}

func CreateComponent(name string, version string, fns ...ComponentMountable) *Component {
	cmp := &Component{
		name:      name,
		version:   version,
		functions: map[string]ComponentMountable{},
	}

	for i := range fns {
		fn := fns[i]
		fname := fn.getName()
		cmp.functions[fname] = fn
	}

	return cmp
}

type ComponentResultPayload = componentipc.ComponentResultPayload

type ComponentFunctionFn = func(body []byte, ctx *ComponentContext) *ComponentResultPayload
type ComponentSetupFn = func(ctx *ComponentContext) string

type ComponentContext struct {
	Storage  core.Storage
	Keychain core.Keychain
	Logger   core.Logger
}

func (c *Component) buildContext() *ComponentContext {
	return &ComponentContext{
		Storage:  newComponentStorage(c.ioChannel),
		Keychain: newComponentKeychain(c.ioChannel),
		Logger:   newComponentLogger(c.ioChannel),
	}
}

func (c *Component) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.ioChannel = componentipc.NewIO(c.ctx, c.cancel, os.Stdin, os.Stdout)

	thread := c.ioChannel.NewThread()
	_, _ = componentipc.SendAndReceive[struct{}](
		thread,
		componentipc.MessageTypeHello,
		componentipc.ComponentMessageHello{Name: c.name, Version: c.version},
		componentipc.MessageTypeSetup,
	)

	var err string
	if c.setupFn != nil {
		err = c.setupFn(c.buildContext())
	}
	thread.Send(componentipc.MessageTypeReady, componentipc.ComponentMessageReady{Error: err})

	go func() {
		cn := componentipc.Receive[componentipc.ComponentMessageExecute](c.ioChannel, componentipc.MessageTypeExecute)
		for {
			ev := <-cn
			if ev.Error {
				return
			}
			d := ev.Payload
			if handler, ok := c.functions[d.Name]; ok {
				ret := handler.Execute([]byte(d.Arguments), c.buildContext())
				ev.Thread.Send(componentipc.MessageTypeRet, ret)
				continue
			}
			ev.Thread.Send(componentipc.MessageTypeRet, componentipc.ComponentResultPayload{Error: "Function not found."})
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

type ComponentMountable interface {
	getName() string
	Execute(args []byte, context *ComponentContext) *ComponentResultPayload
}

type ComponentMountableFn[I any, O any] = func(input *ComponentInput[I, O], ctx *ComponentContext) *ComponentReturn[O]

type ComponentDefinition[I any, O any] struct {
	Name string
}
type ComponentMount[I any, O any] struct {
	Definition *ComponentDefinition[I, O]
	Function   ComponentMountableFn[I, O]
}

func (m *ComponentMount[I, O]) getName() string {
	return m.Definition.Name
}
func (m *ComponentMount[I, O]) Execute(args []byte, ctx *ComponentContext) *ComponentResultPayload {
	var inputB I
	ctx.Logger.Log("Recieve '%s'", string(args))
	err := json.Unmarshal(args, &inputB)
	if err != nil {
		return &ComponentResultPayload{
			Success: false,
			Error:   err.Error(),
		}
	}
	input := &ComponentInput[I, O]{
		Body: inputB,
	}
	res := m.Function(input, ctx)
	if res.Succeeded {
		by, _ := json.Marshal(res.Result)
		ctx.Logger.Log("Send success '%s'", string(by))
		return &ComponentResultPayload{
			Success:  true,
			Response: by,
		}
	} else {
		ctx.Logger.Log("Send error '%s'", res.Error)
		return &ComponentResultPayload{
			Success: false,
			Error:   res.Error,
		}
	}
}

type ComponentInput[I any, O any] struct {
	Body I
}

func (c *ComponentInput[I, O]) Return(ret O) *ComponentReturn[O] {
	return &ComponentReturn[O]{
		Result:    ret,
		Succeeded: true,
	}
}
func (c *ComponentInput[I, O]) Error(msg string) *ComponentReturn[O] {
	return &ComponentReturn[O]{
		Error:     msg,
		Succeeded: false,
	}
}

func Define[I any, O any](name string) *ComponentDefinition[I, O] {
	return &ComponentDefinition[I, O]{
		Name: name,
	}
}

func Mount[I any, O any](definition *ComponentDefinition[I, O], fn ComponentMountableFn[I, O]) *ComponentMount[I, O] {
	return &ComponentMount[I, O]{
		Definition: definition,
		Function:   fn,
	}
}

type ComponentReturn[O any] struct {
	Result    O
	Error     string
	Succeeded bool
}

func (c *ComponentDefinition[I, O]) Execute(module string, mod core.ForeignComponent, input I) (O, bool) {
	var output O
	dec, err := mod.ExecuteFunction(module, c.Name, input)
	if err != nil {
		return output, false
	}
	err = json.Unmarshal(dec, &output)
	if err != nil {
		return output, false
	}
	return output, true
}
