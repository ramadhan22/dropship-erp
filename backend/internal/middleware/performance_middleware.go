// File: backend/internal/middleware/performance_middleware.go

package middleware

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics holds performance monitoring data
type PerformanceMetrics struct {
	RequestCount       int64
	TotalResponseTime  int64 // in nanoseconds
	SlowQueryCount     int64
	ErrorCount         int64
	AvgResponseTime    float64 // in milliseconds
	SlowQueryThreshold time.Duration
}

// Global metrics instance
var globalMetrics PerformanceMetrics

// SetSlowQueryThreshold sets the threshold for slow queries
func SetSlowQueryThreshold(threshold time.Duration) {
	globalMetrics.SlowQueryThreshold = threshold
}

// PerformanceMiddleware returns a gin middleware that tracks request performance
func PerformanceMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		// Record metrics
		atomic.AddInt64(&globalMetrics.RequestCount, 1)
		responseTime := params.Latency
		atomic.AddInt64(&globalMetrics.TotalResponseTime, responseTime.Nanoseconds())

		// Check for slow queries
		if responseTime > globalMetrics.SlowQueryThreshold {
			atomic.AddInt64(&globalMetrics.SlowQueryCount, 1)
			log.Printf("SLOW QUERY: %s %s took %v (threshold: %v)",
				params.Method, params.Path, responseTime, globalMetrics.SlowQueryThreshold)
		}

		// Count errors
		if params.StatusCode >= 400 {
			atomic.AddInt64(&globalMetrics.ErrorCount, 1)
		}

		// Update average response time
		count := atomic.LoadInt64(&globalMetrics.RequestCount)
		total := atomic.LoadInt64(&globalMetrics.TotalResponseTime)
		if count > 0 {
			globalMetrics.AvgResponseTime = float64(total/int64(time.Millisecond)) / float64(count)
		}

		// Standard gin log format with timestamp
		return fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %s\n",
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			params.StatusCode,
			params.Latency,
			params.ClientIP,
			params.Method,
			params.Path,
		)
	})
}

// GetMetrics returns current performance metrics
func GetPerformanceMetrics() PerformanceMetrics {
	return PerformanceMetrics{
		RequestCount:       atomic.LoadInt64(&globalMetrics.RequestCount),
		TotalResponseTime:  atomic.LoadInt64(&globalMetrics.TotalResponseTime),
		SlowQueryCount:     atomic.LoadInt64(&globalMetrics.SlowQueryCount),
		ErrorCount:         atomic.LoadInt64(&globalMetrics.ErrorCount),
		AvgResponseTime:    globalMetrics.AvgResponseTime,
		SlowQueryThreshold: globalMetrics.SlowQueryThreshold,
	}
}
