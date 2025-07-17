package service

import (
	"context"
	"time"
)

// ContextTimeouts holds standard timeout durations for different operations
var ContextTimeouts = struct {
	DatabaseOperation time.Duration
	ExternalAPICall   time.Duration
	FileProcessing    time.Duration
	BatchOperation    time.Duration
	QuickOperation    time.Duration
}{
	DatabaseOperation: 30 * time.Second,
	ExternalAPICall:   60 * time.Second,
	FileProcessing:    5 * time.Minute,
	BatchOperation:    10 * time.Minute,
	QuickOperation:    10 * time.Second,
}

// WithTimeout creates a context with a timeout for different operation types
func WithDatabaseTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ContextTimeouts.DatabaseOperation)
}

func WithExternalAPITimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ContextTimeouts.ExternalAPICall)
}

func WithFileProcessingTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ContextTimeouts.FileProcessing)
}

func WithBatchOperationTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ContextTimeouts.BatchOperation)
}

func WithQuickOperationTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ContextTimeouts.QuickOperation)
}
