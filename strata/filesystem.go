package strata

import "github.com/jacksonzamorano/tasks/strata/core"

func (c *Container) ReadFile(name string) ([]byte, bool) {
	approved := c.hostService.RequestPermission(core.Permission{
		Action:    core.PermissionActionReadFile,
		Container: c.namespace,
		Scope:     &name,
	})
	if !approved {
		return []byte{}, false
	}
	return []byte{}, true
}
