package strata

import "github.com/jacksonzamorano/strata/core"

func AllowAll(nm string, pm core.PermissionAction) core.Permission {
	return core.Permission{
		Container: nm,
		Action:    pm,
		Scope:     "*",
	}
}

func AllowOne(nm string, pm core.PermissionAction, scope string) core.Permission {
	return core.Permission{
		Container: nm,
		Action:    pm,
		Scope:     scope,
	}
}
