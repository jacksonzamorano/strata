package strata

import "github.com/jacksonzamorano/tasks/strata/core"

func PreapprovedPermissions(pm ...core.ApprovedComponentPermission) *ConfigurationModification {
	return &ConfigurationModification{
		Permissions: pm,
	}
}
