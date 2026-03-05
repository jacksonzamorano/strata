package main

import (
	"time"

	"github.com/jacksonzamorano/componentexample/types"
	"github.com/jacksonzamorano/strata/component"
)

func sayFeature(r *component.ComponentInput[types.SayRequest, types.SayResponse], ctx *component.ComponentContainer) *component.ComponentReturn[types.SayResponse] {
	last := ctx.Storage.GetString("last")
	ctx.Storage.SetString("last", r.Body.Name)
	ctx.Keychain.Set("last", r.Body.Name)

	tenx := ctx.Storage.GetInt("tenx")
	tenx += 10
	ctx.Storage.SetInt("tenx", tenx)

	types.TestTrigger.Send(ctx, types.TriggerTest{
		Time: time.Now(),
	})

	return r.Return(types.SayResponse{
		CurrentValue: r.Body.Name,
		LastValue:    last,
		TenXValue:    tenx,
	})
}
func reset(r *component.ComponentInput[types.EmptyRequest, string], ctx *component.ComponentContainer) *component.ComponentReturn[string] {
	ctx.Storage.SetString("last", "")
	ctx.Storage.SetInt("tenx", 0)
	ctx.Logger.Log("Reset!")
	return r.Return("Done!")
}
func main() {
	component.CreateComponent("example", "1.0.2",
		component.Mount(types.SayFeature, sayFeature),
		component.Mount(types.Reset, reset),
	).Start()
}
