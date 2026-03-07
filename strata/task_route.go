package strata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type RouteTask struct {
	path    string
	secured bool
	handler func(*http.Request, *TaskContext) *RouteResult
}

type RouteTaskNoInput struct{}

type RouteTaskFunction[T any] func(input T, container *TaskContext) *RouteResult
type RouteResultStatus int

const (
	RouteResultStatusSuccess     RouteResultStatus = 200
	RouteResultStatusBadRequest  RouteResultStatus = 400
	RouteResultStatusServerError RouteResultStatus = 500
)

type RouteResult struct {
	status   RouteResultStatus
	response []byte
}

func RouteResultSuccess(data any) *RouteResult {
	enc, err := json.Marshal(data)
	if err != nil {
		return RouteRequestInvalid("Could not encode data: " + err.Error())
	}
	return &RouteResult{
		response: enc,
		status:   RouteResultStatusSuccess,
	}
}
func RouteRequestInvalid(error string) *RouteResult {
	return &RouteResult{
		status:   RouteResultStatusBadRequest,
		response: []byte(error),
	}
}

func newRouteTask[T any](secured bool, fn RouteTaskFunction[T]) Task {
	name_ugly := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	last_idx := strings.LastIndex(name_ugly, ".")
	name := name_ugly[last_idx+1:]

	return NewTask(
		fn,
		&RouteTask{
			path:    fmt.Sprintf("/tasks/%s", name),
			secured: secured,
			handler: func(req *http.Request, container *TaskContext) *RouteResult {
				body, err := io.ReadAll(req.Body)
				if req.Body != http.NoBody && err != nil {
					return RouteRequestInvalid("Could not read body.")
				}

				var input T
				json.Unmarshal(body, &input)
				parseQuery(req.URL.Query(), req.Header, &input)

				return fn(input, container)
			},
		},
	)
}

func NewRouteTask[T any](fn RouteTaskFunction[T]) Task {
	return newRouteTask(true, fn)
}
func NewPublicRouteTask[T any](fn RouteTaskFunction[T]) Task {
	return newRouteTask(false, fn)
}

func (tt *RouteTask) Attach(ctx *TaskAttachContext) {
	ctx.HTTP(tt.path, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				errorString := ""
				if e, ok := err.(error); ok {
					errorString = e.Error()
				}
				ctx.Logger.Log("Task crashed with error %s", errorString)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error."))
			}
		}()
		if tt.secured {
			c := r.Header.Get("Authorization")
			if !ctx.VerifyAuthentication(c) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized."))
				return
			}
		}
		res := tt.handler(r, ctx.TaskContext(r.Context()))
		w.WriteHeader(int(res.status))
		w.Write(res.response)
	})
}
