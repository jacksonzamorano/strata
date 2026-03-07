package strata

import (
	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/keychain"
)

type Container struct {
	Storage  core.Storage
	Keychain core.Keychain

	permissions map[string]bool
	persistence core.PersistenceProvider
	hostService *HostIO
	namespace   string
}

func (as *AppState) buildContainer(namespace string) *Container {
	return &Container{
		Storage:     as.persistence.Storage.Container(namespace),
		Keychain:    keychain.PlatformKeychain.Container(namespace),
		permissions: map[string]bool{},
		persistence: as.persistence,
		namespace:   namespace,
		hostService: as.host,
	}
}

func (c *Container) HasPermission(act core.PermissionAction, scope string) bool {
	p := core.Permission{
		Container: c.namespace,
		Action:    act,
		Scope:     &scope,
	}
	h := p.Hash()
	if v, ok := c.permissions[h]; ok {
		return v
	}
	approved := c.hostService.RequestPermission(p)
	c.permissions[h] = approved
	return approved
}
