package strata

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/jacksonzamorano/strata/core"
)

type TaskContext struct {
	Container *Container
	Logger    core.Logger

	terminal   core.Terminal
	context    context.Context
	components map[string]*ComponentIO
}

func BuildTaskContext(container *Container, logger core.Logger, cmps map[string]*ComponentIO, ctx context.Context) *TaskContext {
	return &TaskContext{
		Container:  container,
		Logger:     logger,
		components: cmps,
		terminal:   &NativeTerminal{},
		context:    ctx,
	}
}

func (c *TaskContext) OpenUrl(url string) bool {
	ctx, cancel := context.WithTimeout(c.context, time.Second*20)
	defer cancel()
	var command string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
	case "linux":
		command = "xdg-open"
	default:
		return false
	}

	return c.terminal.Execute(ctx, "", command, url).Ok
}

func (c *TaskContext) Run(maxTime time.Duration, cmd string, args ...string) core.TerminalResult {
	ctx, cancel := context.WithTimeout(c.context, maxTime)
	defer cancel()
	return c.terminal.Execute(ctx, "", cmd, args...)
}
func (c *TaskContext) RunInDirectory(maxTime time.Duration, wd, cmd string, args ...string) core.TerminalResult {
	ctx, cancel := context.WithTimeout(c.context, maxTime)
	defer cancel()
	return c.terminal.Execute(ctx, wd, cmd, args...)
}
func (c *TaskContext) ExecuteFunction(cname, fname string, args any) ([]byte, error) {
	if cmp, ok := c.components[cname]; ok {
		res := cmp.Execute(fname, args)
		if res == nil {
			return nil, errors.New("Could not read response.")
		}
		if res.Success {
			return res.Response, nil
		}
		return nil, errors.New(res.Error)
	}
	return nil, errors.New("Module not found.")
}
