package main

import (
	"os"
	"path"
	"time"

	cex "github.com/jacksonzamorano/componentexample/definitions"
	"github.com/jacksonzamorano/strata"
)

type SayHelloData struct {
	Name string `query:"name"`
}
type SayHelloResponse struct {
	Name          string
	OldName       string
	ComponentData cex.SayResponse
	Counter       int
}

type Visitor struct {
	Name        string
	CountAtTime int
}

func sayHello(data SayHelloData, container *strata.Container) *strata.RouteResult {
	oldName := container.Storage.GetString("username")
	container.Storage.SetString("username", data.Name)

	count := container.Storage.GetInt("count")
	count += 1
	container.Storage.SetInt("count", count)

	entityContainer := strata.NewEntityStorage[Visitor](container)
	entityContainer.Insert(Visitor{
		Name:        data.Name,
		CountAtTime: count,
	})

	msg, _ := cex.SayFeature.Execute(cex.Manifest, container, cex.SayRequest{
		Name: data.Name,
	})

	return strata.RouteResultSuccess(SayHelloResponse{
		Name:          data.Name,
		OldName:       oldName,
		ComponentData: msg,
		Counter:       count,
	})
}

func getVisitorLog(data strata.RouteTaskNoInput, container *strata.Container) *strata.RouteResult {
	c, _ := container.ReadFile("../README.md")
	entityContainer := strata.NewEntityStorage[Visitor](container)
	allNonEmpty := entityContainer.Find(func(v Visitor) bool { return len(v.Name) > 0 })
	return strata.RouteResultSuccess(map[string]any{"v": allNonEmpty, "c": string(c)})
}

func reset(data strata.RouteTaskNoInput, container *strata.Container) *strata.RouteResult {
	container.Storage.SetInt("count", 0)
	container.Storage.SetString("username", "")
	cex.Reset.Execute(cex.Manifest, container, cex.EmptyRequest{})
	return strata.RouteResultSuccess("Reset.")
}

func testTime(container *strata.Container) {
	container.Logger.LogLiteral("Timer hit!")
}

func testTrigger(data cex.TriggerTest, container *strata.Container) {
	container.Logger.Log("Got '%s'", data.Time.String())
}

func main() {
	cd, _ := os.Getwd()
	rt := strata.NewRuntime([]strata.Task{
		strata.NewPublicRouteTask(sayHello),
		strata.NewRouteTask(getVisitorLog),
		strata.NewRouteTask(reset),
		strata.NewTimedTask(2*time.Minute, testTime),
		strata.NewTriggerTask(cex.Manifest, cex.TestTrigger, testTrigger),
	}, strata.Import(
		// strata.Binary("component-example"),
		strata.ImportLocal(path.Join(path.Dir(cd), "component-example")),
		// strata.ImportGitSubdirectory("git@github.com:jacksonzamorano/strata.git", "component-example"),
	))
	e := rt.Start()
	panic(e)
}
