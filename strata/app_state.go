package strata

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var initScript []byte

type AppState struct {
	persistence core.PersistenceProvider
	components  map[string]*ComponentIO
	host        *HostIO
}

func newAppState() AppState {
	persistence, fresh := core.DefaultPersistence(string(initScript))
	hostCtx, hostCancel := context.WithCancel(context.Background())
	hostService := newAppHostService(persistence, hostio.NewIO(hostCtx, hostCancel, os.Stdin, os.Stdout))

	if fresh {
		auth := persistence.Authorization.NewAuthorization("core", "Master")
		hostService.Log("Created initial token '%s'", auth.Secret)
		log.Printf("Initial token: %s", auth.Secret)
	}

	as := AppState{
		persistence: persistence,
		host:        hostService,
		components:  map[string]*ComponentIO{},
	}

	return as
}
