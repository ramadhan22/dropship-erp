// File: backend/internal/middleware/metrics_middleware.go

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsData holds detailed metrics for monitoring
type MetricsData struct {
	sync.RWMutex
	StartTime        time.Time
	TotalRequests    int64
	RequestsByMethod map[string]int64
	RequestsByPath   map[string]int64
	RequestsByStatus map[int]int64
	ResponseTimes    []time.Duration
	MaxResponseTime  time.Duration
	MinResponseTime  time.Duration
	DatabaseQueries  int64
	CacheHits        int64
	CacheMisses      int64
}

// Global metrics instance
var appMetrics = &MetricsData{
	StartTime:        time.Now(),
	RequestsByMethod: make(map[string]int64),
	RequestsByPath:   make(map[string]int64),
	RequestsByStatus: make(map[int]int64),
	ResponseTimes:    make([]time.Duration, 0, 1000), // Keep last 1000 response times
	MinResponseTime:  time.Hour, // Initialize with high value
}

// MetricsMiddleware returns middleware that collects detailed metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Record metrics
		duration := time.Since(start)
		
		appMetrics.Lock()
		defer appMetrics.Unlock()
		
		appMetrics.TotalRequests++
		appMetrics.RequestsByMethod[c.Request.Method]++
		appMetrics.RequestsByPath[c.FullPath()]++
		appMetrics.RequestsByStatus[c.Writer.Status()]++
		
		// Update response times
		if len(appMetrics.ResponseTimes) >= 1000 {
			// Remove oldest entry (sliding window)
			appMetrics.ResponseTimes = appMetrics.ResponseTimes[1:]
		}
		appMetrics.ResponseTimes = append(appMetrics.ResponseTimes, duration)
		
		// Update min/max response times
		if duration > appMetrics.MaxResponseTime {
			appMetrics.MaxResponseTime = duration
		}
		if duration < appMetrics.MinResponseTime {
			appMetrics.MinResponseTime = duration
		}
	}
}

// GetAppMetrics returns current application metrics
func GetAppMetrics() MetricsData {
	appMetrics.RLock()
	defer appMetrics.RUnlock()
	
	// Create a copy to avoid concurrent access issues
	metrics := MetricsData{
		StartTime:        appMetrics.StartTime,
		TotalRequests:    appMetrics.TotalRequests,
		RequestsByMethod: make(map[string]int64),
		RequestsByPath:   make(map[string]int64),
		RequestsByStatus: make(map[int]int64),
		MaxResponseTime:  appMetrics.MaxResponseTime,
		MinResponseTime:  appMetrics.MinResponseTime,
		DatabaseQueries:  appMetrics.DatabaseQueries,
		CacheHits:        appMetrics.CacheHits,
		CacheMisses:      appMetrics.CacheMisses,
	}
	
	for k, v := range appMetrics.RequestsByMethod {
		metrics.RequestsByMethod[k] = v
	}
	for k, v := range appMetrics.RequestsByPath {
		metrics.RequestsByPath[k] = v
	}
	for k, v := range appMetrics.RequestsByStatus {
		metrics.RequestsByStatus[k] = v
	}
	
	// Calculate average response time from recent samples
	if len(appMetrics.ResponseTimes) > 0 {
		var total time.Duration
		for _, rt := range appMetrics.ResponseTimes {
			total += rt
		}
		metrics.ResponseTimes = []time.Duration{total / time.Duration(len(appMetrics.ResponseTimes))}
	}
	
	return metrics
}

// RecordDatabaseQuery increments database query counter
func RecordDatabaseQuery() {
	appMetrics.Lock()
	appMetrics.DatabaseQueries++
	appMetrics.Unlock()
}

// RecordCacheHit increments cache hit counter
func RecordCacheHit() {
	appMetrics.Lock()
	appMetrics.CacheHits++
	appMetrics.Unlock()
}

// RecordCacheMiss increments cache miss counter
func RecordCacheMiss() {
	appMetrics.Lock()
	appMetrics.CacheMisses++
	appMetrics.Unlock()
}

// GetMetricsHandler returns a gin handler that exposes metrics as JSON
func GetMetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := GetAppMetrics()
		performance := GetPerformanceMetrics()
		
		response := gin.H{
			"app_metrics": gin.H{
				"uptime_seconds":     time.Since(metrics.StartTime).Seconds(),
				"total_requests":     metrics.TotalRequests,
				"requests_by_method": metrics.RequestsByMethod,
				"requests_by_path":   metrics.RequestsByPath,
				"requests_by_status": metrics.RequestsByStatus,
				"max_response_time":  metrics.MaxResponseTime.Milliseconds(),
				"min_response_time":  metrics.MinResponseTime.Milliseconds(),
				"database_queries":   metrics.DatabaseQueries,
				"cache_hits":         metrics.CacheHits,
				"cache_misses":       metrics.CacheMisses,
			},
			"performance_metrics": gin.H{
				"request_count":       performance.RequestCount,
				"avg_response_time":   performance.AvgResponseTime,
				"slow_query_count":    performance.SlowQueryCount,
				"error_count":         performance.ErrorCount,
				"slow_query_threshold": performance.SlowQueryThreshold.String(),
			},
		}
		
		// Calculate cache hit rate
		totalCacheRequests := metrics.CacheHits + metrics.CacheMisses
		if totalCacheRequests > 0 {
			response["cache_hit_rate"] = float64(metrics.CacheHits) / float64(totalCacheRequests)
		}
		
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, response)
	}
}