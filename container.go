package strata

import (
	"crypto/rand"
	"os"
	"path"
	"strings"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/keychain"
	"github.com/jacksonzamorano/strata/internal/runtimehost"
)

type Container struct {
	Storage      core.Storage
	Keychain     core.Keychain
	StorageDir   string
	temporaryDir string

	permissions map[core.PermissionAction]map[string]bool
	persistence core.PersistenceProvider
	hostService *runtimehost.HostService
	namespace   string
}

func buildContainer(host *runtimehost.HostService, persistence core.PersistenceProvider, namespace string, pms []core.Permission) *Container {

	cfgRoot, _ := os.UserConfigDir()
	tmpRoot, _ := os.UserCacheDir()

	cfgPath := path.Join(cfgRoot, "strata", "containers", namespace)
	tmpPath := path.Join(tmpRoot, "strata", "containers", namespace)
	os.MkdirAll(cfgPath, 0755)
	os.MkdirAll(tmpPath, 0755)

	cnt := &Container{
		Storage:      persistence.Storage.Container(namespace),
		Keychain:     keychain.PlatformKeychain.Container(namespace),
		StorageDir:   cfgPath,
		temporaryDir: tmpPath,
		permissions:  map[core.PermissionAction]map[string]bool{},
		persistence:  persistence,
		namespace:    namespace,
		hostService:  host,
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

func (c *Container) TemporaryFile() string {
	rand_name := rand.Text()
	return path.Join(c.temporaryDir, rand_name)
}

func (c *Container) GetStorage() core.Storage {
	return c.Storage
}

func (c *Container) GetKeychain() core.Keychain {
	return c.Keychain
}

func (c *Container) Namespace() string {
	return c.namespace
}
