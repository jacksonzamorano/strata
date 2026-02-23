package tasklib

import (
	"encoding/json"
	"errors"
)

type Container struct {
	Storage       *ContainerStorage
	Logger        ContainerLogger
	Authorization *Authorization
	components    map[string]*ComponentRunner
	appState      *AppState
}

func ExecuteFunction[T any](c *Container, cname, fname string, args any) (*T, error) {
	c.Logger.Info("Dispatching %s.%s", cname, fname)
	if cmp, ok := c.components[cname]; ok {
		res := cmp.Execute(fname, args)
		if res == nil {
			return nil, errors.New("Could not read response.")
		}
		if res.Success {
			var v T
			err := json.Unmarshal(res.Response, &v)
			if err != nil {
				c.Logger.Info("Could not decode %s.%s: '%s'", cname, fname, err.Error())
				return nil, err
			}
			c.Logger.Info("Executed %s.%s", cname, fname)
			return &v, nil
		} else {
			c.Logger.Info("Could not call %s.%s: '%s'", cname, fname, res.Error)
			return nil, errors.New(res.Error)
		}
	}
	c.Logger.Info("Could not call %s.%s: Module not found", cname, fname)
	return nil, errors.New("Module not found.")
}

func (as *AppState) buildContainer(namespace string) *Container {
	return &Container{
		Storage:    as.storage.Container(namespace),
		Logger:     as.Logger.Container(namespace),
		components: as.components,
	}
}
