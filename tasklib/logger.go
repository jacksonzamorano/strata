package tasklib

import "log"

type ConsoleLogger struct{}

func (cl *ConsoleLogger) Info(v string, args ...any) {
	log.Printf(v, args...)
}

type AppServerLogger interface {
	Info(v string, args ...any)
}
