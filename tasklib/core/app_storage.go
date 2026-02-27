package core

import (
	"database/sql"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(db *sql.DB) StorageProvider {
	return &SQLiteStorage{
		db,
	}
}

func (as *SQLiteStorage) Container(name string) Storage {
	container := &ContainerStorage{
		db:        as.db,
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
