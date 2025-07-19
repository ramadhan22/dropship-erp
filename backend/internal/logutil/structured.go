package logutil

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents different levels of logging
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Level         LogLevel
	Message       string
	Timestamp     time.Time
	Service       string
	Operation     string
	CorrelationID string
	UserID        string
	Duration      *time.Duration
	Error         error
	Fields        map[string]interface{}
}

// Logger provides structured logging functionality
type Logger struct {
	level   LogLevel
	service string
}

// NewLogger creates a new structured logger for the given service
func NewLogger(service string, level LogLevel) *Logger {
	return &Logger{
		level:   level,
		service: service,
	}
}

// DefaultLogger is the default logger instance
var DefaultLogger = NewLogger("dropship-erp", INFO)

// getCorrelationID extracts correlation ID from context
func getCorrelationID(ctx context.Context) string {
	return GetCorrelationID(ctx)
}

// getUserID extracts user ID from context
func getUserID(ctx context.Context) string {
	return GetUserID(ctx)
}

// getCaller returns the caller information
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	// Extract just the filename from the full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// formatLogEntry formats a log entry for output
func (l *Logger) formatLogEntry(entry *LogEntry) string {
	var parts []string

	// Timestamp
	parts = append(parts, entry.Timestamp.Format("2006-01-02 15:04:05.000"))

	// Level
	parts = append(parts, fmt.Sprintf("[%s]", entry.Level.String()))

	// Service
	if entry.Service != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Service))
	}

	// Operation
	if entry.Operation != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Operation))
	}

	// Correlation ID
	if entry.CorrelationID != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.CorrelationID))
	}

	// Duration
	if entry.Duration != nil {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Duration.String()))
	}

	// Message
	parts = append(parts, entry.Message)

	// Error
	if entry.Error != nil {
		parts = append(parts, fmt.Sprintf("error=%v", entry.Error))
	}

	// Additional fields
	for k, v := range entry.Fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(parts, " ")
}

// shouldLog determines if a log entry should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// log outputs a log entry
func (l *Logger) log(entry *LogEntry) {
	if !l.shouldLog(entry.Level) {
		return
	}

	if entry.Service == "" {
		entry.Service = l.service
	}

	formatted := l.formatLogEntry(entry)
	log.Println(formatted)
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, operation, message string, fields ...map[string]interface{}) {
	entry := &LogEntry{
		Level:         DEBUG,
		Message:       message,
		Timestamp:     time.Now(),
		Service:       l.service,
		Operation:     operation,
		CorrelationID: getCorrelationID(ctx),
		UserID:        getUserID(ctx),
		Fields:        make(map[string]interface{}),
	}

	// Merge fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	l.log(entry)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, operation, message string, fields ...map[string]interface{}) {
	entry := &LogEntry{
		Level:         INFO,
		Message:       message,
		Timestamp:     time.Now(),
		Service:       l.service,
		Operation:     operation,
		CorrelationID: getCorrelationID(ctx),
		UserID:        getUserID(ctx),
		Fields:        make(map[string]interface{}),
	}

	// Merge fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	l.log(entry)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, operation, message string, fields ...map[string]interface{}) {
	entry := &LogEntry{
		Level:         WARN,
		Message:       message,
		Timestamp:     time.Now(),
		Service:       l.service,
		Operation:     operation,
		CorrelationID: getCorrelationID(ctx),
		UserID:        getUserID(ctx),
		Fields:        make(map[string]interface{}),
	}

	// Merge fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	l.log(entry)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, operation, message string, err error, fields ...map[string]interface{}) {
	entry := &LogEntry{
		Level:         ERROR,
		Message:       message,
		Timestamp:     time.Now(),
		Service:       l.service,
		Operation:     operation,
		CorrelationID: getCorrelationID(ctx),
		UserID:        getUserID(ctx),
		Error:         err,
		Fields:        make(map[string]interface{}),
	}

	// Merge fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	// Add caller information for errors
	entry.Fields["caller"] = getCaller(2)

	l.log(entry)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(ctx context.Context, operation, message string, err error, fields ...map[string]interface{}) {
	entry := &LogEntry{
		Level:         FATAL,
		Message:       message,
		Timestamp:     time.Now(),
		Service:       l.service,
		Operation:     operation,
		CorrelationID: getCorrelationID(ctx),
		UserID:        getUserID(ctx),
		Error:         err,
		Fields:        make(map[string]interface{}),
	}

	// Merge fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	// Add caller information for fatal errors
	entry.Fields["caller"] = getCaller(2)

	l.log(entry)
	Fatalf("Fatal error: %s", message)
}

// WithOperation creates a new logger with a specific operation context
func (l *Logger) WithOperation(operation string) *OperationLogger {
	return &OperationLogger{
		logger:    l,
		operation: operation,
	}
}

// WithTimer creates a new logger that tracks operation duration
func (l *Logger) WithTimer(ctx context.Context, operation string) *TimerLogger {
	return &TimerLogger{
		logger:    l,
		operation: operation,
		ctx:       ctx,
		startTime: time.Now(),
	}
}

// OperationLogger is a logger bound to a specific operation
type OperationLogger struct {
	logger    *Logger
	operation string
}

// Debug logs a debug message for the operation
func (ol *OperationLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) {
	ol.logger.Debug(ctx, ol.operation, message, fields...)
}

// Info logs an info message for the operation
func (ol *OperationLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) {
	ol.logger.Info(ctx, ol.operation, message, fields...)
}

// Warn logs a warning message for the operation
func (ol *OperationLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) {
	ol.logger.Warn(ctx, ol.operation, message, fields...)
}

// Error logs an error message for the operation
func (ol *OperationLogger) Error(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	ol.logger.Error(ctx, ol.operation, message, err, fields...)
}

// TimerLogger is a logger that tracks operation duration
type TimerLogger struct {
	logger    *Logger
	operation string
	ctx       context.Context
	startTime time.Time
}

// Finish logs the completion of the operation with duration
func (tl *TimerLogger) Finish(message string, fields ...map[string]interface{}) {
	duration := time.Since(tl.startTime)
	entry := &LogEntry{
		Level:         INFO,
		Message:       message,
		Timestamp:     time.Now(),
		Service:       tl.logger.service,
		Operation:     tl.operation,
		CorrelationID: getCorrelationID(tl.ctx),
		UserID:        getUserID(tl.ctx),
		Duration:      &duration,
		Fields:        make(map[string]interface{}),
	}

	// Merge fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	tl.logger.log(entry)
}

// FinishWithError logs the completion of the operation with an error
func (tl *TimerLogger) FinishWithError(message string, err error, fields ...map[string]interface{}) {
	duration := time.Since(tl.startTime)
	entry := &LogEntry{
		Level:         ERROR,
		Message:       message,
		Timestamp:     time.Now(),
		Service:       tl.logger.service,
		Operation:     tl.operation,
		CorrelationID: getCorrelationID(tl.ctx),
		UserID:        getUserID(tl.ctx),
		Duration:      &duration,
		Error:         err,
		Fields:        make(map[string]interface{}),
	}

	// Merge fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	// Add caller information for errors
	entry.Fields["caller"] = getCaller(2)

	tl.logger.log(entry)
}

// Convenience functions for default logger
func Debug(ctx context.Context, operation, message string, fields ...map[string]interface{}) {
	DefaultLogger.Debug(ctx, operation, message, fields...)
}

func Info(ctx context.Context, operation, message string, fields ...map[string]interface{}) {
	DefaultLogger.Info(ctx, operation, message, fields...)
}

func Warn(ctx context.Context, operation, message string, fields ...map[string]interface{}) {
	DefaultLogger.Warn(ctx, operation, message, fields...)
}

func Error(ctx context.Context, operation, message string, err error, fields ...map[string]interface{}) {
	DefaultLogger.Error(ctx, operation, message, err, fields...)
}

func Fatal(ctx context.Context, operation, message string, err error, fields ...map[string]interface{}) {
	DefaultLogger.Fatal(ctx, operation, message, err, fields...)
}
