package core

import (
	"database/sql"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(db *sql.DB) *SQLiteStorage {
	return &SQLiteStorage{
		db,
	}
}

func (as *SQLiteStorage) database() *sql.DB {
	return as.db
}

func (as *SQLiteStorage) Container(name string) Storage {
	container := &ContainerStorage{
		db:        as,
		namespace: name,
		keys:      map[string]struct{}{},
	}

	keys, err := GetStorageRowKeyNamesInNamespace(as.db, name)
	if err != nil {
		panic(err)
	}
	for k := range keys {
		container.keys[keys[k].Key] = struct{}{}
	}

	return container
}
