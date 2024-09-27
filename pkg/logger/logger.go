package logger

import "github.com/go-chi/chi/v5/middleware"

type Logger interface {
	middleware.LoggerInterface
	Debug(i ...any)
	Debugf(format string, args ...any)
	Info(i ...any)
	Infof(format string, args ...any)
	Warn(i ...any)
	Warnf(format string, args ...any)
	Error(i ...any)
	Errorf(format string, args ...any)
	Fatal(i ...any)
	Fatalf(format string, args ...any)
	Panic(i ...any)
	Panicf(format string, args ...any)
}
