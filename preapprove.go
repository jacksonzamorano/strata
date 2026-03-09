package strata

import "github.com/jacksonzamorano/strata/core"

func AllowAll(cmp core.ComponentManifest, pm core.PermissionAction) core.Permission {
	return core.Permission{
		Container: cmp.Name,
		Action:    pm,
		Scope:     "*",
	}
}

func AllowOne(cmp core.ComponentManifest, pm core.PermissionAction, scope string) core.Permission {
	return core.Permission{
		Container: cmp.Name,
		Action:    pm,
		Scope:     scope,
	}
}
