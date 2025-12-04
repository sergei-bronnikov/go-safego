package safego

import "fmt"

// PanicError represents a panic that occurred during goroutine execution.
// It captures both the panic value and the stack trace at the moment of panic.
//
// Example:
//
//	if panicErr, ok := err.(*safego.PanicError); ok {
//	    fmt.Printf("Panic: %v\n", panicErr.Value)
//	    fmt.Printf("Stack:\n%s\n", panicErr.StackTrace)
//	}
type PanicError struct {
	Value      interface{} // The value passed to panic()
	StackTrace string      // The full stack trace at the moment of panic
}

// Error implements the error interface for PanicError.
// It returns a formatted string containing the panic value.
func (e *PanicError) Error() string {
	return fmt.Sprintf("goroutine panic: %v", e.Value)
}

// CancelError represents a context cancellation that occurred before or during goroutine execution.
// The Cause field contains the underlying context error (context.Canceled or context.DeadlineExceeded).
//
// Example:
//
//	if cancelErr, ok := err.(*safego.CancelError); ok {
//	    if errors.Is(cancelErr.Cause, context.DeadlineExceeded) {
//	        fmt.Println("Operation timed out")
//	    }
//	}
type CancelError struct {
	Cause error // The underlying context error
}

// Error implements the error interface for CancelError.
// It returns a formatted string containing the cancellation cause.
func (e *CancelError) Error() string {
	return fmt.Sprintf("goroutine cancelled: %v", e.Cause)
}
