package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
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
