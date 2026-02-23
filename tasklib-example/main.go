package main

import (
	"fmt"

	cex "github.com/jacksonzamorano/tasks/componentexample/types"
	"github.com/jacksonzamorano/tasks/tasklib"
)

type SayHelloData struct {
	Name string `query:"name"`
}
type SayHelloRespose struct {
	Message string
	Count   int
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

	msg, err := tasklib.ExecuteFunction[cex.SayResponse](container, "componentexample", "say", cex.SayRequest{
		Name: data.Name,
	})
	if err != nil {
		return tasklib.Error(err.Error())
	}

	return tasklib.Done(SayHelloRespose{
		Message: fmt.Sprintf("Got '%s', old '%s'", msg.Said, oldName),
		Count:   count,
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
	return tasklib.Done("Reset.")
}

func main() {
	as := tasklib.NewAppServer([]tasklib.Task{
		tasklib.UsePublicTask(sayHello),
		tasklib.UseTask(getVisitorLog),
		tasklib.UseTask(reset),
	}, []string{
		"/Users/jackson/Developer/tasklib/component-example/componentexample",
	})
	e := as.Start()
	panic(e)
}
