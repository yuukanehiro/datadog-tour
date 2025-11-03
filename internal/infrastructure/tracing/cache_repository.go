package tracing

import (
	"context"
	"time"

	"github.com/kanehiroyuu/datadog-tour/internal/usecase/port"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// CacheRepositoryTracer wraps a CacheRepository with tracing
type CacheRepositoryTracer struct {
	repo port.CacheRepository
	ttl  time.Duration
}

// NewCacheRepositoryTracer creates a new tracing decorator for CacheRepository
func NewCacheRepositoryTracer(repo port.CacheRepository, ttl time.Duration) port.CacheRepository {
	return &CacheRepositoryTracer{
		repo: repo,
		ttl:  ttl,
	}
}

// Set wraps the Set method with tracing
func (r *CacheRepositoryTracer) Set(ctx context.Context, key string, value interface{}) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "redis.set")
	defer span.Finish()

	// Add metadata
	span.SetTag("db.type", "redis")
	span.SetTag("db.operation", "SET")
	span.SetTag("cache.key", key)
	span.SetTag("cache.ttl", r.ttl.Seconds())

	err := r.repo.Set(ctx, key, value)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("cache.success", false)
		return err
	}

	span.SetTag("cache.success", true)
	return nil
}

// Get wraps the Get method with tracing
func (r *CacheRepositoryTracer) Get(ctx context.Context, key string) (string, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "redis.get")
	defer span.Finish()

	// Add metadata
	span.SetTag("db.type", "redis")
	span.SetTag("db.operation", "GET")
	span.SetTag("cache.key", key)

	value, err := r.repo.Get(ctx, key)
	if err != nil {
		// Cache miss is not an error
		span.SetTag("cache.hit", false)
		span.SetTag("cache.success", true)
		return "", err
	}

	span.SetTag("cache.hit", true)
	span.SetTag("cache.success", true)
	return value, nil
}

// Delete wraps the Delete method with tracing
func (r *CacheRepositoryTracer) Delete(ctx context.Context, key string) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "redis.delete")
	defer span.Finish()

	// Add metadata
	span.SetTag("db.type", "redis")
	span.SetTag("db.operation", "DELETE")
	span.SetTag("cache.key", key)

	err := r.repo.Delete(ctx, key)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("cache.success", false)
		return err
	}

	span.SetTag("cache.success", true)
	return nil
}
