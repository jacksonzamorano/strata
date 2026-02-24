package main

import (
	"github.com/jacksonzamorano/tasks/componentexample/types"
	"github.com/jacksonzamorano/tasks/tasklib/component"
)

func sayFeature(r types.SayRequest, ctx *component.ComponentContext) *component.ComponentResultPayload {
	last := ctx.Storage.GetString("last")
	ctx.Storage.SetString("last", r.Name)

	tenx := ctx.Storage.GetInt("tenx")
	tenx += 10
	ctx.Storage.SetInt("tenx", tenx)

	return component.Result(types.SayResponse{
		CurrentValue: r.Name,
		LastValue:    last,
		TenXValue:    tenx,
	})
}

func reset(r types.EmptyRequest, ctx *component.ComponentContext) *component.ComponentResultPayload {
	ctx.Storage.SetString("last", "")
	ctx.Storage.SetInt("tenx", 0)
	return component.Result("Done")
}

func main() {
	component.CreateComponent("example", "1.0.1",
		component.CreateFunction("say", sayFeature),
		component.CreateFunction("reset", reset),
	).Start()
}
