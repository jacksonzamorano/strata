module github.com/jacksonzamorano/strata-example

go 1.26.0

require github.com/jacksonzamorano/strata v0.0.0

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.34 // indirect
)

replace github.com/jacksonzamorano/strata => ../strata

require github.com/jacksonzamorano/componentexample v0.0.0

replace github.com/jacksonzamorano/componentexample => ../component-example
