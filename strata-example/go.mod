module github.com/jacksonzamorano/strata-example

go 1.26.0

require github.com/jacksonzamorano/strata v0.0.0

require github.com/mattn/go-sqlite3 v1.14.34 // indirect

replace github.com/jacksonzamorano/strata => ..

require github.com/jacksonzamorano/componentexample v0.0.0

replace github.com/jacksonzamorano/componentexample => ../component-example
