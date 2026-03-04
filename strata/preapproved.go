package strata

import "github.com/jacksonzamorano/strata/core"

func PreapprovedPermissions(pm ...core.ApprovedComponentPermission) *ConfigurationModification {
	return &ConfigurationModification{
		Permissions: pm,
	}
}
