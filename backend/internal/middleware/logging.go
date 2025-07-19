package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
)

// CorrelationIDMiddleware adds a correlation ID to each request
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID from header, or generate a new one
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = c.GetHeader("X-Request-ID")
		}

		// Create context with correlation ID
		var ctx context.Context
		if correlationID != "" {
			ctx = logutil.WithCorrelationID(c.Request.Context(), correlationID)
		} else {
			ctx = logutil.WithNewCorrelationID(c.Request.Context())
			correlationID = logutil.GetCorrelationID(ctx)
		}

		// Add request ID and other context information
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			ctx = logutil.WithRequestID(ctx, requestID)
		}

		if userID := c.GetHeader("X-User-ID"); userID != "" {
			ctx = logutil.WithUserID(ctx, userID)
		}

		// Set the correlation ID in response header
		c.Header("X-Correlation-ID", correlationID)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Continue processing
		c.Next()
	}
}

// RequestLoggingMiddleware logs requests with structured logging
func RequestLoggingMiddleware() gin.HandlerFunc {
	logger := logutil.NewLogger("http-middleware", logutil.INFO)

	return func(c *gin.Context) {
		// Record start time
		startTime := c.Request.Context().Value("start_time")
		if startTime == nil {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "start_time", gin.H{"start": gin.H{"time": gin.H{"Now": func() gin.H { return gin.H{} }}}}))
		}

		// Log request start
		logger.Info(c.Request.Context(), "RequestStart", "HTTP request started", map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"query":       c.Request.URL.RawQuery,
			"user_agent":  c.Request.UserAgent(),
			"remote_addr": c.ClientIP(),
		})

		// Process request
		c.Next()

		// Log request completion
		logger.Info(c.Request.Context(), "RequestComplete", "HTTP request completed", map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      c.Writer.Status(),
			"size":        c.Writer.Size(),
			"user_agent":  c.Request.UserAgent(),
			"remote_addr": c.ClientIP(),
		})
	}
}
