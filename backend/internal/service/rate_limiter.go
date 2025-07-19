// File: backend/internal/service/rate_limiter.go

package service

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a minute-based rate limiter for Shopee API
type RateLimiter struct {
	mu            sync.Mutex
	requestCount  int
	maxRequests   int
	currentMinute int64
}

// NewRateLimiter creates a new rate limiter for Shopee API
// maxRequests: maximum number of requests per minute (should be 100 for Shopee)
// refillRate parameter is ignored and kept for compatibility
func NewRateLimiter(maxRequests int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		requestCount:  0,
		maxRequests:   maxRequests,
		currentMinute: time.Now().Unix() / 60,
	}
}

// Allow checks if a request can proceed based on minute-based rate limiting
func (rl *RateLimiter) Allow(ctx context.Context) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	currentMinute := time.Now().Unix() / 60

	// Reset counter if we've moved to a new minute
	if currentMinute != rl.currentMinute {
		rl.currentMinute = currentMinute
		rl.requestCount = 0
	}

	// Check if we have requests available this minute
	if rl.requestCount < rl.maxRequests {
		rl.requestCount++
		return true
	}

	return false
}

// Wait blocks until a request can proceed or context is cancelled
// If rate limit is reached, waits until the next minute
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow(ctx) {
			return nil
		}

		// Calculate how long to wait until the next minute
		now := time.Now()
		nextMinute := now.Truncate(time.Minute).Add(time.Minute)
		waitDuration := nextMinute.Sub(now)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
			// Continue the loop to check again in the new minute
		}
	}
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() (availableRequests int, maxRequests int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	currentMinute := time.Now().Unix() / 60

	// Reset counter if we've moved to a new minute
	if currentMinute != rl.currentMinute {
		rl.currentMinute = currentMinute
		rl.requestCount = 0
	}

	available := rl.maxRequests - rl.requestCount
	if available < 0 {
		available = 0
	}

	return available, rl.maxRequests
}
