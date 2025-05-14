# Mockoho Refactoring Plan

## 1. Remove Unused Code

### 1.1. Remove Unused `isBinaryContent` Function

The `isBinaryContent` function in `internal/proxy/proxy.go` (lines 288-364) is defined but not used anywhere in the codebase. This complex function can be safely removed.

### 1.2. Clean Up Comments

Remove non-informative comments like "// No need to log headers" for cleaner code.

## 2. Improve Inefficient Implementations

### 2.1. Optimize `PrependWriter` Implementation

While keeping the functionality of having newest logs at the top and putting space between each session, we can optimize the implementation:

**Current Implementation Issues:**

- Reads the entire file content and rewrites it for each log entry
- Inefficient for large log files
- Creates a new file handle for each write operation

**Proposed Solution:**

- Keep a small in-memory buffer of recent log entries (configurable size)
- Write to the file in batches rather than for each log entry
- Use a more efficient file handling approach that doesn't require reading the entire file each time

```go
// Example optimized implementation
type PrependWriter struct {
    filePath string
    buffer   [][]byte
    maxBuffer int
    mu       sync.Mutex
}

func NewPrependWriter(filePath string, bufferSize int) *PrependWriter {
    return &PrependWriter{
        filePath: filePath,
        buffer:   make([][]byte, 0, bufferSize),
        maxBuffer: bufferSize,
    }
}

func (w *PrependWriter) Write(p []byte) (n int, err error) {
    w.mu.Lock()
    defer w.mu.Unlock()

    // Add to buffer
    w.buffer = append(w.buffer, append([]byte{}, p...))

    // If buffer is full, flush to file
    if len(w.buffer) >= w.maxBuffer {
        if err := w.flush(); err != nil {
            return 0, err
        }
    }

    return len(p), nil
}

func (w *PrependWriter) flush() error {
    // Read existing content (only if file exists)
    var existingContent []byte
    if _, err := os.Stat(w.filePath); err == nil {
        existingContent, err = os.ReadFile(w.filePath)
        if err != nil {
            return err
        }
    }

    // Create or truncate the file
    file, err := os.Create(w.filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Write buffered entries in reverse order (newest first)
    for i := len(w.buffer) - 1; i >= 0; i-- {
        if _, err := file.Write(w.buffer[i]); err != nil {
            return err
        }
    }

    // Write existing content
    if len(existingContent) > 0 {
        if _, err := file.Write(existingContent); err != nil {
            return err
        }
    }

    // Clear buffer
    w.buffer = w.buffer[:0]

    return nil
}

// Close flushes any remaining entries
func (w *PrependWriter) Close() error {
    w.mu.Lock()
    defer w.mu.Unlock()

    if len(w.buffer) > 0 {
        return w.flush()
    }

    return nil
}
```

### 2.2. Optimize Router Setup in Server Component

The `setupRoutes` function in `internal/server/server.go` recreates the entire router each time it's called. We should refactor this to only update routes when necessary.

## 3. Consolidate Duplicate Code

### 3.1. Centralize CORS Handling

CORS handling is duplicated in both proxy and server components. Create a single CORS middleware:

```go
// internal/middleware/cors.go
package middleware

import "github.com/gin-gonic/gin"

// CORSMiddleware returns a middleware that adds CORS headers to all responses
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

        // Handle preflight OPTIONS requests
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

### 3.2. Consolidate Logging Functions

Multiple logging functions (`LogDebug`, `Info`, `Warn`) all check `IsDebugMode` before logging. Create a single helper function:

```go
// internal helper function
func logIfDebug(level, format string, args ...interface{}) {
    if IsDebugMode && Logger != nil {
        Logger.Println(formatMessage(level, format, args...))
    }
}

// Then update the logging functions
func LogDebug(format string, args ...interface{}) {
    logIfDebug("DEBUG", format, args...)
}

func Info(format string, args ...interface{}) {
    logIfDebug("INFO", format, args...)
}

func Warn(format string, args ...interface{}) {
    logIfDebug("WARN", format, args...)
}
```

## 4. Improve Code Structure

### 4.1. Standardize Error Handling

Create consistent error handling patterns across the codebase, especially in the config, mock, and proxy components.

### 4.2. Improve Configuration Management

The configuration loading and saving could be optimized to reduce file I/O operations.

## Implementation Approach

I recommend implementing these changes in the following order:

1. First, remove unused code (like `isBinaryContent`) as this is low-risk and immediately reduces code complexity.
2. Next, consolidate duplicate code (CORS handling, logging functions) to improve maintainability.
3. Then, improve inefficient implementations (optimize `PrependWriter`, optimize router setup).
4. Finally, improve code structure (standardize error handling, improve configuration management).

Each change should be made in isolation, with tests run after each change to ensure functionality is preserved.
