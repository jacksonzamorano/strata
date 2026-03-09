package strata

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/keychain"
)

type Container struct {
	Storage    core.Storage
	Keychain   core.Keychain
	StorageDir string

	permissions map[core.PermissionAction]map[string]bool
	persistence core.PersistenceProvider
	hostService *HostIO
	namespace   string
}

func (as *AppState) buildContainer(namespace string, pms []core.Permission) *Container {
	base, _ := os.UserConfigDir()
	storageDir := filepath.Join(base, "com.strata", "storage", namespace)
	os.MkdirAll(storageDir, 0755)
	cnt := &Container{
		Storage:     as.persistence.Storage.Container(namespace),
		Keychain:    keychain.PlatformKeychain.Container(namespace),
		StorageDir:  storageDir,
		permissions: map[core.PermissionAction]map[string]bool{},
		persistence: as.persistence,
		namespace:   namespace,
		hostService: as.host,
	}

	for i := range pms {
		if pms[i].Container != namespace {
			continue
		}
		if _, ok := cnt.permissions[pms[i].Action]; !ok {
			cnt.permissions[pms[i].Action] = map[string]bool{}
		}
		cnt.permissions[pms[i].Action][pms[i].Scope] = true
	}

	return cnt
}

func (c *Container) HasPermission(act core.PermissionAction, scope string) bool {
	p := core.Permission{
		Container: c.namespace,
		Action:    act,
		Scope:     scope,
	}
	if g, ok := c.permissions[act]; ok {
		if p, ok := g[scope]; ok && p {
			return true
		}
		if p, ok := g["*"]; ok && p {
			return true
		}
		for k, v := range g {
			if p, f := strings.CutSuffix(k, "*"); v && f && strings.HasPrefix(scope, p) {
				return true
			}
		}
	} else {
		c.permissions[act] = map[string]bool{}
	}
	approved := c.hostService.RequestPermission(p)
	c.permissions[act][scope] = approved
	return approved
}
