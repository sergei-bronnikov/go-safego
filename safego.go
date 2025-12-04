package safego

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
)

type Done struct {
	Error error
}

func Go(fn func(), ctx ...context.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("recovered from panic in goroutine: %v", r)
			}
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				logger.Printf("goroutine cancelled: %v", c.Err())
				return
			default:
				fn()
			}
		} else {
			fn()
		}
	}()
}

func GoWithErrorHandler(fn func() error, errorHandler func(error), ctx ...context.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				e := errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))
				errorHandler(e)
			}
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				e := errors.New(fmt.Sprintf("goroutine cancelled: %v", c.Err()))
				errorHandler(e)
				return
			default:
				if err := fn(); err != nil {
					errorHandler(err)
				}
			}
		} else {
			if err := fn(); err != nil {
				errorHandler(err)
			}
		}
	}()
}

func ChanGo(fn func(), ctx ...context.Context) chan Done {
	var err error
	doneCh := make(chan Done, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = &PanicError{
					Value:      r,
					StackTrace: string(debug.Stack()),
				}
			}
			doneCh <- Done{Error: err}
			close(doneCh)
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				err = &CancelError{Cause: c.Err()}
				return
			default:
				fn()
			}
		} else {
			fn()
		}
	}()
	return doneCh
}

func ChanGoWithError(fn func() error, ctx ...context.Context) chan Done {
	var err error
	doneCh := make(chan Done, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = &PanicError{
					Value:      r,
					StackTrace: string(debug.Stack()),
				}
			}
			doneCh <- Done{Error: err}
			close(doneCh)
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				err = &CancelError{Cause: c.Err()}
				return
			default:
				if e := fn(); e != nil {
					err = e
				}
			}
		} else {
			if e := fn(); e != nil {
				err = e
			}
		}
	}()
	return doneCh
}
