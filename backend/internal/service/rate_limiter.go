// File: backend/internal/service/rate_limiter.go

package service

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
// maxTokens: maximum number of tokens in the bucket
// refillRate: duration between token refills (e.g., time.Hour / 1000 for 1000 requests per hour)
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request can proceed
func (rl *RateLimiter) Allow(ctx context.Context) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}

// Wait blocks until a request can proceed or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow(ctx) {
			return nil
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(rl.refillRate):
			// Continue the loop to check again
		}
	}
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() (availableTokens int, maxTokens int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.tokens, rl.maxTokens
}