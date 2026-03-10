package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type ProjectKind string

const (
	ProjectKindApp       ProjectKind = "app"
	ProjectKindComponent ProjectKind = "component"
)

type GenerateProjectOptions struct {
	Kind       ProjectKind
	Directory  string
	ModulePath string
}

type GenerateProjectResult struct {
	Directory  string
	Kind       ProjectKind
	ModulePath string
}

type BootstrapError struct {
	Command string
	Output  string
	Err     error
}

func (e *BootstrapError) Error() string {
	return fmt.Sprintf("%s failed: %v", e.Command, e.Err)
}

func (e *BootstrapError) Unwrap() error {
	return e.Err
}

func (e *BootstrapError) Detail() string {
	if len(strings.TrimSpace(e.Output)) == 0 {
		return e.Error()
	}
	return strings.TrimSpace(e.Output)
}

var goCommandRunner = func(ctx context.Context, directory string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = directory
	return cmd.CombinedOutput()
}

//go:embed templates/* templates/**/*
var embeddedTemplates embed.FS

const (
	appTemplateModule             = "example.com/strata-app"
	componentTemplateModule       = "example.com/strata-component"
	componentTemplateDefinitions  = "example.com/strata-component/definitions"
	componentTemplateManifestName = "component-template"
)

func GenerateProject(ctx context.Context, opts GenerateProjectOptions) (*GenerateProjectResult, error) {
	if opts.Kind != ProjectKindApp && opts.Kind != ProjectKindComponent {
		return nil, errors.New("unknown project kind")
	}
	if len(strings.TrimSpace(opts.Directory)) == 0 {
		return nil, errors.New("target directory is required")
	}

	targetDir := filepath.Clean(opts.Directory)
	modulePath := strings.TrimSpace(opts.ModulePath)
	if len(modulePath) == 0 {
		modulePath = defaultModulePath(targetDir)
	}
	if len(modulePath) == 0 {
		return nil, errors.New("could not determine module path")
	}

	if err := prepareTargetDirectory(targetDir); err != nil {
		return nil, err
	}
	if err := writeTemplate(opts.Kind, targetDir, modulePath); err != nil {
		return nil, err
	}

	result := &GenerateProjectResult{
		Directory:  targetDir,
		Kind:       opts.Kind,
		ModulePath: modulePath,
	}

	if output, err := goCommandRunner(ctx, targetDir, "mod", "tidy"); err != nil {
		return result, &BootstrapError{
			Command: "go mod tidy",
			Output:  string(output),
			Err:     err,
		}
	}

	return result, nil
}

func defaultModulePath(targetDir string) string {
	base := filepath.Base(targetDir)
	if base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

func prepareTargetDirectory(targetDir string) error {
	info, err := os.Stat(targetDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s already exists and is not a directory", targetDir)
		}

		entries, readErr := os.ReadDir(targetDir)
		if readErr != nil {
			return readErr
		}
		if len(entries) > 0 {
			return fmt.Errorf("%s is not empty", targetDir)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return os.MkdirAll(targetDir, 0o755)
}

func writeTemplate(kind ProjectKind, targetDir, modulePath string) error {
	templateFS, err := fs.Sub(embeddedTemplates, path.Join("templates", string(kind)))
	if err != nil {
		return err
	}

	replacer := strings.NewReplacer(
		appTemplateModule, modulePath,
		componentTemplateModule, modulePath,
		componentTemplateDefinitions, path.Join(modulePath, "definitions"),
		componentTemplateManifestName, defaultComponentName(modulePath, targetDir),
	)

	return fs.WalkDir(templateFS, ".", func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		destination := filepath.Join(targetDir, filepath.FromSlash(name))
		if d.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		if before, ok := strings.CutSuffix(destination, ".tpl"); ok {
			destination = before
		}

		content, readErr := fs.ReadFile(templateFS, name)
		if readErr != nil {
			return readErr
		}

		rendered := replacer.Replace(string(content))
		return os.WriteFile(destination, []byte(rendered), 0o644)
	})
}

func defaultComponentName(modulePath, targetDir string) string {
	name := modulePath
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	if len(strings.TrimSpace(name)) == 0 {
		name = filepath.Base(targetDir)
	}

	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	return name
}
