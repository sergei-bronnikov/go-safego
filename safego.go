// Package safego provides safe wrappers for goroutines with panic recovery,
// context cancellation support, and error handling.
//
// The package offers multiple patterns for launching goroutines:
//   - Go: Fire-and-forget with panic recovery and logging
//   - GoWithErrorHandler: Fire-and-forget with custom error handling
//   - ChanGo: Returns a channel to wait for completion
//   - ChanGoWithError: Returns a channel to wait for completion with error support
//
// All functions support optional context.Context for cancellation.
//
// Example usage:
//
//	// Simple usage
//	safego.Go(func() {
//	    fmt.Println("Hello from goroutine")
//	})
//
//	// With context cancellation
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	safego.Go(func() {
//	    // Long running task
//	}, ctx)
//
//	// Wait for completion
//	done := safego.ChanGoWithError(func() error {
//	    return doSomething()
//	})
//	result := <-done
//	if result.Error != nil {
//	    log.Printf("Error: %v", result.Error)
//	}
package safego

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
)

// Done represents the completion status of a goroutine.
// It contains any error that occurred during execution, including
// panics (as PanicError) and cancellations (as CancelError).
type Done struct {
	Error error
}

// Go launches a goroutine with automatic panic recovery.
// If a panic occurs, it will be logged using the configured logger.
// The function supports optional context for cancellation.
//
// Parameters:
//   - fn: The function to execute in a goroutine
//   - ctx: Optional context for cancellation support
//
// Example:
//
//	// Basic usage
//	safego.Go(func() {
//	    fmt.Println("Hello from goroutine")
//	})
//
//	// With context
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	safego.Go(func() {
//	    time.Sleep(10 * time.Second) // Will be cancelled
//	}, ctx)
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

// GoWithErrorHandler launches a goroutine that can return errors.
// Any error returned by fn, including panics and cancellations, will be passed to errorHandler.
// This is useful for fire-and-forget operations that need custom error handling.
//
// Parameters:
//   - fn: The function to execute that returns an error
//   - errorHandler: Callback function invoked when an error occurs
//   - ctx: Optional context for cancellation support
//
// Example:
//
//	safego.GoWithErrorHandler(
//	    func() error {
//	        return doSomethingRisky()
//	    },
//	    func(err error) {
//	        log.Printf("Error occurred: %v", err)
//	    },
//	)
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

// ChanGo launches a goroutine and returns a channel that signals completion.
// The channel receives a Done struct containing any error that occurred,
// including panics (as PanicError) and cancellations (as CancelError).
// The channel is buffered and will be closed after sending the result.
//
// Parameters:
//   - fn: The function to execute in a goroutine
//   - ctx: Optional context for cancellation support
//
// Returns:
//   - A buffered channel that will receive exactly one Done value
//
// Example:
//
//	done := safego.ChanGo(func() {
//	    time.Sleep(1 * time.Second)
//	    fmt.Println("Task completed")
//	})
//
//	result := <-done
//	if result.Error != nil {
//	    if panicErr, ok := result.Error.(*safego.PanicError); ok {
//	        fmt.Printf("Panic: %v\n", panicErr.Value)
//	    }
//	}
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

// ChanGoWithError launches a goroutine that can return errors and provides
// a channel to wait for completion. The channel receives a Done struct containing
// any error returned by fn, or errors from panics and cancellations.
// The channel is buffered and will be closed after sending the result.
//
// Parameters:
//   - fn: The function to execute that returns an error
//   - ctx: Optional context for cancellation support
//
// Returns:
//   - A buffered channel that will receive exactly one Done value
//
// Example:
//
//	done := safego.ChanGoWithError(func() error {
//	    return doSomething()
//	})
//
//	result := <-done
//	if result.Error != nil {
//	    log.Printf("Task failed: %v", result.Error)
//	}
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
