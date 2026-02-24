package tasklib

type AppDependancyType int

const (
	AppDependancyTypeBinary AppDependancyType = iota
	AppDependancyTypeLocalProject
	AppDependancyTypeGit
)

type AppDependancy struct {
	url      string
	branch   string
	subdir   string
	dep_type AppDependancyType
}

func Binary(url string) AppDependancy {
	return AppDependancy{
		url:      url,
		dep_type: AppDependancyTypeBinary,
	}
}
func LocalProject(url string) AppDependancy {
	return AppDependancy{
		url:      url,
		dep_type: AppDependancyTypeLocalProject,
	}
}
func Git(url string, branch string) AppDependancy {
	return AppDependancy{
		url:      url,
		branch:   branch,
		dep_type: AppDependancyTypeGit,
	}
}
func GitSubdirectory(url string, branch string, subdir string) AppDependancy {
	return AppDependancy{
		url:      url,
		branch:   branch,
		subdir:   subdir,
		dep_type: AppDependancyTypeGit,
	}
}
func Import(deps ...AppDependancy) []AppDependancy {
	return deps
}
