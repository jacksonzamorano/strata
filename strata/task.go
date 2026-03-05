package strata

import (
	"net/http"

	"github.com/jacksonzamorano/strata/core"
)

type TaskAttachContext struct {
	mux          *http.ServeMux
	authorizaton core.AuthorizationProvider
	Container    *Container
}

func (tac *TaskAttachContext) HTTP(path string, handler http.HandlerFunc) {
	tac.mux.HandleFunc(path, handler)
}

func (tac *TaskAttachContext) VerifyAuthentication(secret string) bool {
	return tac.authorizaton.GetAuthorization(secret) != nil
}

type Task struct {
	Name           string
	Implementation TaskImpl
}

type TaskImpl interface {
	Attach(ctx *TaskAttachContext)
}
