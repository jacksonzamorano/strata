package strata

import "github.com/jacksonzamorano/strata/core"

var ImportLocal = core.ImportLocal
var ImportBinary = core.ImportBinary
var ImportGit = core.ImportGit
var ImportGitSubdirectory = core.ImportGitSubdirectory

func Import(deps ...core.ComponentImport) []core.ComponentImport {
	return deps
}
