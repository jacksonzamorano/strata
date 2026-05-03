package strata

import "github.com/jacksonzamorano/strata/core"

var ImportLocal = core.ImportLocal
var ImportBinary = core.ImportBinary
var ImportGit = core.ImportGit
var ImportGitSubdirectory = core.ImportGitSubdirectory
var ImportModule = core.ImportModule
var ImportModuleSubdirectory = core.ImportModuleSubdirectory

func Import(deps ...core.ComponentImport) []core.ComponentImport {
	return deps
}
