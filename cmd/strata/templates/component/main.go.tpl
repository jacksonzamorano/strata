package main

import (
	d "example.com/strata-component/definitions"
	"github.com/jacksonzamorano/strata/component"
)

func echo(input d.EchoRequest, ctx *component.ComponentContainer) (*d.EchoResponse, error) {
	ctx.Logger.Log("Echo called")
	return &d.EchoResponse{
		Message: input.Message,
	}, nil
}

func main() {
	component.CreateComponent(
		d.Manifest,
		component.Mount(d.Echo, echo),
	).Start()
}
