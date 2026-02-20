package tasklib

import (
	"crypto/rand"
	"encoding/base64"
)

func makeSecret() string {
	b := make([]byte, 256)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (as *AppState) getAuthorization(sec string) *Authorization {
	a, err := UseAuthorization(as.database, sec)
	if err != nil {
		panic(err)
	}
	return a
}

func (as *AppState) createAuthorization(source string, nickname string) *Authorization {
	sec := makeSecret()
	a, err := CreateAuthorization(as.database, &nickname, sec, source)
	if err != nil {
		panic(err)
	}
	return a
}
