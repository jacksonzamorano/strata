package tasklib

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
)

type TaskResult struct {
	Success    bool
	StatusCode int
	Result     any
}

type NoTaskBody struct{}

type TaskFn = func(request *RequestInfo, container *Container) *TaskResult
type TypedTaskFn[T any] = func(data T, container *Container) TaskResult

type Task struct {
	Name     string
	Function TaskFn
}

func buildTask[T any](fn TypedTaskFn[T], validation TaskFn) Task {
	fnReflect := reflect.ValueOf(fn)
	pc := fnReflect.Pointer()
	f := runtime.FuncForPC(pc)

	nameDirty := f.Name()
	nameIdx := strings.LastIndex(nameDirty, ".")
	name := nameDirty[nameIdx+1:]

	return Task{
		Name: name,
		Function: func(request *RequestInfo, container *Container) *TaskResult {
			if validation != nil {
				if err := validation(request, container); err != nil {
					return err
				}
			}

			var decodedBody T
			if len(request.Body) > 0 {
				err := json.Unmarshal(request.Body, &decodedBody)
				if err != nil {
					return &TaskResult{
						Success:    false,
						StatusCode: 400,
						Result:     "An invalid payload was provided.",
					}
				}
			}

			parseQuery(request.Query, request.Headers, &decodedBody)

			res := fn(decodedBody, container)
			return &res
		},
	}
}

func UseTask[T any](fn TypedTaskFn[T]) Task {
	return buildTask(fn, func(request *RequestInfo, container *Container) *TaskResult {
		if container.Authorization == nil || !container.Authorization.Active {
			return &TaskResult{
				Success:    false,
				StatusCode: 403,
				Result:     "Invalid authorization.",
			}
		}
		return nil
	})
}

func UsePublicTask[T any](fn TypedTaskFn[T]) Task {
	return buildTask(fn, nil)
}
