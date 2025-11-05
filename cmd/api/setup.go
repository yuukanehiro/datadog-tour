package main

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/database"
	infraredis "github.com/kanehiroyuu/datadog-tour/internal/infrastructure/redis"
	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/tracing"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/interface-adapter/handler"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/router"
)

// SetupRepositories creates and configures all repositories
func SetupRepositories(db *sql.DB, redisClient redis.UniversalClient, logger *logrus.Logger) *appcontext.RepoLocator {
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
func SetupRouter(logger *logrus.Logger, repoLocator *appcontext.RepoLocator) *echo.Echo {
	// Setup handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler()
	testHandler := handler.NewTestHandler()

	// Setup router with tracing
	return router.Setup(userHandler, healthHandler, testHandler, logger, repoLocator)
}
