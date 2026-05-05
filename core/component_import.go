package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
)

type ComponentExecuteCommand struct {
	CanonicalName    string
	WorkingDirectory string
	Command          string
	Args             []string
}

type ComponentBinaryImport struct {
	BinaryName string
}

func (i *ComponentBinaryImport) Setup() (*ComponentExecuteCommand, error) {
	proj := path.Base(i.BinaryName)
	return &ComponentExecuteCommand{
		CanonicalName: proj,
		Command:       i.BinaryName,
	}, nil
}
func ImportBinary(name string) *ComponentBinaryImport {
	return &ComponentBinaryImport{
		BinaryName: name,
	}
}

func buildGoPackage(pkg, output string) error {
	e := exec.Command("go", "build", "-o", output, pkg)
	b, err := e.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Build: '%s': '%s'", err.Error(), string(b))
	}
	return nil
}

type ComponentModuleImport struct {
	ModulePath   string
	Subdirectory string
}

type goModuleInfo struct {
	Path    string
	Version string
	Dir     string
	Replace *goModuleInfo
}

var componentBinarySafeName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func PrepareComponent(compModule, subdir string) (*ComponentExecuteCommand, error) {
	info, err := goModule(compModule)
	if err != nil {
		return nil, err
	}

	importPath := compModule
	if compModule != "" {
		importPath = path.Join(importPath, subdir)
	}

	cacheDir, err := componentBuildCacheDir()
	if err != nil {
		return nil, err
	}

	proj := path.Base(importPath)
	if proj == "." || proj == "/" {
		return nil, fmt.Errorf("invalid component import path %q", importPath)
	}
	version := info.Version
	if version == "" {
		version = "dev"
	}
	binaryName := componentBinarySafeName.ReplaceAllString(importPath+"@"+version, "_")
	output := filepath.Join(cacheDir, binaryName)

	if err := buildGoPackage(importPath, output); err != nil {
		return nil, err
	}

	workingDirectory := info.Dir
	if subdir != "" {
		workingDirectory = filepath.Join(workingDirectory, filepath.FromSlash(subdir))
	}

	return &ComponentExecuteCommand{
		Command:          output,
		CanonicalName:    proj,
		WorkingDirectory: workingDirectory,
	}, nil
}

func goModule(modulePath string) (*goModuleInfo, error) {
	cmd := exec.Command("go", "list", "-m", "-json", modulePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Go module lookup: '%s': '%s'", err.Error(), string(out))
	}

	var info goModuleInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, err
	}
	if info.Replace != nil && info.Replace.Dir != "" {
		info.Dir = info.Replace.Dir
	}
	if info.Dir == "" {
		return nil, fmt.Errorf("Go module lookup: module %q has no local directory; make sure it is required in go.mod", modulePath)
	}
	return &info, nil
}

func componentBuildCacheDir() (string, error) {
	tmp, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(tmp, "com.strata.component-build-cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}
