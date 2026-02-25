package core

import "time"

type Storage interface {
	GetString(key string) string
	SetString(key string, value string) error
	GetInt(key string) int
	SetInt(key string, value int) error
	GetFloat(key string) float64
	GetBool(key string) bool
	GetDate(key string) time.Time
}

type StorageProvider interface {
	Container(namespace string) Storage
}

type Keychain interface {
	Get(key string) string
	Set(key, value string)
}

type KeychainProvider interface {
	Container(namespace string) Keychain
}

type Logger interface {
	Log(v string, args ...any)
}

type LoggerProvider interface {
	Info(v string, args ...any)
	Container(namespace string) Logger
}

type ForeignComponent interface {
	ExecuteFunction(cname, fname string, args any) ([]byte, error)
}
