package strata

import (
	"errors"
	"reflect"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/keychain"
)

type Container struct {
	Storage  core.Storage
	Logger   core.Logger
	Keychain core.Keychain
	Terminal TerminalProvider

	permissions map[string]bool
	persistence core.PersistenceProvider
	hostService *HostIO
	components  map[string]*ComponentIO
	namespace   string
}

func NewEntityStorage[T any](c *Container) *ContainerEntityStorage[T] {
	var zero T
	t := reflect.TypeOf(zero)
	return &ContainerEntityStorage[T]{
		storage:   c.persistence.EntityStorage,
		namespace: c.namespace,
		kind:      t.String(),
	}
}

func wrapExecuteFunction(c *Container, cname, fname string, args any) ([]byte, error) {
	if cmp, ok := c.components[cname]; ok {
		res := cmp.Execute(fname, args)
		if res == nil {
			return nil, errors.New("Could not read response.")
		}
		if res.Success {
			return res.Response, nil
		}
		return nil, errors.New(res.Error)
	}
	return nil, errors.New("Module not found.")
}

func (c *Container) ExecuteFunction(cname, fname string, args any) ([]byte, error) {
	return wrapExecuteFunction(c, cname, fname, args)
}

func (as *AppState) buildContainer(namespace string) *Container {
	return &Container{
		Storage:     as.persistence.Storage.Container(namespace),
		Logger:      as.host.Container(namespace),
		Keychain:    keychain.PlatformKeychain.Container(namespace),
		Terminal:    TerminalProvider{&NativeTerminal{}},
		permissions: map[string]bool{},
		persistence: as.persistence,
		components:  as.components,
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
