module github.com/jacksonzamorano/tasklib-example

go 1.26.0

require github.com/jacksonzamorano/tasks/tasklib v0.0.0

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.34 // indirect
)

replace github.com/jacksonzamorano/tasks/tasklib => ../tasklib

require github.com/jacksonzamorano/tasks/componentexample v0.0.0

replace github.com/jacksonzamorano/tasks/componentexample => ../component-example
