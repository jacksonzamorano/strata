package strata

type StartupTask struct {
	handler func(*TaskContext)
}

func (tt *StartupTask) Attach(ctx *TaskAttachContext) {
	tt.handler(ctx.TaskContextGlobal())
}

func NewStartupTask(handler func(*TaskContext)) Task {
	return NewTask(handler, &StartupTask{
		handler,
	})
}
