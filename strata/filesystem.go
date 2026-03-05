package strata

import (
	"os"

	"github.com/jacksonzamorano/strata/core"
)

func (c *Container) ReadFile(name string) ([]byte, bool) {
	approved := c.hostService.RequestPermission(core.Permission{
		Action:    core.PermissionActionReadFile,
		Container: c.namespace,
		Scope:     &name,
	})
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
