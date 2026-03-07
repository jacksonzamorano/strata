package strata

import (
	"context"
	"time"
)

type TimedEveryTask struct {
	duration time.Duration
	handler  func(*TaskContext)
}

func (tt *TimedEveryTask) Attach(ctx *TaskAttachContext) {
	go func() {
		timer := time.Tick(tt.duration)
		for {
			select {
			case <-timer:
				cn, cnc := context.WithTimeout(ctx.Context, tt.duration)
				defer cnc()
				tt.handler(ctx.TaskContext(cn))
			case <-ctx.Context.Done():
				return
			}
		}
	}()
}

func NewTimedTask(duration time.Duration, handler func(*TaskContext)) Task {
	return NewTask(handler, &TimedEveryTask{
		duration,
		handler,
	})
}

type HourSpecificTask struct {
	hour    int
	minute  int
	handler func(*Container)
}

func (tt *HourSpecificTask) Attach(ctx *TaskAttachContext) {
	go func() {
		lastDay := -1
		timer := time.Tick(time.Second * 5)
		for {
			select {
			case <-timer:
				now := time.Now()
				if now.Minute() == tt.minute && now.Hour() == tt.hour && now.Day() != lastDay {
					tt.handler(ctx.Container)
					lastDay = now.Day()
				}
			case <-ctx.Context.Done():
				return
			}
		}
	}()
}

func NewTimeSpecificTask(hour int, minute int, handler func(*Container)) Task {
	return NewTask(handler, &HourSpecificTask{
		hour,
		minute,
		handler,
	})
}
