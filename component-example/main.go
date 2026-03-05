package main

import (
	"time"

	d "github.com/jacksonzamorano/componentexample/definitions"
	"github.com/jacksonzamorano/strata/component"
)

func sayFeature(r *component.ComponentInput[d.SayRequest, d.SayResponse], ctx *component.ComponentContainer) *component.ComponentReturn[d.SayResponse] {
	last := ctx.Storage.GetString("last")
	ctx.Storage.SetString("last", r.Body.Name)
	ctx.Keychain.Set("last", r.Body.Name)

	tenx := ctx.Storage.GetInt("tenx")
	tenx += 10
	ctx.Storage.SetInt("tenx", tenx)

	d.TestTrigger.Send(ctx, d.TriggerTest{
		Time: time.Now(),
	})

	return r.Return(d.SayResponse{
		CurrentValue: r.Body.Name,
		LastValue:    last,
		TenXValue:    tenx,
	})
}
func reset(r *component.ComponentInput[d.EmptyRequest, string], ctx *component.ComponentContainer) *component.ComponentReturn[string] {
	ctx.Storage.SetString("last", "")
	ctx.Storage.SetInt("tenx", 0)
	ctx.Logger.Log("Reset!")
	return r.Return("Done!")
}
func getSecret(r *component.ComponentInput[d.EmptyRequest, string], ctx *component.ComponentContainer) *component.ComponentReturn[string] {
	sec, ok := ctx.RequestSecret("What's the secret?")
	if ok {
		ctx.Logger.Log("Got secret '%s'", sec)
	} else {
		ctx.Logger.Log("Didn't get secret")
	}
	return r.Return(sec)
}

func main() {
	component.CreateComponent(d.Manifest,
		component.Mount(d.SayFeature, sayFeature),
		component.Mount(d.Reset, reset),
		component.Mount(d.GetSecret, getSecret),
	).Start()
}
