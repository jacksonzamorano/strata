package strata

func Binary(url string) AppDependency {
	return AppDependency{
		url:     url,
		depType: AppDependencyTypeBinary,
	}
}

func LocalProject(url string) AppDependency {
	return AppDependency{
		url:     url,
		depType: AppDependencyTypeLocalProject,
	}
}

func Git(url string, branch string) AppDependency {
	return AppDependency{
		url:     url,
		branch:  branch,
		depType: AppDependencyTypeGit,
	}
}

func GitSubdirectory(url string, branch string, subdir string) AppDependency {
	return AppDependency{
		url:     url,
		branch:  branch,
		subdir:  subdir,
		depType: AppDependencyTypeGit,
	}
}

func Import(deps ...AppDependency) []AppDependency {
	return deps
}
