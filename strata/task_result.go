package strata

type TaskResult struct {
	Success    bool
	StatusCode int
	Result     any
}

func Done(res any) *TaskResult {
	return &TaskResult{
		Success:    true,
		StatusCode: 200,
		Result:     res,
	}
}

func Error(res any) *TaskResult {
	if e, ok := res.(error); ok {
		res = e.Error()
	}

	return &TaskResult{
		Success:    false,
		StatusCode: 500,
		Result:     res,
	}
}

func Invalid(res any) *TaskResult {
	return &TaskResult{
		Success:    false,
		StatusCode: 400,
		Result:     res,
	}
}
