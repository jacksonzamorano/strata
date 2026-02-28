package strata

import (
	"errors"
	"reflect"
	"time"

	"github.com/jacksonzamorano/tasks/strata/core"
)

type Container struct {
	Storage       core.Storage
	Logger        core.Logger
	Keychain      core.Keychain
	Authorization *core.Authorization
	filesystem    *core.Filesystem
	persistence   core.PersistenceProvider
	busLogger     core.HostBusChannel
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
	id := makeId()
	start := time.Now()
	c.busLogger.Event(core.EventKindComponentFunctionStarted, core.EventComponentFunctionStartedPayload{
		Id:        id,
		Component: cname,
		Function:  fname,
		Date:      start,
	})
	bytes, err := wrapExecuteFunction(c, cname, fname, args)
	end := time.Now()
	if err == nil {
		c.busLogger.Event(core.EventKindComponentFunctionFinished, core.EventComponentFunctionFinishedPayload{
			Id:        id,
			Component: cname,
			Function:  fname,
			Date:      end,
			Duration:  end.Sub(start).Seconds(),
			Succeeded: true,
			Value:     string(bytes),
		})
	} else {
		c.busLogger.Event(core.EventKindComponentFunctionFinished, core.EventComponentFunctionFinishedPayload{
			Id:        id,
			Component: cname,
			Function:  fname,
			Date:      end,
			Duration:  end.Sub(start).Seconds(),
			Succeeded: false,
			Value:     string(bytes),
			Error:     err.Error(),
		})
	}
	return bytes, err
}

func (as *AppState) buildContainer(namespace string) *Container {
	return &Container{
		Storage:     as.persistence.Storage.Container(namespace),
		Logger:      as.logger.Container(namespace),
		Keychain:    newPlatformKeychain().Container(namespace),
		persistence: as.persistence,
		components:  as.components,
		namespace:   namespace,
		busLogger:   as.logger,
	}
}
