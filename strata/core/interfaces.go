package core

import (
	"database/sql"
	"os"
	"time"
)

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

type PersistenceProvider struct {
	Storage            StorageProvider
	Authorization      AuthorizationProvider
	KVStorage          KVStorageProvider
	EntityStorage      EntityStorageProvider
	TaskHistoryStorage TaskHistoryProvider
}

func isDatabaseEmpty(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type='table'
		AND name NOT LIKE 'sqlite_%';
	`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func DefaultPersistence(initScript string) (PersistenceProvider, bool) {
	db_url := os.Getenv("DATABASE_URL")
	if len(db_url) == 0 {
		db_url = "./strata.db"
	}

	db, err := sql.Open("sqlite3", db_url)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	var v string
	_ = db.QueryRow(`SELECT sqlite_version()`).Scan(&v)

	e, err := isDatabaseEmpty(db)
	if err != nil {
		panic(err)
	}
	if e {
		db.Exec(initScript)
	}

	sqlite := NewSQLiteStorage(db)
	return PersistenceProvider{
		Storage:            sqlite,
		Authorization:      sqlite,
		KVStorage:          sqlite,
		EntityStorage:      sqlite,
		TaskHistoryStorage: sqlite,
	}, e
}

// This is owned by the application
// and many channels will be created.
type HostBus interface {
	Initialize(data PersistenceProvider)
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
