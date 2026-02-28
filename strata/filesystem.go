package strata

import "github.com/jacksonzamorano/tasks/strata/core"

func (c *Container) ReadFile(name string) ([]byte, bool) {
	c.busLogger.RequestPermission(core.Permission{
		Action:    core.PermissionActionReadFile,
		Container: c.namespace,
		Scope:     &name,
	})
	return []byte{}, true
}
