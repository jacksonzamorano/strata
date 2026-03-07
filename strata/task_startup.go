package strata

type StartupTask struct {
	handler func(*Container)
}

func (tt *StartupTask) Attach(ctx *TaskAttachContext) {
	tt.handler(ctx.Container)
}

func NewStartupTask(handler func(*Container)) Task {
	return NewTask(handler, &StartupTask{
		handler,
	})
}
