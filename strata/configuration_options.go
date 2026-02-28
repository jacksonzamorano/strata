package strata

import "github.com/jacksonzamorano/tasks/strata/core"

type ConfigurationModification struct {
	NewHost     func() core.HostBus
	Permissions []core.ApprovedComponentPermission
}

