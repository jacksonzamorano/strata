package strata

type AppDependencyType int

const (
	AppDependencyTypeBinary AppDependencyType = iota
	AppDependencyTypeLocalProject
	AppDependencyTypeGit
)

type AppDependency struct {
	url     string
	branch  string
	subdir  string
	depType AppDependencyType
}
