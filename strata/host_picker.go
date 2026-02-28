package strata

import (
	"github.com/jacksonzamorano/tasks/strata/core"
	"github.com/jacksonzamorano/tasks/strata/internal/hosts"
)

func UseConsole() *ConfigurationModification {
	return &ConfigurationModification{
		NewHost: hosts.NewConsoleHost,
	}
}

func UseWebSockets() *ConfigurationModification {
	return &ConfigurationModification{
		NewHost: func() core.HostBus {
			return hosts.NewWebHost(false)
		},
	}
}

func UseWebUI() *ConfigurationModification {
	return &ConfigurationModification{
		NewHost: func() core.HostBus {
			return hosts.NewWebHost(true)
		},
	}
}
