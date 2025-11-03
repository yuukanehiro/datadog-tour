package context

import (
	"context"

	"github.com/kanehiroyuu/datadog-tour/internal/usecase/port"
	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	loggerKey      contextKey = "logger"
	repoLocatorKey contextKey = "repo_locator"
	interactorKey  contextKey = "interactor"
)

// RepoLocator holds all repositories
type RepoLocator struct {
	UserRepo  port.UserRepository
	CacheRepo port.CacheRepository
}

// RUser returns UserRepository
func (r *RepoLocator) RUser() port.UserRepository {
	return r.UserRepo
}

// RCache returns CacheRepository
func (r *RepoLocator) RCache() port.CacheRepository {
	return r.CacheRepo
}

// SetLogger sets logger in context
func SetLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger retrieves logger from context
func GetLogger(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(loggerKey).(*logrus.Logger); ok {
		return logger
	}
	return logrus.New() // fallback
}

// SetRepoLocator sets repository locator in context
func SetRepoLocator(ctx context.Context, locator *RepoLocator) context.Context {
	return context.WithValue(ctx, repoLocatorKey, locator)
}

// GetRepoLocator retrieves repository locator from context
func GetRepoLocator(ctx context.Context) *RepoLocator {
	if locator, ok := ctx.Value(repoLocatorKey).(*RepoLocator); ok {
		return locator
	}
	return nil
}

// SetInteractor sets interactor in context
func SetInteractor(ctx context.Context, interactor any) context.Context {
	return context.WithValue(ctx, interactorKey, interactor)
}

// GetInteractor retrieves interactor from context
func GetInteractor(ctx context.Context) any {
	return ctx.Value(interactorKey)
}
