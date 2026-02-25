package tasklib

import (
	"errors"
	"reflect"
	"time"

	"github.com/jacksonzamorano/tasks/tasklib/core"
)

type Container struct {
	Storage       core.Storage
	Logger        ContainerLogger
	Keychain      core.Keychain
	Authorization *Authorization
	components    map[string]*ComponentRunner
	appState      *AppState
	namespace     string
}

func NewEntityStorage[T any](c *Container) *ContainerEntityStorage[T] {
	var zero T
	t := reflect.TypeOf(zero)
	return &ContainerEntityStorage[T]{
		db:        c.appState.database,
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
	c.Logger.Event(EventKindComponentFunctionStarted, EventComponentFunctionStartedPayload{
		Id:        id,
		Component: cname,
		Function:  fname,
		Date:      start,
	})
	bytes, err := wrapExecuteFunction(c, cname, fname, args)
	end := time.Now()
	if err != nil {
		c.Logger.Event(EventKindComponentFunctionFinished, EventComponentFunctionFinishedPayload{
			Id:        id,
			Component: cname,
			Function:  fname,
			Date:      end,
			Duration:  end.Sub(start).Seconds(),
			Succeeded: false,
			Value:     string(bytes),
			Error:     new(err.Error()),
		})
	} else {
		c.Logger.Event(EventKindComponentFunctionFinished, EventComponentFunctionFinishedPayload{
			Id:        id,
			Component: cname,
			Function:  fname,
			Date:      end,
			Duration:  end.Sub(start).Seconds(),
			Succeeded: true,
			Value:     string(bytes),
		})
	}
	return bytes, err
}

func (as *AppState) buildContainer(namespace string) *Container {
	storage := as.storage.Container(namespace)
	return &Container{
		Storage:    storage,
		Logger:     as.Logger.Container(namespace),
		Keychain:   newPlatformKeychain().Container(namespace),
		components: as.components,
		namespace:  namespace,
	}
}
