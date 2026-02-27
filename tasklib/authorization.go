package tasklib

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/jacksonzamorano/tasks/tasklib/core"
)

func makeSecret() string {
	b := make([]byte, 256)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (as *AppState) getAuthorization(sec string) *core.Authorization {
	a, err := core.UseAuthorization(as.database, sec)
	if err != nil {
		panic(err)
	}
	return a
}

func (as *AppState) createAuthorization(source string, nickname string) *core.Authorization {
	sec := makeSecret()
	a, err := core.CreateAuthorization(as.database, &nickname, sec, source)
	if err != nil {
		panic(err)
	}
	return a
}
