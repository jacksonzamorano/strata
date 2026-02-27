package strata

import (
	_ "embed"
	"log"

	"github.com/jacksonzamorano/tasks/strata/core"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var initScript []byte

type AppState struct {
	persistence core.PersistenceProvider
	components  map[string]*ComponentRunner
	logger      core.HostBusChannel
}

func newAppState(bus core.HostBus) AppState {
	persistence, fresh := core.DefaultPersistence(string(initScript))
	bus.Initialize(persistence)

	logger := bus.Channel()
	if fresh {
		auth := persistence.Authorization.NewAuthorization("core", "Master")
		logger.Info("Created initial token '%s'", auth.Secret)
		log.Printf("Initial token: %s", auth.Secret)
	}

	as := AppState{
		persistence: persistence,
		logger:      bus.Channel(),
		components:  map[string]*ComponentRunner{},
	}

	return as
}
