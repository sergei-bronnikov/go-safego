package safego

import "log"

// Logger is an interface for logging messages from safego.
// It can be implemented to provide custom logging behavior.
//
// Example:
//
//	type customLogger struct {
//	    logger *log.Logger
//	}
//
//	func (l *customLogger) Printf(format string, v ...interface{}) {
//	    l.logger.Printf(format, v...)
//	}
//
//	safego.SetLogger(&customLogger{logger: log.New(os.Stdout, "[SAFEGO] ", log.LstdFlags)})
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

// SetLogger sets a custom logger for safego to use.
// If nil is passed, logging will be disabled (noop logger).
// By default, safego uses Go's standard logger.
//
// Example:
//
//	type myLogger struct{}
//
//	func (l *myLogger) Printf(format string, v ...interface{}) {
//	    log.Printf("[SAFEGO] "+format, v...)
//	}
//
//	safego.SetLogger(&myLogger{})
func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}
	logger = &noopLogger{}
}
