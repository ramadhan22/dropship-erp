// File: backend/internal/cache/cache_interface.go

package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the cache connection
	Close() error

	// Ping checks if cache is available
	Ping(ctx context.Context) error

	// GetMetrics returns cache statistics
	GetMetrics() CacheMetrics
}

// CacheMetrics holds cache performance metrics
type CacheMetrics struct {
	Hits      int64
	Misses    int64
	Sets      int64
	Deletes   int64
	Errors    int64
	LastError string
	HitRate   float64
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	RedisURL     string
	Password     string
	DB           int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DefaultTTL   time.Duration
}
