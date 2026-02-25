package main

import (
	"os"
	"path"

	cex "github.com/jacksonzamorano/tasks/componentexample/types"
	"github.com/jacksonzamorano/tasks/tasklib"
)

type SayHelloData struct {
	Name string `query:"name"`
}
type SayHelloResponse struct {
	Name          string
	OldName       string
	KeychainName  string
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
	keychainName := container.Keychain.Get("username")
	container.Keychain.Set("old_username", oldName)
	container.Keychain.Set("username", data.Name)

	count := container.Storage.GetInt("count")
	count += 1
	container.Storage.SetInt("count", count)

	entityContainer := tasklib.NewEntityStorage[Visitor](container)
	entityContainer.Insert(Visitor{
		Name:        data.Name,
		CountAtTime: count,
	})

	msg, _ := cex.SayFeature.Execute("example", container, cex.SayRequest{
		Name: data.Name,
	})

	return tasklib.Done(SayHelloResponse{
		Name:          data.Name,
		KeychainName:  keychainName,
		OldName:       oldName,
		ComponentData: msg,
		Counter:       count,
	})
}

func getVisitorLog(data tasklib.NoTaskBody, container *tasklib.Container) *tasklib.TaskResult {
	entityContainer := tasklib.NewEntityStorage[Visitor](container)
	allNonEmpty := entityContainer.Find(func(v Visitor) bool { return len(v.Name) > 0 })
	return tasklib.Done(allNonEmpty)
}

func reset(data tasklib.NoTaskBody, container *tasklib.Container) *tasklib.TaskResult {
	container.Storage.SetInt("count", 0)
	container.Storage.SetString("username", "")
	cex.Reset.Execute("example", container, cex.EmptyRequest{})
	return tasklib.Done("Reset.")
}

func main() {
	cd, _ := os.Getwd()
	as := tasklib.NewAppServer([]tasklib.Task{
		tasklib.UsePublicTask(sayHello),
		tasklib.UseTask(getVisitorLog),
		tasklib.UseTask(reset),
	}, tasklib.Import(
		// tasklib.Binary("component-example"),
		tasklib.LocalProject(path.Join(path.Dir(cd), "component-example")),
	))
	e := as.Start()
	panic(e)
}
