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


// This is owned by the application
// and many channels will be created.
type HostBus interface {
	Initialize(storage *StorageProvider)
	Channel() HostBusChannel
}

// Must be async-safe.
// Use a channel or other primitive.
type HostBusChannel interface {
	Info(v string, args ...any)
	Event(ev EventKind, payload any)
	Container(namespace string) Logger
}

// A containerized logger.
type Logger interface {
	Log(v string, args ...any)
}

type ForeignComponent interface {
	ExecuteFunction(cname, fname string, args any) ([]byte, error)
}
