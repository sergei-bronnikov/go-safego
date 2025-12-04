package safego

import (
	"context"
	"errors"
	"fmt"
	"log"
)

type Done struct {
	Error error
}

func Go(fn func(), ctx ...context.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("recovered from panic in goroutine:", r)
			}
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				log.Println("goroutine cancelled:", c.Err())
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
	doneCh := make(chan Done, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				doneCh <- Done{Error: errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))}
			}
			doneCh <- Done{Error: nil}
			close(doneCh)
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				doneCh <- Done{Error: errors.New(fmt.Sprintf("goroutine cancelled: %v", c.Err()))}
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
	doneCh := make(chan Done, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				doneCh <- Done{Error: errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))}
			}
			doneCh <- Done{Error: nil}
			close(doneCh)
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				doneCh <- Done{Error: errors.New(fmt.Sprintf("goroutine cancelled: %v", c.Err()))}
				return
			default:
				if err := fn(); err != nil {
					doneCh <- Done{Error: err}
				}
			}
		} else {
			if err := fn(); err != nil {
				doneCh <- Done{Error: err}
			}
		}
	}()
	return doneCh
}
