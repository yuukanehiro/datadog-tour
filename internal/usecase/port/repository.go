package port

import (
	"context"
	"log/slog"

	"github.com/kanehiroyuu/datadog-tour/internal/domain/entities"
)

// UserRepository is a port for user repository
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id int) (*entities.User, error)
	FindAll(ctx context.Context) ([]*entities.User, error)
	TestPanic(ctx context.Context) error // For testing panic recovery
}

// CacheRepository is a port for cache repository
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

// Logger is a port for logger
type Logger = *slog.Logger
