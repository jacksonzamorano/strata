package strata

import (
	"crypto/rand"
	"encoding/base64"
)

func makeId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
