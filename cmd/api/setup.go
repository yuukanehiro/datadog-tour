package main

import (
	"database/sql"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/database"
	infraredis "github.com/kanehiroyuu/datadog-tour/internal/infrastructure/redis"
	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/tracing"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/interface-adapter/handler"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/router"
)

// SetupRepositories creates and configures all repositories
func SetupRepositories(db *sql.DB, redisClient redis.UniversalClient, logger *slog.Logger) *appcontext.RepoLocator {
	// Setup repositories
	userRepo := database.NewUserRepository(db, logger)
	cacheRepoBase := infraredis.NewCacheRepository(redisClient)
	cacheRepo := tracing.NewCacheRepositoryTracer(cacheRepoBase, cacheRepoBase.GetTTL())

	// Setup RepoLocator
	return &appcontext.RepoLocator{
		UserRepo:  userRepo,
		CacheRepo: cacheRepo,
	}
}

// SetupRouter creates and configures the application router with all handlers
func SetupRouter(logger *slog.Logger, repoLocator *appcontext.RepoLocator) *echo.Echo {
	// Setup handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler()
	testHandler := handler.NewTestHandler()

	// Setup router with tracing
	return router.Setup(userHandler, healthHandler, testHandler, logger, repoLocator)
}
