package strata

import (
	"os"
	"path/filepath"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/keychain"
)

type Container struct {
	Storage    core.Storage
	Keychain   core.Keychain
	StorageDir string

	permissions map[string]bool
	persistence core.PersistenceProvider
	hostService *HostIO
	namespace   string
}

func (as *AppState) buildContainer(namespace string) *Container {
	base, _ := os.UserConfigDir()
	storageDir := filepath.Join(base, "com.strata", "storage", namespace)
	os.MkdirAll(storageDir, 0755)
	return &Container{
		Storage:     as.persistence.Storage.Container(namespace),
		Keychain:    keychain.PlatformKeychain.Container(namespace),
		StorageDir:  storageDir,
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
