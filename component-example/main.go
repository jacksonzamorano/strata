package main

import (
	"fmt"

	"github.com/jacksonzamorano/tasks/tasklib/component"
)

type SayRequest struct {
	Name string
}
type SayResponse struct {
	Said string
}

func sayFeature(r SayRequest, ctx *component.ComponentContext) *component.ComponentResultPayload {
	return component.Result(SayResponse{Said: fmt.Sprintf("Said %s", r.Name)})
}

func main() {
	component.CreateComponent("example", "1.0.0",
		component.CreateFunction("say", sayFeature),
	).Start()
}
