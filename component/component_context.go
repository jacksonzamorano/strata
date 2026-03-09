package component

import (
	"github.com/jacksonzamorano/strata/core"
	"github.com/jacksonzamorano/strata/internal/componentipc"
)

type ComponentContainer struct {
	Storage    core.Storage
	Keychain   core.Keychain
	Logger     core.Logger
	StorageDir string
	channel    *componentipc.IO
}

func (c *Component) buildContext() *ComponentContainer {
	return &ComponentContainer{
		Storage:    newComponentStorage(c.ioChannel),
		Keychain:   newComponentKeychain(c.ioChannel),
		Logger:     newComponentLogger(c.ioChannel),
		StorageDir: c.storageDir,
		channel:    c.ioChannel,
	}
}
