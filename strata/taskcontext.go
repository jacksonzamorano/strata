package strata

import (
	"context"
	"runtime"
	"time"

	"github.com/jacksonzamorano/strata/core"
)

type TaskContext struct {
	Container *Container

	terminal core.Terminal
	context  context.Context
}

func BuildTaskContext(container *Container, ctx context.Context) *TaskContext {
	return &TaskContext{
		Container: container,
		terminal:  &NativeTerminal{},
		context:   ctx,
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
	return c.terminal.Execute(ctx, "", cmd, args...)
}
