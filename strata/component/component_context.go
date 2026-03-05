package component

import "github.com/jacksonzamorano/strata/core"

type ComponentContainer struct {
	Storage  core.Storage
	Keychain core.Keychain
	Logger   core.Logger
}

func (c *Component) buildContext() *ComponentContainer {
	return &ComponentContainer{
		Storage:  newComponentStorage(c.ioChannel),
		Keychain: newComponentKeychain(c.ioChannel),
		Logger:   newComponentLogger(c.ioChannel),
	}
}
