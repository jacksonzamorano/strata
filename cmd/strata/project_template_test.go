package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateProjectAppWritesTemplateAndRunsTidy(t *testing.T) {
	oldRunner := goCommandRunner
	t.Cleanup(func() {
		goCommandRunner = oldRunner
	})

	var tidyDir string
	var tidyArgs []string
	goCommandRunner = func(ctx context.Context, directory string, args ...string) ([]byte, error) {
		tidyDir = directory
		tidyArgs = append([]string(nil), args...)
		return nil, nil
	}

	target := filepath.Join(t.TempDir(), "demo-app")
	result, err := GenerateProject(context.Background(), GenerateProjectOptions{
		Kind:      ProjectKindApp,
		Directory: target,
	})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}
	if result.ModulePath != "demo-app" {
		t.Fatalf("ModulePath = %q, want demo-app", result.ModulePath)
	}

	goMod, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) returned error: %v", err)
	}
	if !strings.Contains(string(goMod), "module demo-app") {
		t.Fatalf("go.mod did not include inferred module path:\n%s", goMod)
	}

	mainGo, err := os.ReadFile(filepath.Join(target, "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) returned error: %v", err)
	}
	if !strings.Contains(string(mainGo), "component-example") {
		t.Fatalf("main.go did not include component guidance:\n%s", mainGo)
	}

	if tidyDir != target {
		t.Fatalf("go mod tidy directory = %q, want %q", tidyDir, target)
	}
	if strings.Join(tidyArgs, " ") != "mod tidy" {
		t.Fatalf("go mod tidy args = %q, want %q", strings.Join(tidyArgs, " "), "mod tidy")
	}
}

func TestGenerateProjectComponentRewritesModuleImports(t *testing.T) {
	oldRunner := goCommandRunner
	t.Cleanup(func() {
		goCommandRunner = oldRunner
	})
	goCommandRunner = func(ctx context.Context, directory string, args ...string) ([]byte, error) {
		return nil, nil
	}

	target := filepath.Join(t.TempDir(), "echo-component")
	result, err := GenerateProject(context.Background(), GenerateProjectOptions{
		Kind:       ProjectKindComponent,
		Directory:  target,
		ModulePath: "example.com/acme/echo",
	})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}
	if result.ModulePath != "example.com/acme/echo" {
		t.Fatalf("ModulePath = %q, want example.com/acme/echo", result.ModulePath)
	}

	mainGo, err := os.ReadFile(filepath.Join(target, "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) returned error: %v", err)
	}
	if !strings.Contains(string(mainGo), `d "example.com/acme/echo/definitions"`) {
		t.Fatalf("main.go did not rewrite the definitions import:\n%s", mainGo)
	}

	definitionGo, err := os.ReadFile(filepath.Join(target, "definitions", "definition.go"))
	if err != nil {
		t.Fatalf("ReadFile(definition.go) returned error: %v", err)
	}
	if !strings.Contains(string(definitionGo), `Name:    "echo"`) {
		t.Fatalf("definition.go did not rewrite manifest name:\n%s", definitionGo)
	}
}

func TestGenerateProjectRejectsNonEmptyDirectory(t *testing.T) {
	oldRunner := goCommandRunner
	t.Cleanup(func() {
		goCommandRunner = oldRunner
	})

	target := filepath.Join(t.TempDir(), "demo-app")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := GenerateProject(context.Background(), GenerateProjectOptions{
		Kind:      ProjectKindApp,
		Directory: target,
	})
	if err == nil {
		t.Fatalf("expected error for non-empty directory")
	}
}

func TestGenerateProjectKeepsFilesWhenTidyFails(t *testing.T) {
	oldRunner := goCommandRunner
	t.Cleanup(func() {
		goCommandRunner = oldRunner
	})

	goCommandRunner = func(ctx context.Context, directory string, args ...string) ([]byte, error) {
		return []byte("network unavailable"), errors.New("exit status 1")
	}

	target := filepath.Join(t.TempDir(), "demo-component")
	result, err := GenerateProject(context.Background(), GenerateProjectOptions{
		Kind:      ProjectKindComponent,
		Directory: target,
	})
	if err == nil {
		t.Fatalf("expected error when go mod tidy fails")
	}
	if result == nil {
		t.Fatalf("expected partial result when go mod tidy fails")
	}

	var bootstrapErr *BootstrapError
	if !errors.As(err, &bootstrapErr) {
		t.Fatalf("expected BootstrapError, got %T", err)
	}

	if _, statErr := os.Stat(filepath.Join(target, "main.go")); statErr != nil {
		t.Fatalf("expected generated files to remain after tidy failure: %v", statErr)
	}
}
