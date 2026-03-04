package strata

import (
	"errors"
	"reflect"

	"github.com/jacksonzamorano/strata/core"
)

type Container struct {
	Storage       core.Storage
	Logger        core.Logger
	Keychain      core.Keychain
	Authorization *core.Authorization
	filesystem    *core.Filesystem
	persistence   core.PersistenceProvider
	hostService   *HostIOService
	components    map[string]*ComponentRunner
	namespace     string
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
		Keychain:    newPlatformKeychain().Container(namespace),
		persistence: as.persistence,
		components:  as.components,
		namespace:   namespace,
		hostService: as.host,
	}
}
