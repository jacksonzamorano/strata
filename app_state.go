package strata

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/hostio"
	"github.com/jacksonzamorano/strata/internal/runtimecomponent"
	"github.com/jacksonzamorano/strata/internal/runtimehost"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var initScript []byte

type appState struct {
	persistence core.PersistenceProvider
	components  map[string]*runtimecomponent.Runner
	host        *runtimehost.HostService
}

func newAppState() *appState {
	persistence, fresh := core.DefaultPersistence(string(initScript))
	hostCtx, hostCancel := context.WithCancel(context.Background())
	hostService := runtimehost.NewHostService(persistence, hostio.NewIO(hostCtx, hostCancel, os.Stdin, os.Stdout))

	if fresh {
		auth := persistence.Authorization.NewAuthorization("core", "Master")
		hostService.Log("Created initial token '%s'", auth.Secret)
		log.Printf("Initial token: %s", auth.Secret)
	}

	as := &appState{
		persistence: persistence,
		host:        hostService,
		components:  map[string]*runtimecomponent.Runner{},
	}

	return as
}
