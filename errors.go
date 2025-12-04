package safego

import "fmt"

type PanicError struct {
	Value interface{}
	StackTrace string
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("goroutine panic: %v", e.Value)
}

type CancelError struct {
	Cause error
}

func (e *CancelError) Error() string {
	return fmt.Sprintf("goroutine cancelled: %v", e.Cause)
}
