package logutil

import (
	"context"
	"github.com/google/uuid"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// CorrelationIDKey is the key for correlation ID in context
	CorrelationIDKey ContextKey = "correlation_id"
	
	// UserIDKey is the key for user ID in context
	UserIDKey ContextKey = "user_id"
	
	// RequestIDKey is the key for request ID in context
	RequestIDKey ContextKey = "request_id"
	
	// ShopKey is the key for shop name in context
	ShopKey ContextKey = "shop"
	
	// OperationKey is the key for operation name in context
	OperationKey ContextKey = "operation"
)

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithNewCorrelationID adds a new generated correlation ID to the context
func WithNewCorrelationID(ctx context.Context) context.Context {
	correlationID := uuid.New().String()
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithShop adds a shop name to the context
func WithShop(ctx context.Context, shop string) context.Context {
	return context.WithValue(ctx, ShopKey, shop)
}

// WithOperation adds an operation name to the context
func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, OperationKey, operation)
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetShop extracts shop name from context
func GetShop(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if shop, ok := ctx.Value(ShopKey).(string); ok {
		return shop
	}
	return ""
}

// GetOperation extracts operation name from context
func GetOperation(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if operation, ok := ctx.Value(OperationKey).(string); ok {
		return operation
	}
	return ""
}

// WithRequestContext creates a new context with request-specific information
func WithRequestContext(ctx context.Context, requestID, userID, shop string) context.Context {
	ctx = WithRequestID(ctx, requestID)
	ctx = WithUserID(ctx, userID)
	ctx = WithShop(ctx, shop)
	return WithNewCorrelationID(ctx)
}