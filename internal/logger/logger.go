package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	// Logger is the global logger instance
	Logger *log.Logger
	
	// File is the log file
	File *os.File
	
	// IsDebugMode determines whether debug messages are logged
	IsDebugMode bool
)

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

	// Open the log file for appending
	var err error
	File, err = os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Initialize the logger
	Logger = log.New(File, "", 0)

	// Add a line break to separate sessions
	if _, err := File.WriteString("\n\n"); err != nil {
		return fmt.Errorf("failed to write session separator: %w", err)
	}

	// Log initialization
	Info("Logger initialized, debug mode: %v", debug)

	return nil
}

// Close closes the log file
func Close() {
	if File != nil {
		Info("Logger shutting down")
		File.Close()
		File = nil
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

// LogDebug logs a debug message
func LogDebug(format string, args ...interface{}) {
	if IsDebugMode {
		Logger.Println(formatMessage("DEBUG", format, args...))
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	Logger.Println(formatMessage("INFO", format, args...))
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	Logger.Println(formatMessage("WARN", format, args...))
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	Logger.Println(formatMessage("ERROR", format, args...))
}

// Fatal logs a fatal message and exits
func Fatal(format string, args ...interface{}) {
	Logger.Println(formatMessage("FATAL", format, args...))
	os.Exit(1)
}

// HTTPRequest logs an HTTP request
func HTTPRequest(method, path, ip string, statusCode int, duration time.Duration) {
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
	Logger.Println(formatMessage("ERROR", "Proxy error to %s: %v", target, err))
}