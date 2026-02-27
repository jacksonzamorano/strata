package core

import (
	"context"
	"github.com/jacksonzamorano/tasks/strata/internal/componentipc"
)

type SandboxProvider interface {
	Execute(ctx context.Context, wd, command string, args ...string) (*componentipc.IO, error)
}
