package component

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type Component struct {
	name       string
	version    string
	functions  map[string]ComponentBindable
	ctx        context.Context
	cancel     context.CancelFunc
	setupFn    ComponentSetupFn
	ioChannel  *componentipc.IO
	storageDir string
}

type ComponentManifest = core.ComponentManifest

func CreateComponent(manifest ComponentManifest, fns ...ComponentBindable) *Component {
	cmp := &Component{
		name:      manifest.Name,
		version:   manifest.Version,
		functions: map[string]ComponentBindable{},
	}

	for i := range fns {
		fn := fns[i]
		fname := fn.getName()
		cmp.functions[fname] = fn
	}

	return cmp
}

type ComponentResultPayload = componentipc.ComponentResultPayload

type ComponentFunctionFn = func(body []byte, ctx *ComponentContainer) *ComponentResultPayload
type ComponentSetupFn = func(ctx *ComponentContainer) string

func (c *Component) Start() {
	c.ctx, c.cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer c.cancel()
	c.ioChannel = componentipc.NewIO(c.ctx, c.cancel, os.Stdin, os.Stdout)

	thread := c.ioChannel.NewThread()
	setup, _ := componentipc.SendAndReceive[componentipc.ComponentMessageSetup](
		thread,
		componentipc.ComponentMessageTypeHello,
		componentipc.ComponentMessageHello{Name: c.name, Version: c.version},
		componentipc.ComponentMessageTypeSetup,
	)
	c.storageDir = setup.StorageDir

	var err string
	if c.setupFn != nil {
		err = c.setupFn(c.buildContext())
	}
	thread.Send(componentipc.ComponentMessageTypeReady, componentipc.ComponentMessageReady{Error: err})

	go func() {
		cn := componentipc.Receive[componentipc.ComponentMessageExecute](c.ioChannel, componentipc.ComponentMessageTypeExecute)
		for {
			ev := <-cn
			if ev.Error {
				return
			}
			d := ev.Payload
			if handler, ok := c.functions[d.Name]; ok {
				ret := handler.Execute([]byte(d.Arguments), c.buildContext())
				ev.Thread.Send(componentipc.ComponentMessageTypeRet, ret)
				continue
			}
			ev.Thread.Send(componentipc.ComponentMessageTypeRet, componentipc.ComponentResultPayload{Error: "Function not found."})
		}
	}()

	select {
	case <-c.ctx.Done():
	case <-c.ioChannel.Done():
	}
}

func (c *Component) StartWithSetup(setup ComponentSetupFn) {
	c.setupFn = setup
	c.Start()
}

type ComponentBindable interface {
	getName() string
	Execute(args []byte, context *ComponentContainer) *ComponentResultPayload
}

type ComponentBindableFn[I any, O any] = func(input *ComponentInput[I, O], ctx *ComponentContainer) *ComponentReturn[O]

type ComponentDefinition[I any, O any] struct {
	componentName string
	functionName  string
}
type ComponentMount[I any, O any] struct {
	Definition *ComponentDefinition[I, O]
	Function   ComponentBindableFn[I, O]
}

func (m *ComponentMount[I, O]) getName() string {
	return m.Definition.functionName
}
func (m *ComponentMount[I, O]) Execute(args []byte, ctx *ComponentContainer) *ComponentResultPayload {
	defer func() {
		if err := recover(); err != nil {
			errorString := ""
			if e, ok := err.(error); ok {
				errorString = e.Error()
			}
			ctx.Logger.Log("Component function crashed with error %s", errorString)
		}
	}()
	var inputB I
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
		return &ComponentResultPayload{
			Success:  true,
			Response: by,
		}
	} else {
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

func Define[I any, O any](manifest core.ComponentManifest, name string) *ComponentDefinition[I, O] {
	return &ComponentDefinition[I, O]{
		componentName: manifest.Name,
		functionName:  name,
	}
}

func Mount[I any, O any](definition *ComponentDefinition[I, O], fn ComponentBindableFn[I, O]) *ComponentMount[I, O] {
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

func (c *ComponentDefinition[I, O]) Execute(mod core.ForeignComponent, input I) (O, bool) {
	var output O
	dec, err := mod.ExecuteFunction(c.componentName, c.functionName, input)
	if err != nil {
		return output, false
	}
	err = json.Unmarshal(dec, &output)
	if err != nil {
		return output, false
	}
	return output, true
}
