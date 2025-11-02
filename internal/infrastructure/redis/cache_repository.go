package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheRepository implements cache operations for Redis (without tracing)
type CacheRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewCacheRepository creates a new CacheRepository
func NewCacheRepository(client redis.UniversalClient) *CacheRepository {
	return &CacheRepository{
		client: client,
		ttl:    5 * time.Minute,
	}
}

// GetTTL returns the configured TTL
func (r *CacheRepository) GetTTL() time.Duration {
	return r.ttl
}

// Set stores a value in cache
func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}) error {
	if err := r.client.Set(ctx, key, value, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

// Get retrieves a value from cache
func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}

	if err != nil {
		return "", fmt.Errorf("failed to get cache: %w", err)
	}

	return value, nil
}

// Delete removes a value from cache
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}
