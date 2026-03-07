package strata

import (
	"context"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/jacksonzamorano/strata/core"
)

type TaskAttachContext struct {
	mux          *http.ServeMux
	authorizaton core.AuthorizationProvider
	components   map[string]*ComponentIO
	triggers     *RuntimeTriggers
	Logger       core.Logger
	Container    *Container
	Context      context.Context
}

func (tac *TaskAttachContext) TaskContextGlobal() *TaskContext {
	return BuildTaskContext(tac.Container, tac.Logger, tac.components, tac.Context)
}

func (tac *TaskAttachContext) TaskContext(ctx context.Context) *TaskContext {
	return BuildTaskContext(tac.Container, tac.Logger, tac.components, ctx)
}

func (tac *TaskAttachContext) HTTP(path string, handler http.HandlerFunc) {
	tac.mux.HandleFunc(path, handler)
}

func (tac *TaskAttachContext) Trigger(ns, name string, body func([]byte)) {
	tac.triggers.Add(ns, name, body)
}

func (tac *TaskAttachContext) VerifyAuthentication(secret string) bool {
	return tac.authorizaton.GetAuthorization(secret) != nil
}

type Task struct {
	Name           string
	Implementation TaskImpl
}

func NewTask(fn any, impl TaskImpl) Task {
	pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	if pc == nil {
		panic("Passed a non-function to a function type.")
	}
	name_ugly := pc.Name()
	last_idx := strings.LastIndex(name_ugly, ".")
	name := name_ugly[last_idx+1:]

	return Task{
		Name:           name,
		Implementation: impl,
	}
}

type TaskImpl interface {
	Attach(ctx *TaskAttachContext)
}
