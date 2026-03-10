package terminal

import (
	"context"
	"os/exec"
	"strings"

	"github.com/jacksonzamorano/strata/core"
)

type NativeTerminal struct{}

func (t *NativeTerminal) Execute(ctx context.Context, wd, cm string, args ...string) core.TerminalResult {
	cmd := exec.CommandContext(ctx, cm, args...)
	if len(wd) > 0 {
		cmd.Dir = wd
	}
	output, err := cmd.CombinedOutput()
	outputS := strings.TrimSpace(string(output))
	if err != nil {
		return core.TerminalResult{
			Error:  err.Error(),
			Code:   cmd.ProcessState.ExitCode(),
			Output: outputS,
			Ok:     false,
		}
	}
	return core.TerminalResult{
		Output: outputS,
		Code:   cmd.ProcessState.ExitCode(),
		Ok:     true,
	}
}
