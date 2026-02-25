package tasklib

import (
	"database/sql"

	"github.com/jacksonzamorano/tasks/tasklib/core"
)

type AppStorage struct {
	db *sql.DB
}

func (as *AppStorage) Container(name string) core.Storage {
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
