package strata

import (
	"os"

	"github.com/jacksonzamorano/strata/core"
)

func (c *Container) ReadFile(name string) ([]byte, bool) {
	approved := c.HasPermission(core.PermissionActionReadFile, name)
	if !approved {
		return []byte{}, false
	}

	b, err := os.ReadFile(name)
	if err != nil {
		c.hostService.Log("Could not read '%s': '%s'", name, err.Error())
		return []byte{}, false
	}

	return b, true
}

func (c *Container) WriteFile(name string, contents []byte) bool {
	approved := c.HasPermission(core.PermissionActionWriteFile, name)
	if !approved {
		return false
	}

	err := os.WriteFile(name, contents, 0755)
	if err != nil {
		c.hostService.Log("Could not write file: %s", err.Error())
		return false
	}
	return true
}

func (c *Container) MakeDirectory(name string) bool {
	approved := c.HasPermission(core.PermissionActionMakeDirectory, name)
	if !approved {
		return false
	}

	err := os.MkdirAll(name, 0755)
	if err != nil {
		c.hostService.Log("Could not write file: %s", err.Error())
		return false
	}
	return true
}
