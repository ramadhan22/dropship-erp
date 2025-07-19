package service

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_HundredRequestsPerMinute(t *testing.T) {
	// Create a rate limiter with 100 requests per minute
	rateLimiter := NewRateLimiter(100, time.Minute)
	ctx := context.Background()

	// Test making 100 requests quickly - should all succeed
	for i := 0; i < 100; i++ {
		if err := rateLimiter.Wait(ctx); err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}

	// Verify stats show 0 available requests
	available, max := rateLimiter.GetStats()
	if available != 0 {
		t.Errorf("Expected 0 available requests, got %d", available)
	}
	if max != 100 {
		t.Errorf("Expected max 100 requests, got %d", max)
	}
}

func TestRateLimiter_ResetOnNewMinute(t *testing.T) {
	// Create a rate limiter with 5 requests per minute for faster testing
	rateLimiter := NewRateLimiter(5, time.Minute)
	ctx := context.Background()

	// Use up all requests
	for i := 0; i < 5; i++ {
		if err := rateLimiter.Wait(ctx); err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}

	// Check that we have 0 available requests
	available, _ := rateLimiter.GetStats()
	if available != 0 {
		t.Errorf("Expected 0 available requests, got %d", available)
	}

	// Manually advance to next minute by updating the current minute
	rateLimiter.mu.Lock()
	rateLimiter.currentMinute++
	rateLimiter.mu.Unlock()

	// Now we should have full quota available again
	available, max := rateLimiter.GetStats()
	if available != 5 {
		t.Errorf("Expected 5 available requests after minute reset, got %d", available)
	}
	if max != 5 {
		t.Errorf("Expected max 5 requests, got %d", max)
	}
}

func TestRateLimiter_AllowSuccessfulRequests(t *testing.T) {
	rateLimiter := NewRateLimiter(3, time.Minute)
	ctx := context.Background()

	// First request should succeed
	if !rateLimiter.Allow(ctx) {
		t.Error("First request should succeed")
	}

	// Second request should succeed
	if !rateLimiter.Allow(ctx) {
		t.Error("Second request should succeed")
	}

	// Third request should succeed
	if !rateLimiter.Allow(ctx) {
		t.Error("Third request should succeed")
	}

	// Fourth request should fail (rate limit exceeded)
	if rateLimiter.Allow(ctx) {
		t.Error("Fourth request should fail due to rate limit")
	}

	// Check stats
	available, max := rateLimiter.GetStats()
	if available != 0 {
		t.Errorf("Expected 0 available requests, got %d", available)
	}
	if max != 3 {
		t.Errorf("Expected max 3 requests, got %d", max)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	rateLimiter := NewRateLimiter(1, time.Minute)

	// Use up the single request
	ctx := context.Background()
	if err := rateLimiter.Wait(ctx); err != nil {
		t.Fatalf("First request failed: %v", err)
	}

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should fail with context deadline exceeded
	if err := rateLimiter.Wait(ctx); err == nil {
		t.Error("Expected context deadline exceeded error")
	} else if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got %v", err)
	}
}
