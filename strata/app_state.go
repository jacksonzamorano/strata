package strata

import (
	_ "embed"
	"log"

	"github.com/jacksonzamorano/tasks/strata/core"
	"github.com/jacksonzamorano/tasks/strata/internal/hosts"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var initScript []byte

type AppState struct {
	persistence core.PersistenceProvider
	components  map[string]*ComponentRunner
	host        *appHostService
}

func newAppState(bus core.HostBus) AppState {
	persistence, fresh := core.DefaultPersistence(string(initScript))
	bus.Initialize(persistence)
	hostService := newAppHostService(persistence, hosts.NewHostBusCoordinator(bus))

	if fresh {
		auth := persistence.Authorization.NewAuthorization("core", "Master")
		hostService.Info("Created initial token '%s'", auth.Secret)
		log.Printf("Initial token: %s", auth.Secret)
	}

	as := AppState{
		persistence: persistence,
		host:        hostService,
		components:  map[string]*ComponentRunner{},
	}

	return as
}
