package tasklib

import (
	"encoding/json"
	"errors"
	"time"
)

type Container struct {
	Storage       *ContainerStorage
	Logger        ContainerLogger
	Keychain      ContainerKeychainProvider
	Authorization *Authorization
	components    map[string]*ComponentRunner
	appState      *AppState
}

func wrapExecuteFunction[T any](c *Container, cname, fname string, args any) (*T, []byte, error) {
	if cmp, ok := c.components[cname]; ok {
		res := cmp.Execute(fname, args)
		if res == nil {
			return nil, nil, errors.New("Could not read response.")
		}
		if res.Success {
			var v T
			err := json.Unmarshal(res.Response, &v)
			if err != nil {
				return nil, res.Response, err
			}
			return &v, res.Response, nil
		} else {
			return nil, nil, errors.New(res.Error)
		}
	}
	return nil, nil, errors.New("Module not found.")
}

func ExecuteFunction[T any](c *Container, cname, fname string, args any) (*T, error) {
	id := makeId()
	start := time.Now()
	c.Logger.Event(EventKindComponentFunctionStarted, EventComponentFunctionStartedPayload{
		Id:        id,
		Component: cname,
		Function:  fname,
		Date:      start,
	})
	res, bytes, err := wrapExecuteFunction[T](c, cname, fname, args)
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
	return res, err
}

func (as *AppState) buildContainer(namespace string) *Container {
	return &Container{
		Storage:    as.storage.Container(namespace),
		Logger:     as.Logger.Container(namespace),
		Keychain:   newPlatformKeychain().Container(namespace),
		components: as.components,
	}
}
