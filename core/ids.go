package core

import (
	"crypto/rand"
	"encoding/base64"
)

func Identifier() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
