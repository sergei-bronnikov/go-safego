package safego_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/sergei-bronnikov/go-safego"
)

// Example demonstrates basic usage of Go function
func Example_go() {
	// Simple fire-and-forget goroutine
	safego.Go(func() {
		fmt.Println("Hello from safe goroutine!")
	})

	// Wait a bit for goroutine to complete
	time.Sleep(100 * time.Millisecond)
	// Output:
	// Hello from safe goroutine!
}

// Example_goWithContext demonstrates Go with context cancellation
func Example_goWithContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	safego.Go(func() {
		fmt.Println("This task completed before timeout")
	}, ctx)

	time.Sleep(200 * time.Millisecond)
	// Output:
	// This task completed before timeout
}

// Example_goWithErrorHandler demonstrates error handling with callback
func Example_goWithErrorHandler() {
	safego.GoWithErrorHandler(
		func() error {
			return errors.New("task failed")
		},
		func(err error) {
			fmt.Printf("Error occurred: %v\n", err)
		},
	)

	time.Sleep(100 * time.Millisecond)
	// Output:
	// Error occurred: task failed
}

// Example_chanGo demonstrates waiting for goroutine completion
func Example_chanGo() {
	done := safego.ChanGo(func() {
		fmt.Println("Task executing")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Task completed")
	})

	result := <-done
	if result.Error != nil {
		log.Printf("Error: %v", result.Error)
	} else {
		fmt.Println("Success!")
	}
	// Output:
	// Task executing
	// Task completed
	// Success!
}

// Example_chanGoWithError demonstrates error handling with channels
func Example_chanGoWithError() {
	done := safego.ChanGoWithError(func() error {
		return errors.New("operation failed")
	})

	result := <-done
	if result.Error != nil {
		fmt.Printf("Task failed: %v\n", result.Error)
	}
	// Output:
	// Task failed: operation failed
}

// Example_panicRecovery demonstrates panic recovery
func Example_panicRecovery() {
	done := safego.ChanGo(func() {
		panic("critical error")
	})

	result := <-done
	if panicErr, ok := result.Error.(*safego.PanicError); ok {
		fmt.Printf("Recovered from panic: %v\n", panicErr.Value)
	}
	// Output:
	// Recovered from panic: critical error
}

// Example_cancelError demonstrates context cancellation detection
func Example_cancelError() {
	ctx, cancel := context.WithCancel(context.Background())

	done := safego.ChanGo(func() {
		time.Sleep(5 * time.Second)
	}, ctx)

	// Cancel immediately
	cancel()

	result := <-done
	if cancelErr, ok := result.Error.(*safego.CancelError); ok {
		fmt.Printf("Operation cancelled: %v\n", cancelErr.Cause)
	}
	// Output:
	// Operation cancelled: context canceled
}

// Example_workerPool demonstrates a simple worker pool pattern
func Example_workerPool() {
	const numWorkers = 3
	const numTasks = 6

	tasks := make(chan int, numTasks)
	results := make([]chan safego.Done, 0, numWorkers)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		workerID := i
		done := safego.ChanGo(func() {
			for task := range tasks {
				fmt.Printf("Worker %d processing task %d\n", workerID, task)
			}
		})
		results = append(results, done)
	}

	// Send tasks
	for i := 0; i < numTasks; i++ {
		tasks <- i
	}
	close(tasks)

	// Wait for all workers
	for _, done := range results {
		<-done
	}

	fmt.Println("All tasks completed")
}

// Example_batchProcessing demonstrates concurrent batch processing with error collection
func Example_batchProcessing() {
	items := []int{1, 2, 3, 4, 5}
	results := make([]chan safego.Done, len(items))

	// Process all items concurrently
	for i, item := range items {
		id := item
		results[i] = safego.ChanGoWithError(func() error {
			if id%2 == 0 {
				return fmt.Errorf("item %d failed", id)
			}
			return nil
		})
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for _, ch := range results {
		result := <-ch
		if result.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("Success: %d, Failed: %d\n", successCount, errorCount)
	// Output:
	// Success: 3, Failed: 2
}

// customLogger is a custom logger implementation for demonstration
type customLogger struct{}

func (l *customLogger) Printf(format string, v ...interface{}) {
	log.Printf("[SAFEGO] "+format, v...)
}

// Example_customLogger demonstrates setting a custom logger
func Example_customLogger() {
	safego.SetLogger(&customLogger{})

	// Now all logs will use the custom logger
	safego.Go(func() {
		fmt.Println("Using custom logger")
	})

	time.Sleep(100 * time.Millisecond)
	// Output:
	// Using custom logger
}
