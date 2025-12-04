# go-safego

[![Go Reference](https://pkg.go.dev/badge/github.com/sergei-bronnikov/go-safego.svg)](https://pkg.go.dev/github.com/sergei-bronnikov/go-safego)

A safe goroutine wrapper library for Go that provides panic recovery, context cancellation support, and error handling for concurrent operations.

## Features

- 🛡️ **Panic Recovery**: Automatically recovers from panics in goroutines
- 🎯 **Context Support**: Built-in context cancellation handling
- 📊 **Error Handling**: Multiple patterns for error handling in goroutines
- 📝 **Logging**: Configurable logging for debugging
- 🔄 **Channel-based Communication**: Wait for goroutine completion with channels
- 📚 **Type-safe Error Types**: Distinguish between panics and cancellations

## Installation

```bash
go get github.com/sergei-bronnikov/go-safego
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/sergei-bronnikov/go-safego"
)

func main() {
    // Simple fire-and-forget goroutine with panic recovery
    safego.Go(func() {
        fmt.Println("Hello from safe goroutine!")
    })
}
```

## API Documentation

### Go

```go
func Go(fn func(), ctx ...context.Context)
```

Launches a goroutine with panic recovery. If a panic occurs, it will be logged and the program will continue running.

**Parameters:**
- `fn`: Function to execute in a goroutine
- `ctx`: Optional context for cancellation support

**Example:**

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/sergei-bronnikov/go-safego"
)

func main() {
    // Basic usage
    safego.Go(func() {
        fmt.Println("Task executed")
    })

    // With context cancellation
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    safego.Go(func() {
        time.Sleep(5 * time.Second)
        fmt.Println("This won't be printed")
    }, ctx)

    // With panic - will be recovered and logged
    safego.Go(func() {
        panic("something went wrong")
    })

    time.Sleep(3 * time.Second)
}
```

### GoWithErrorHandler

```go
func GoWithErrorHandler(fn func() error, errorHandler func(error), ctx ...context.Context)
```

Launches a goroutine that can return errors. Errors are passed to the error handler function.

**Parameters:**
- `fn`: Function that returns an error
- `errorHandler`: Callback function to handle errors
- `ctx`: Optional context for cancellation support

**Example:**

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
    "github.com/sergei-bronnikov/go-safego"
)

func main() {
    safego.GoWithErrorHandler(
        func() error {
            // Simulate some work
            time.Sleep(1 * time.Second)
            return errors.New("task failed")
        },
        func(err error) {
            fmt.Printf("Error occurred: %v\n", err)
        },
    )

    // With context
    ctx, cancel := context.WithCancel(context.Background())
    
    safego.GoWithErrorHandler(
        func() error {
            time.Sleep(2 * time.Second)
            return nil
        },
        func(err error) {
            fmt.Printf("Handler called: %v\n", err)
        },
        ctx,
    )

    cancel() // Will trigger error handler with cancellation error
    time.Sleep(3 * time.Second)
}
```

### ChanGo

```go
func ChanGo(fn func(), ctx ...context.Context) chan Done
```

Launches a goroutine and returns a channel that signals completion. Useful when you need to wait for the goroutine to finish.

**Parameters:**
- `fn`: Function to execute
- `ctx`: Optional context for cancellation support

**Returns:**
- Channel that receives a `Done` struct when the goroutine completes

**Example:**

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/sergei-bronnikov/go-safego"
)

func main() {
    // Wait for completion
    done := safego.ChanGo(func() {
        time.Sleep(1 * time.Second)
        fmt.Println("Task completed")
    })

    result := <-done
    if result.Error != nil {
        fmt.Printf("Goroutine failed: %v\n", result.Error)
    } else {
        fmt.Println("Goroutine succeeded")
    }

    // Handle panic
    done = safego.ChanGo(func() {
        panic("critical error")
    })

    result = <-done
    if panicErr, ok := result.Error.(*safego.PanicError); ok {
        fmt.Printf("Panic occurred: %v\n", panicErr.Value)
        fmt.Printf("Stack trace:\n%s\n", panicErr.StackTrace)
    }

    // With context cancellation
    ctx, cancel := context.WithCancel(context.Background())
    
    done = safego.ChanGo(func() {
        time.Sleep(5 * time.Second)
    }, ctx)

    cancel() // Cancel immediately
    
    result = <-done
    if cancelErr, ok := result.Error.(*safego.CancelError); ok {
        fmt.Printf("Cancelled: %v\n", cancelErr.Cause)
    }
}
```

### ChanGoWithError

```go
func ChanGoWithError(fn func() error, ctx ...context.Context) chan Done
```

Similar to `ChanGo`, but for functions that return errors.

**Parameters:**
- `fn`: Function that returns an error
- `ctx`: Optional context for cancellation support

**Returns:**
- Channel that receives a `Done` struct when the goroutine completes

**Example:**

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
    "github.com/sergei-bronnikov/go-safego"
)

func main() {
    // Handle function errors
    done := safego.ChanGoWithError(func() error {
        time.Sleep(1 * time.Second)
        return errors.New("operation failed")
    })

    result := <-done
    if result.Error != nil {
        fmt.Printf("Error: %v\n", result.Error)
    }

    // Multiple goroutines
    results := make([]chan safego.Done, 0)
    
    for i := 0; i < 5; i++ {
        id := i
        ch := safego.ChanGoWithError(func() error {
            time.Sleep(time.Duration(id) * time.Second)
            if id%2 == 0 {
                return fmt.Errorf("task %d failed", id)
            }
            return nil
        })
        results = append(results, ch)
    }

    // Wait for all
    for i, ch := range results {
        result := <-ch
        if result.Error != nil {
            fmt.Printf("Task %d error: %v\n", i, result.Error)
        } else {
            fmt.Printf("Task %d completed successfully\n", i)
        }
    }
}
```

## Error Types

### PanicError

Represents a panic that occurred in a goroutine.

```go
type PanicError struct {
    Value      interface{}  // The panic value
    StackTrace string       // Full stack trace
}
```

**Example:**

```go
done := safego.ChanGo(func() {
    panic("something went wrong")
})

result := <-done
if panicErr, ok := result.Error.(*safego.PanicError); ok {
    fmt.Printf("Panic value: %v\n", panicErr.Value)
    fmt.Printf("Stack trace:\n%s\n", panicErr.StackTrace)
}
```

### CancelError

Represents a context cancellation.

```go
type CancelError struct {
    Cause error  // The underlying context error (context.Canceled or context.DeadlineExceeded)
}
```

**Example:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()

done := safego.ChanGo(func() {
    time.Sleep(5 * time.Second)
}, ctx)

result := <-done
if cancelErr, ok := result.Error.(*safego.CancelError); ok {
    fmt.Printf("Operation cancelled: %v\n", cancelErr.Cause)
}
```

## Custom Logger

By default, `safego` uses Go's standard logger. You can provide a custom logger:

```go
package main

import (
    "log"
    "os"
    "github.com/sergei-bronnikov/go-safego"
)

type customLogger struct {
    logger *log.Logger
}

func (l *customLogger) Printf(format string, v ...interface{}) {
    l.logger.Printf(format, v...)
}

func main() {
    // Create custom logger
    logger := &customLogger{
        logger: log.New(os.Stdout, "[SAFEGO] ", log.LstdFlags),
    }
    
    safego.SetLogger(logger)
    
    // Now all logs will use your custom logger
    safego.Go(func() {
        panic("this will be logged with custom logger")
    })
}
```

## Advanced Examples

### Worker Pool

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/sergei-bronnikov/go-safego"
)

func main() {
    const numWorkers = 5
    const numTasks = 20

    tasks := make(chan int, numTasks)
    var wg sync.WaitGroup

    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        workerID := i
        
        safego.Go(func() {
            defer wg.Done()
            for task := range tasks {
                fmt.Printf("Worker %d processing task %d\n", workerID, task)
                time.Sleep(100 * time.Millisecond)
            }
        })
    }

    // Send tasks
    for i := 0; i < numTasks; i++ {
        tasks <- i
    }
    close(tasks)

    wg.Wait()
    fmt.Println("All tasks completed")
}
```

### Batch Processing with Error Handling

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/sergei-bronnikov/go-safego"
)

func processBatch(items []int) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    results := make([]chan safego.Done, len(items))

    // Process all items concurrently
    for i, item := range items {
        id := item
        results[i] = safego.ChanGoWithError(func() error {
            // Simulate processing
            time.Sleep(time.Duration(id%3) * time.Second)
            
            if id%5 == 0 {
                return fmt.Errorf("item %d failed validation", id)
            }
            
            return nil
        }, ctx)
    }

    // Collect results
    successCount := 0
    errorCount := 0

    for i, ch := range results {
        result := <-ch
        if result.Error != nil {
            fmt.Printf("Item %d failed: %v\n", items[i], result.Error)
            errorCount++
        } else {
            successCount++
        }
    }

    fmt.Printf("Batch complete: %d succeeded, %d failed\n", successCount, errorCount)
}

func main() {
    items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    processBatch(items)
}
```

## Best Practices

1. **Use `Go` for fire-and-forget operations** where you don't need to wait for completion
2. **Use `ChanGo` or `ChanGoWithError`** when you need to wait for results or handle errors
3. **Always pass context** for long-running operations to enable cancellation
4. **Handle errors appropriately** - distinguish between PanicError, CancelError, and regular errors
5. **Set custom logger** in production for better observability
6. **Don't ignore the Done channel** - always read from channels returned by ChanGo/ChanGoWithError

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
