package core

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

type AuthorizationProvider interface {
	GetAuthorization(sec string) *Authorization
	NewAuthorization(source string, nickname string) *Authorization
}

func (s *SQLiteStorage) GetAuthorization(sec string) *Authorization {
	a, err := UseAuthorization(s.db, sec)
	if err != nil {
		panic(err)
	}
	return a
}

func (s *SQLiteStorage) NewAuthorization(source string, nickname string) *Authorization {
	sec := makeSecret()
	a, err := CreateAuthorization(s.db, &nickname, sec, source)
	if err != nil {
		panic(err)
	}
	return a
}
