package safego

import "log"

type Logger interface {
	Printf(format string, v ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

type noopLogger struct{}

func (l *noopLogger) Printf(format string, v ...interface{}) {}

var logger Logger = &defaultLogger{}

func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}
	logger = &noopLogger{}
}
