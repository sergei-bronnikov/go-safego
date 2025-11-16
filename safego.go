package safego

import (
	"context"
	"errors"
	"fmt"
	"log"
)

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

func ChanGo(fn func(), ctx ...context.Context) chan error {
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))
			}
			close(errChan)
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				errChan <- errors.New(fmt.Sprintf("goroutine cancelled: %v", c.Err()))
				return
			default:
				fn()
			}
		} else {
			fn()
		}
	}()
	return errChan
}

func ChanGoWithError(fn func() error, ctx ...context.Context) chan error {
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))
			}
			close(errChan)
		}()
		if len(ctx) > 0 {
			c := ctx[0]
			select {
			case <-c.Done():
				errChan <- errors.New(fmt.Sprintf("goroutine cancelled: %v", c.Err()))
				return
			default:
				if err := fn(); err != nil {
					errChan <- err
				}
			}
		} else {
			if err := fn(); err != nil {
				errChan <- err
			}
		}
	}()
	return errChan
}
