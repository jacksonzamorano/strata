package main

import (
	"fmt"
	"os"
	"path"
	"time"

	example "github.com/jacksonzamorano/componentexample/definitions"
	"github.com/jacksonzamorano/strata"
)

type SayHelloData struct {
	Name string `query:"name"`
}
type SayHelloResponse struct {
	Name          string
	OldName       string
	ComponentData example.SayResponse
	Counter       int
}

type Visitor struct {
	Name        string
	CountAtTime int
}

func sayHello(data SayHelloData, context *strata.TaskContext) *strata.RouteResult {
	oldName := context.Container.Storage.GetString("username")
	context.Container.Storage.SetString("username", data.Name)

	count := context.Container.Storage.GetInt("count")
	count += 1
	context.Container.Storage.SetInt("count", count)

	entityContainer := strata.NewEntityStorage[Visitor](context.Container)
	entityContainer.Insert(Visitor{
		Name:        data.Name,
		CountAtTime: count,
	})

	msg, _ := example.SayFeature.Execute(context.Container, example.SayRequest{
		Name: data.Name,
	})

	return strata.RouteResultSuccess(SayHelloResponse{
		Name:          data.Name,
		OldName:       oldName,
		ComponentData: msg,
		Counter:       count,
	})
}

func getVisitorLog(data strata.RouteTaskNoInput, context *strata.TaskContext) *strata.RouteResult {
	c, _ := context.Container.ReadFile("../README.md")
	entityContainer := strata.NewEntityStorage[Visitor](context.Container)
	allNonEmpty := entityContainer.Find(func(v Visitor) bool { return len(v.Name) > 0 })
	return strata.RouteResultSuccess(map[string]any{"v": allNonEmpty, "c": string(c)})
}

func reset(data strata.RouteTaskNoInput, context *strata.TaskContext) *strata.RouteResult {
	context.Container.Storage.SetInt("count", 0)
	context.Container.Storage.SetString("username", "")
	example.Reset.Execute(context.Container, example.EmptyRequest{})
	return strata.RouteResultSuccess("Reset.")
}

func getSecret(data strata.RouteTaskNoInput, context *strata.TaskContext) *strata.RouteResult {
	res, ok := example.GetSecret.Execute(context.Container, example.EmptyRequest{})
	return strata.RouteResultSuccess(fmt.Sprintf("%s: %v", res, ok))
}

func testTime(context *strata.TaskContext) {
	context.Container.Logger.LogLiteral("Timer hit!")
}

func testTrigger(data example.TriggerTest, context *strata.TaskContext) {
	context.Container.Logger.Log("Got '%s'", data.Time.String())
}

func main() {
	cd, _ := os.Getwd()
	rt := strata.NewRuntime([]strata.Task{
		strata.NewPublicRouteTask(sayHello),
		strata.NewRouteTask(getVisitorLog),
		strata.NewRouteTask(reset),
		strata.NewPublicRouteTask(getSecret),
		strata.NewTimedTask(2*time.Minute, testTime),
		strata.NewTriggerTask(example.TestTrigger, testTrigger),
	}, strata.Import(
		// strata.Binary("component-example"),
		strata.ImportLocal(path.Join(path.Dir(cd), "component-example")),
		// strata.ImportGitSubdirectory("git@github.com:jacksonzamorano/strata.git", "component-example"),
	))
	e := rt.Start()
	panic(e)
}
