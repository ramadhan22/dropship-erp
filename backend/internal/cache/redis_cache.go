// File: backend/internal/cache/redis_cache.go

package cache

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client   *redis.Client
	config   CacheConfig
	metrics  atomicMetrics
}

// atomicMetrics holds cache metrics with atomic operations for thread safety
type atomicMetrics struct {
	hits      int64
	misses    int64
	sets      int64
	deletes   int64
	errors    int64
	lastError atomic.Value
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config CacheConfig) (*RedisCache, error) {
	// Set defaults if not provided
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}

	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override with config values
	if config.Password != "" {
		opt.Password = config.Password
	}
	opt.DB = config.DB
	opt.MaxRetries = config.MaxRetries
	opt.DialTimeout = config.DialTimeout
	opt.ReadTimeout = config.ReadTimeout
	opt.WriteTimeout = config.WriteTimeout

	client := redis.NewClient(opt)

	cache := &RedisCache{
		client: client,
		config: config,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cache.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return cache, nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			atomic.AddInt64(&r.metrics.misses, 1)
			return nil, fmt.Errorf("cache miss for key %s", key)
		}
		atomic.AddInt64(&r.metrics.errors, 1)
		r.metrics.lastError.Store(err.Error())
		return nil, fmt.Errorf("cache get error: %w", err)
	}
	
	atomic.AddInt64(&r.metrics.hits, 1)
	return result, nil
}

// Set stores a value in cache with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.config.DefaultTTL
	}
	
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		atomic.AddInt64(&r.metrics.errors, 1)
		r.metrics.lastError.Store(err.Error())
		return fmt.Errorf("cache set error: %w", err)
	}
	
	atomic.AddInt64(&r.metrics.sets, 1)
	return nil
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		atomic.AddInt64(&r.metrics.errors, 1)
		r.metrics.lastError.Store(err.Error())
		return fmt.Errorf("cache delete error: %w", err)
	}
	
	atomic.AddInt64(&r.metrics.deletes, 1)
	return nil
}

// Exists checks if a key exists in cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		atomic.AddInt64(&r.metrics.errors, 1)
		r.metrics.lastError.Store(err.Error())
		return false, fmt.Errorf("cache exists error: %w", err)
	}
	
	return result > 0, nil
}

// Close closes the cache connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Ping checks if cache is available
func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		atomic.AddInt64(&r.metrics.errors, 1)
		r.metrics.lastError.Store(err.Error())
		return fmt.Errorf("cache ping error: %w", err)
	}
	return nil
}

// GetMetrics returns cache statistics
func (r *RedisCache) GetMetrics() CacheMetrics {
	hits := atomic.LoadInt64(&r.metrics.hits)
	misses := atomic.LoadInt64(&r.metrics.misses)
	total := hits + misses
	
	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}
	
	var lastError string
	if err := r.metrics.lastError.Load(); err != nil {
		lastError = err.(string)
	}
	
	return CacheMetrics{
		Hits:      hits,
		Misses:    misses,
		Sets:      atomic.LoadInt64(&r.metrics.sets),
		Deletes:   atomic.LoadInt64(&r.metrics.deletes),
		Errors:    atomic.LoadInt64(&r.metrics.errors),
		LastError: lastError,
		HitRate:   hitRate,
	}
}