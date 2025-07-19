package logutil

import (
	"context"
	"testing"
	"time"
)

func TestStructuredLogging(t *testing.T) {
	logger := NewLogger("test-service", DEBUG)

	// Test basic logging
	ctx := context.Background()
	ctx = WithCorrelationID(ctx, "test-correlation-123")
	ctx = WithUserID(ctx, "test-user")
	ctx = WithShop(ctx, "test-shop")

	logger.Info(ctx, "TestOperation", "This is a test message", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})

	// Test error logging
	logger.Error(ctx, "TestOperation", "This is an error message", nil, map[string]interface{}{
		"error_code": "E001",
	})

	// Test timer logging
	timer := logger.WithTimer(ctx, "TestTimerOperation")
	time.Sleep(10 * time.Millisecond)
	timer.Finish("Timer operation completed successfully")

	// Test operation logger
	opLogger := logger.WithOperation("TestOpLogger")
	opLogger.Info(ctx, "Operation-specific log message")

	// Test context extraction
	correlationID := GetCorrelationID(ctx)
	if correlationID != "test-correlation-123" {
		t.Errorf("Expected correlation ID 'test-correlation-123', got '%s'", correlationID)
	}

	userID := GetUserID(ctx)
	if userID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got '%s'", userID)
	}

	shop := GetShop(ctx)
	if shop != "test-shop" {
		t.Errorf("Expected shop 'test-shop', got '%s'", shop)
	}
}

func TestLogLevels(t *testing.T) {
	logger := NewLogger("test-service", WARN)

	ctx := context.Background()

	// These should not be logged due to level filter
	logger.Debug(ctx, "TestOp", "Debug message")
	logger.Info(ctx, "TestOp", "Info message")

	// These should be logged
	logger.Warn(ctx, "TestOp", "Warning message")
	logger.Error(ctx, "TestOp", "Error message", nil)
}

func TestDefaultLogger(t *testing.T) {
	ctx := WithNewCorrelationID(context.Background())

	Info(ctx, "TestOperation", "Using default logger")

	correlationID := GetCorrelationID(ctx)
	if correlationID == "" {
		t.Error("Expected correlation ID to be set")
	}
}
