package main

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/mysql"
	infraredis "github.com/kanehiroyuu/datadog-tour/internal/infrastructure/redis"
	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/tracing"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/handler"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/router"
	gorillatrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

// SetupRepositories creates and configures all repositories
func SetupRepositories(db *sql.DB, redisClient redis.UniversalClient, logger *logrus.Logger) *appcontext.RepoLocator {
	// Setup repositories
	userRepo := mysql.NewUserRepository(db, logger)
	cacheRepoBase := infraredis.NewCacheRepository(redisClient)
	cacheRepo := tracing.NewCacheRepositoryTracer(cacheRepoBase, cacheRepoBase.GetTTL())

	// Setup RepoLocator
	return &appcontext.RepoLocator{
		UserRepo:  userRepo,
		CacheRepo: cacheRepo,
	}
}

// SetupRouter creates and configures the application router with all handlers
func SetupRouter() *gorillatrace.Router {
	// Setup handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler()
	testHandler := handler.NewTestHandler()

	// Setup router with tracing
	return router.Setup(userHandler, healthHandler, testHandler)
}
