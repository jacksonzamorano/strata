package runtimecomponent

import (
	"context"
	"runtime"
	"time"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/terminal"
)

type terminalProvider struct {
	terminal core.Terminal
}

func newTerminalProvider() terminalProvider {
	return terminalProvider{terminal: &terminal.NativeTerminal{}}
}

func (c *terminalProvider) OpenUrl(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
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

func (c *terminalProvider) Run(maxTime time.Duration, cmd string, args ...string) core.TerminalResult {
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()
	return c.terminal.Execute(ctx, "", cmd, args...)
}

func (c *terminalProvider) RunInDirectory(maxTime time.Duration, wd, cmd string, args ...string) core.TerminalResult {
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()
	return c.terminal.Execute(ctx, wd, cmd, args...)
}
