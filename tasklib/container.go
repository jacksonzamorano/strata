package tasklib

import (
	"encoding/json"
	"errors"
)

type Container struct {
	Storage       *ContainerStorage
	Authorization *Authorization
	components    map[string]*ComponentRunner
}

func ExecuteFunction[T any](c *Container, cname, fname string, args any) (*T, error) {
	if cmp, ok := c.components[cname]; ok {
		res := cmp.Execute(fname, args)
		if res == nil {
			return nil, errors.New("Could not read response.")
		}
		if res.Success {
			var v T
			err := json.Unmarshal(res.Response, &v)
			if err != nil {
				return nil, err
			}
			return &v, nil
		} else {
			return nil, errors.New(res.Error)
		}
	}
	return nil, errors.New("Module not found.")
}

func (as *AppState) buildContainer(namespace string) *Container {
	return &Container{
		Storage:    as.storage.Container(namespace),
		components: as.components,
	}
}
