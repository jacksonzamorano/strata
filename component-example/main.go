package main

import (
	"time"

	d "github.com/jacksonzamorano/componentexample/definitions"
	"github.com/jacksonzamorano/strata/component"
)

func sayFeature(r d.SayRequest, ctx *component.ComponentContainer) (*d.SayResponse, error) {
	last := ctx.Storage.GetString("last")
	ctx.Storage.SetString("last", r.Name)
	ctx.Keychain.Set("last", r.Name)

	tenx := ctx.Storage.GetInt("tenx")
	tenx += 10
	ctx.Storage.SetInt("tenx", tenx)

	d.TestTrigger.Send(ctx, d.TriggerTest{
		Time: time.Now(),
	})

	return &d.SayResponse{
		CurrentValue: r.Name,
		LastValue:    last,
		TenXValue:    tenx,
	}, nil
}
func reset(r d.EmptyRequest, ctx *component.ComponentContainer) (*string, error) {
	ctx.Storage.SetString("last", "")
	ctx.Storage.SetInt("tenx", 0)
	ctx.Logger.Log("Reset!")
	result := "Done!"
	return &result, nil
}
func getSecret(r d.EmptyRequest, ctx *component.ComponentContainer) (*string, error) {
	ct, ok := ctx.ReadFile("/Users/jacksonzamorano/Downloads/Unknown.pdf")
	ctx.Logger.Log("ok: %s, ct: %s", ok, string(ct))

	sec, ok := ctx.RequestSecret("What's the secret?")
	if ok {
		ctx.Logger.Log("Got secret '%s'", sec)
	} else {
		ctx.Logger.Log("Didn't get secret")
	}
	return &sec, nil
}

func main() {
	component.CreateComponent(d.Manifest,
		component.Mount(d.SayFeature, sayFeature),
		component.Mount(d.Reset, reset),
		component.Mount(d.GetSecret, getSecret),
	).Start()
}
