package tasklib

import (
	"database/sql"
	"os"

	_ "embed"

	"github.com/jacksonzamorano/tasks/tasklib/core"
	_ "github.com/mattn/go-sqlite3"
)

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

//go:embed init.sql
var initScript []byte

type AppState struct {
	storage    core.StorageProvider
	database   *sql.DB
	components map[string]*ComponentRunner
	logger     core.HostBusChannel
}

func newAppState(bus core.HostBus) AppState {
	logger := bus.Channel()
	db_url := os.Getenv("DATABASE_URL")
	if len(db_url) == 0 {
		db_url = "./tasklib.db"
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

	logger.Info("Connected to database using sqlite '%s'.", v)
	e, err := isDatabaseEmpty(db)
	if err != nil {
		panic(err)
	}

	as := AppState{
		storage:    core.NewSQLiteStorage(db),
		logger:     bus.Channel(),
		database:   db,
		components: map[string]*ComponentRunner{},
	}

	if e {
		logger.Info("Initializing database.")
		db.Exec(string(initScript))
		a := as.createAuthorization("core", "Master Token")
		logger.Info("Created primary authorization '%s'", a.Secret)
	}

	return as
}
