// File: backend/internal/cache/noop_cache.go

package cache

import (
	"context"
	"fmt"
	"time"
)

// NoopCache is a cache implementation that does nothing (for when caching is disabled)
type NoopCache struct{}

// NewNoopCache creates a new no-op cache
func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

// Get always returns cache miss
func (n *NoopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, fmt.Errorf("cache miss (noop cache)")
}

// Set does nothing
func (n *NoopCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (n *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Exists always returns false
func (n *NoopCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// Close does nothing
func (n *NoopCache) Close() error {
	return nil
}

// Ping does nothing
func (n *NoopCache) Ping(ctx context.Context) error {
	return nil
}

// GetMetrics returns empty metrics
func (n *NoopCache) GetMetrics() CacheMetrics {
	return CacheMetrics{}
}
