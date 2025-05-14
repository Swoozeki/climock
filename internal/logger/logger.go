package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	// Logger is the global logger instance
	Logger *log.Logger
	
	// IsDebugMode determines whether debug messages are logged
	IsDebugMode bool
	
	// MaxLogSize is the maximum size of the log file in bytes (5MB)
	MaxLogSize int64 = 5 * 1024 * 1024
	
	// BufferSize is the number of log entries to buffer before writing to file
	BufferSize = 10
)

// PrependWriter is a custom writer that prepends log entries to a file
type PrependWriter struct {
	filePath  string
	buffer    [][]byte
	mu        sync.Mutex
}

// Write implements the io.Writer interface
func (w *PrependWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Initialize buffer if needed
	if w.buffer == nil {
		w.buffer = make([][]byte, 0, BufferSize)
	}
	
	// Add to buffer
	w.buffer = append(w.buffer, append([]byte{}, p...))
	
	// If buffer is full, flush to file
	if len(w.buffer) >= BufferSize {
		if err := w.flush(); err != nil {
			return 0, err
		}
	}
	
	return len(p), nil
}

// flush writes the buffered log entries to the file
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

// Colors for console output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

// Init initializes the logger
func Init(debug bool) error {
	IsDebugMode = debug

	if debug {
		// In debug mode, log to debug.log file
		writer := &PrependWriter{filePath: "debug.log"}
		
		// Initialize the logger with the custom writer
		Logger = log.New(writer, "", 0)
		
		// Add a clear session separator with timestamp
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		divider := strings.Repeat("=", 50)
		separator := fmt.Sprintf("\n\n%s\n%s\n%s\n\n",
			divider,
			fmt.Sprintf("=== NEW SESSION STARTED AT %s ===", timestamp),
			divider)
		Logger.Println(separator)

		// Log initialization
		Info("Logger initialized, debug mode: %v", debug)

		// Trim the log file if it's too large
		go trimLogFile("debug.log", MaxLogSize)
	} else {
		// In non-debug mode, don't log to file
		Logger = log.New(io.Discard, "", 0)
	}

	return nil
}

// trimLogFile trims the log file to the specified maximum size
func trimLogFile(filePath string, maxSize int64) {
	// Check if the file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return
	}
	
	// If the file is smaller than the maximum size, do nothing
	if info.Size() <= maxSize {
		return
	}
	
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	
	// Calculate how much to keep (half of the max size)
	keepSize := maxSize / 2
	if int64(len(content)) > keepSize {
		// Keep only the first part of the file
		content = content[:keepSize]
	}
	
	// Write the trimmed content back to the file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		// We can't use Error() here as it would cause a recursive call
		fmt.Printf("Failed to write trimmed log file: %v\n", err)
	}
}

// Close logs a shutdown message
func Close() {
	Info("Logger shutting down")
	
	// Flush any buffered log entries
	if Logger != nil {
		if writer, ok := Logger.Writer().(*PrependWriter); ok && writer != nil {
			writer.mu.Lock()
			defer writer.mu.Unlock()
			
			if len(writer.buffer) > 0 {
				_ = writer.flush()
			}
		}
	}
}

// formatMessage formats a log message with timestamp, level, and caller info
func formatMessage(level, format string, args ...interface{}) string {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		// Extract just the package and file name, not the full path
		file = filepath.Base(file)
		caller = fmt.Sprintf("%s:%d", file, line)
	}
	
	// Format the message
	message := fmt.Sprintf(format, args...)
	
	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// Pad level to ensure consistent alignment
	paddedLevel := fmt.Sprintf("%-7s", level)
	
	// Format the full log entry
	return fmt.Sprintf("[%s] %s (%s) %s", timestamp, paddedLevel, caller, message)
}

// logIfDebug is a helper function that logs a message if debug mode is enabled
func logIfDebug(level, format string, args ...interface{}) {
	if IsDebugMode && Logger != nil {
		Logger.Println(formatMessage(level, format, args...))
	}
}

// LogDebug logs a debug message
func LogDebug(format string, args ...interface{}) {
	logIfDebug("DEBUG", format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	logIfDebug("INFO", format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	logIfDebug("WARN", format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Println(formatMessage("ERROR", format, args...))
	}
}

// Fatal logs a fatal message and exits
func Fatal(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Println(formatMessage("FATAL", format, args...))
	}
	os.Exit(1)
}

// HTTPRequest logs an HTTP request
func HTTPRequest(method, path, ip string, statusCode int, duration time.Duration) {
	if Logger == nil {
		return
	}
	
	level := "INFO"
	if statusCode >= 400 {
		level = "WARN"
	}
	if statusCode >= 500 {
		level = "ERROR"
	}
	
	Logger.Println(formatMessage(level, "%s %s from %s - %d (%s)", method, path, ip, statusCode, duration))
}

// ProxyError logs a proxy error
func ProxyError(target string, err error) {
	if Logger != nil {
		Logger.Println(formatMessage("ERROR", "Proxy error to %s: %v", target, err))
	}
}

// InitTestLogger initializes a logger for testing that doesn't write to any file
func InitTestLogger() {
	// Create a logger that writes to nowhere
	Logger = log.New(io.Discard, "", 0)
	IsDebugMode = false
}