package main

import (
	"fmt"

	"github.com/jacksonzamorano/tasks/tasklib"
)

type SayHelloData struct {
	Name string `query:"name"`
}
type SayHelloRespose struct {
	Message string
	Count   int
}

func sayHello(data SayHelloData, container *tasklib.Container) tasklib.TaskResult {
	oldName := container.Storage.GetString("username")
	container.Storage.SetString("username", data.Name)

	count := container.Storage.GetInt("count")
	count += 1
	container.Storage.SetInt("count", count)

	return tasklib.TaskResult{
		Success: true,
		Result: SayHelloRespose{
			Message: fmt.Sprintf("Hello, %s, bye %s", data.Name, oldName),
			Count:   count,
		},
	}
}

func reset(data tasklib.NoTaskBody, container *tasklib.Container) tasklib.TaskResult {
	container.Storage.SetInt("count", 0)
	container.Storage.SetString("username", "")
	return tasklib.TaskResult{
		Success: true,
		Result:  "Reset",
	}
}

func main() {
	as := tasklib.NewAppServer([]tasklib.Task{
		tasklib.UsePublicTask(sayHello),
		tasklib.UseTask(reset),
	})
	e := as.Start()
	panic(e)
}
