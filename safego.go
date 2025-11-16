package safego

import (
	"errors"
	"fmt"
	"log"
)

func Go(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in goroutine:", r)
		}
	}()
	fn()
}

func GoWithErrorHandler(fn func() error, errorHandler func(error)) {
	defer func() {
		if r := recover(); r != nil {
			e := errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))
			errorHandler(e)
		}
	}()
	if err := fn(); err != nil {
		errorHandler(err)
	}
}

func ChanGo(fn func()) chan error {
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))
			}
			close(errChan)
		}()
		fn()
	}()
	return errChan
}

func ChanGoWithError(fn func() error) chan error {
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.New(fmt.Sprintf("recovered from panic in goroutine: %v", r))
			}
			close(errChan)
		}()
		if err := fn(); err != nil {
			errChan <- err
		}
	}()
	return errChan
}
