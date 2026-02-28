package main

import (
	"os"
	"path"

	cex "github.com/jacksonzamorano/tasks/componentexample/types"
	"github.com/jacksonzamorano/tasks/strata"
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

func sayHello(data SayHelloData, container *strata.Container) *strata.TaskResult {
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

	msg, _ := cex.SayFeature.Execute("example", container, cex.SayRequest{
		Name: data.Name,
	})

	return strata.Done(SayHelloResponse{
		Name:          data.Name,
		OldName:       oldName,
		ComponentData: msg,
		Counter:       count,
	})
}

func getVisitorLog(data strata.NoTaskBody, container *strata.Container) *strata.TaskResult {
	container.ReadFile("test.txt")
	entityContainer := strata.NewEntityStorage[Visitor](container)
	allNonEmpty := entityContainer.Find(func(v Visitor) bool { return len(v.Name) > 0 })
	return strata.Done(allNonEmpty)
}

func reset(data strata.NoTaskBody, container *strata.Container) *strata.TaskResult {
	container.Storage.SetInt("count", 0)
	container.Storage.SetString("username", "")
	cex.Reset.Execute("example", container, cex.EmptyRequest{})
	return strata.Done("Reset.")
}

func main() {
	cd, _ := os.Getwd()
	as := strata.NewAppServer([]strata.Task{
		strata.UsePublicTask(sayHello),
		strata.UseTask(getVisitorLog),
		strata.UseTask(reset),
	}, strata.Import(
		// strata.Binary("component-example"),
		strata.ImportLocal(path.Join(path.Dir(cd), "component-example")),
		// strata.ImportGitSubdirectory("git@github.com:jacksonzamorano/strata.git", "component-example"),
	), strata.UseWebUI())
	e := as.Start()
	panic(e)
}
