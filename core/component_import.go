package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type ComponentExecuteCommand struct {
	CanonicalName    string
	WorkingDirectory string
	Command          string
	Args             []string
}

type ComponentImport interface {
	Setup() (*ComponentExecuteCommand, error)
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

func buildGoProject(dir string) (string, error) {
	proj := path.Base(dir)
	e := exec.Command("go", "build", "-o", proj)
	e.Dir = dir
	b, err := e.CombinedOutput()
	if err != nil {
		return proj, fmt.Errorf("Build: '%s': '%s'", err.Error(), string(b))
	}
	return proj, nil
}

func buildGoPackage(pkg, output string) error {
	e := exec.Command("go", "build", "-o", output, pkg)
	b, err := e.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Build: '%s': '%s'", err.Error(), string(b))
	}
	return nil
}

type ComponentLocalProjectImport struct {
	Path string
}

func (i *ComponentLocalProjectImport) Setup() (*ComponentExecuteCommand, error) {
	nm, err := buildGoProject(i.Path)
	if err != nil {
		return nil, err
	}

	return &ComponentExecuteCommand{
		Command:          "./" + nm,
		CanonicalName:    nm,
		WorkingDirectory: i.Path,
	}, nil
}
func ImportLocal(path string) *ComponentLocalProjectImport {
	return &ComponentLocalProjectImport{
		Path: path,
	}
}

type ComponentGitProjectImport struct {
	Repository   string
	Subdirectory string
}

func (i *ComponentGitProjectImport) Setup() (*ComponentExecuteCommand, error) {
	path, err := checkoutGit(i.Repository, "", i.Subdirectory)

	nm, err := buildGoProject(path)
	if err != nil {
		return nil, err
	}

	return &ComponentExecuteCommand{
		Command:          "./" + nm,
		CanonicalName:    nm,
		WorkingDirectory: path,
	}, nil
}
func ImportGit(repository string) *ComponentGitProjectImport {
	return &ComponentGitProjectImport{
		Repository: repository,
	}
}
func ImportGitSubdirectory(repository, subdir string) *ComponentGitProjectImport {
	return &ComponentGitProjectImport{
		Repository:   repository,
		Subdirectory: subdir,
	}
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

func (i *ComponentModuleImport) Setup() (*ComponentExecuteCommand, error) {
	info, err := goModule(i.ModulePath)
	if err != nil {
		return nil, err
	}

	importPath := i.ModulePath
	if i.Subdirectory != "" {
		importPath = path.Join(importPath, i.Subdirectory)
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
	if i.Subdirectory != "" {
		workingDirectory = filepath.Join(workingDirectory, filepath.FromSlash(i.Subdirectory))
	}

	return &ComponentExecuteCommand{
		Command:          output,
		CanonicalName:    proj,
		WorkingDirectory: workingDirectory,
	}, nil
}

func ImportModule(modulePath string) *ComponentModuleImport {
	return &ComponentModuleImport{
		ModulePath: modulePath,
	}
}

func ImportModuleSubdirectory(modulePath, subdir string) *ComponentModuleImport {
	return &ComponentModuleImport{
		ModulePath:   modulePath,
		Subdirectory: subdir,
	}
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

func runGit(p string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = p
	txt, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(txt))
	}
	return nil
}
func checkoutGit(url, ref, subdir string) (string, error) {
	tmp, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	directoryName := path.Base(url)
	directoryName = strings.TrimSuffix(directoryName, ".git")

	importCachePath := path.Join(tmp, "com.strata.import-cache")
	_, err = os.Stat(importCachePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			os.MkdirAll(importCachePath, 0755)
		} else {
			return "", err
		}
	}

	checkout := path.Join(importCachePath, directoryName)
	_, err = os.Stat(checkout)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := runGit(importCachePath, "clone", url, checkout)
			if err != nil {
				return checkout, err
			}
		} else {
			return checkout, err
		}
	}

	err = runGit(checkout, "pull")
	if err != nil {
		return checkout, err
	}

	if len(ref) > 0 {
		err = runGit(checkout, "switch", ref)
		if err != nil {
			return checkout, err
		}
	}

	return path.Join(checkout, subdir), nil
}
