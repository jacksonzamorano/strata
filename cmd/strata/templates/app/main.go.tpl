package main

import "github.com/jacksonzamorano/strata"

type HelloInput struct {
	Name string `query:"name"`
}

func sayHello(input HelloInput, ctx *strata.TaskContext) *strata.RouteResult {
	ctx.Logger.Log("Saying hello to %s", input.Name)
	return strata.RouteResultSuccess(map[string]any{
		"message": "hello " + input.Name,
	})
}

func main() {
	runtime := strata.NewRuntime([]strata.Task{
		strata.NewPublicRouteTask(sayHello),
	})

	// To add a component later, run `strata add <module-path>` from this app.
	panic(runtime.Start())
}
