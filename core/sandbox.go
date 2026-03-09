package core

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed nofs.sb
var seatbelt []byte

type SandboxProvider interface {
	Execute(ctx context.Context, desc *ComponentExecuteCommand) (*exec.Cmd, error)
}

func PlatformSandboxProvider() SandboxProvider {
	switch runtime.GOOS {
	case "darwin":
		return &SeatbeltSandboxProvider{}
	}
	return &SandboxPrivilegedProvider{}
}

type SandboxPrivilegedProvider struct{}

func (s *SandboxPrivilegedProvider) Execute(ctx context.Context, desc *ComponentExecuteCommand) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, desc.Command, desc.Args...)
	if len(desc.WorkingDirectory) > 0 {
		cmd.Dir = desc.WorkingDirectory
	}
	return cmd, nil
}

type SeatbeltSandboxProvider struct{}

func resolveCommandPath(desc *ComponentExecuteCommand) (string, error) {
	// Is absolute path?
	if path.IsAbs(desc.Command) {
		return path.Clean(desc.Command), nil
	}
	// Is path at all?
	if strings.ContainsRune(desc.Command, '/') {
		if len(desc.WorkingDirectory) > 0 {
			return path.Clean(path.Join(desc.WorkingDirectory, desc.Command)), nil
		}
		abs, err := filepath.Abs(desc.Command)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	// Resolve binary via PATH
	return exec.LookPath(desc.Command)
}

func (s *SeatbeltSandboxProvider) Execute(ctx context.Context, desc *ComponentExecuteCommand) (*exec.Cmd, error) {
	tmp, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	commandPath, err := resolveCommandPath(desc)
	if err != nil {
		return nil, err
	}

	sbContents := append([]byte(nil), seatbelt...)
	sbContents = fmt.Appendf(
		sbContents,
		"(allow process-exec (literal %q))\n",
		commandPath,
	)
	if len(desc.WorkingDirectory) > 0 {
		sbContents = fmt.Appendf(
			sbContents,
			"(allow file-read* (subpath %q))\n",
			desc.WorkingDirectory,
		)
	}
	sbContents = fmt.Appendf(sbContents,
		"(allow file-read* (subpath %q))\n", desc.StorageDir)
	sbContents = fmt.Appendf(sbContents,
		"(allow file-write* (subpath %q))\n", desc.StorageDir)

	sb := path.Join(tmp, "com.strata.cache", "strata-sandbox-def.sb")
	if err := os.MkdirAll(path.Dir(sb), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(sb, sbContents, 0644); err != nil {
		return nil, err
	}

	args := []string{
		"-f",
		sb,
		"--",
		desc.Command,
	}
	args = append(args, desc.Args...)

	cmd := exec.CommandContext(ctx, "sandbox-exec", args...)
	if len(desc.WorkingDirectory) > 0 {
		cmd.Dir = desc.WorkingDirectory
	}
	return cmd, nil
}
