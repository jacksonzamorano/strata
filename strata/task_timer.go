package strata

import "time"

type TimedEveryTask struct {
	duration time.Duration
	handler  func(*Container)
}

func (tt *TimedEveryTask) Attach(ctx *TaskAttachContext) {
	go func() {
		timer := time.Tick(tt.duration)
		for {
			select {
			case <-timer:
				tt.handler(ctx.Container)
			case <-ctx.Context.Done():
				return
			}
		}
	}()
}

func NewTimedTask(duration time.Duration, handler func(*Container)) Task {
	return NewTask(handler, &TimedEveryTask{
		duration,
		handler,
	})
}
