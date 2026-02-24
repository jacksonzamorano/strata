package main

import (
	"github.com/jacksonzamorano/tasks/componentexample/types"
	"github.com/jacksonzamorano/tasks/tasklib/component"
)

func sayFeature(r types.SayRequest, ctx *component.ComponentContext) *component.ComponentResultPayload {
	last := ctx.Storage.GetString("last")
	ctx.Storage.SetString("last", r.Name)

	return component.Result(types.SayResponse{
		CurrentValue: r.Name,
		LastValue:    last,
	})
}

func main() {
	component.CreateComponent("example", "1.0.1",
		component.CreateFunction("say", sayFeature),
	).Start()
}
