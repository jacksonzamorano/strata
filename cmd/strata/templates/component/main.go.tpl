package main

import (
	d "example.com/strata-component/definitions"
	"github.com/jacksonzamorano/strata/component"
)

func echo(
	input *component.ComponentInput[d.EchoRequest, d.EchoResponse],
	ctx *component.ComponentContainer,
) *component.ComponentReturn[d.EchoResponse] {
	ctx.Logger.Log("Echo called")
	return input.Return(d.EchoResponse{
		Message: input.Body.Message,
	})
}

func main() {
	component.CreateComponent(
		d.Manifest,
		component.Mount(d.Echo, echo),
	).Start()
}
