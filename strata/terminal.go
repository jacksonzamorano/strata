package strata

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	"github.com/jacksonzamorano/strata/core"
)

type TerminalProvider struct {
	terminal core.Terminal
}

func (c *TerminalProvider) OpenUrl(url string) bool {
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

func (c *TerminalProvider) Run(maxTime time.Duration, cmd string, args ...string) core.TerminalResult {
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()
	return c.terminal.Execute(ctx, "", cmd, args...)
}
func (c *TerminalProvider) RunInDirectory(maxTime time.Duration, wd, cmd string, args ...string) core.TerminalResult {
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()
	return c.terminal.Execute(ctx, "", cmd, args...)
}

type NativeTerminal struct{}

func (t *NativeTerminal) Execute(ctx context.Context, wd, cm string, args ...string) core.TerminalResult {
	cmd := exec.CommandContext(ctx, cm, args...)
	if len(wd) > 0 {
		cmd.Dir = wd
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return core.TerminalResult{
			Error:  err.Error(),
			Output: string(output),
			Ok:     false,
		}
	}
	return core.TerminalResult{
		Output: string(output),
		Ok:     false,
	}
}
