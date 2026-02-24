package main

import (
	// "os"
	// "path"

	cex "github.com/jacksonzamorano/tasks/componentexample/types"
	"github.com/jacksonzamorano/tasks/tasklib"
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

func sayHello(data SayHelloData, container *tasklib.Container) *tasklib.TaskResult {
	oldName := container.Storage.GetString("username")
	container.Storage.SetString("username", data.Name)

	count := container.Storage.GetInt("count")
	count += 1
	container.Storage.SetInt("count", count)

	entityContainer := tasklib.NewEntityStorage[Visitor](container.Storage)
	entityContainer.Insert(Visitor{
		Name:        data.Name,
		CountAtTime: count,
	})

	msg, err := tasklib.ExecuteFunction[cex.SayResponse](container, "example", "say", cex.SayRequest{
		Name: data.Name,
	})
	if err != nil {
		return tasklib.Error(err.Error())
	}

	return tasklib.Done(SayHelloResponse{
		Name:          data.Name,
		OldName:       oldName,
		ComponentData: *msg,
		Counter:       count,
	})
}

func getVisitorLog(data tasklib.NoTaskBody, container *tasklib.Container) *tasklib.TaskResult {
	entityContainer := tasklib.NewEntityStorage[Visitor](container.Storage)
	allNonEmpty := entityContainer.Find(func(v Visitor) bool { return len(v.Name) > 0 })
	return tasklib.Done(allNonEmpty)
}

func reset(data tasklib.NoTaskBody, container *tasklib.Container) *tasklib.TaskResult {
	container.Storage.SetInt("count", 0)
	container.Storage.SetString("username", "")
	tasklib.ExecuteFunction[tasklib.NoTaskBody](container, "example", "reset", tasklib.NoTaskBody{})
	return tasklib.Done("Reset.")
}

func main() {
	// cd, _ := os.Getwd()
	as := tasklib.NewAppServer([]tasklib.Task{
		tasklib.UsePublicTask(sayHello),
		tasklib.UseTask(getVisitorLog),
		tasklib.UseTask(reset),
	}, tasklib.Import(
		// tasklib.Binary("component-example"),
		// tasklib.LocalProject(path.Join(path.Dir(cd), "component-example")),
		tasklib.GitSubdirectory("git@github.com:jacksonzamorano/tasklib.git", "main", "component-example"),
	))
	e := as.Start()
	panic(e)
}
