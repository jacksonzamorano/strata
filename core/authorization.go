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
	return base64.URLEncoding.EncodeToString(b)
}

type AuthorizationProvider interface {
	GetAuthorization(sec string) *Authorization
	NewAuthorization(source string, nickname string) *Authorization
	GetAuthorizations() []Authorization
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

func (s *SQLiteStorage) GetAuthorizations() []Authorization {
	authorizations, err := GetAuthorizations(s.db)
	if err != nil {
		panic(err)
	}
	return authorizations
}
