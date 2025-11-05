package router

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	echotrace "github.com/DataDog/dd-trace-go/contrib/labstack/echo.v4/v2"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/interface-adapter/handler"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/middleware"
)

// Setup configures all routes with Datadog tracing
func Setup(userHandler *handler.UserHandler, healthHandler *handler.HealthHandler, testHandler *handler.TestHandler, logger interface{}, repoLocator interface{}) *echo.Echo {
	// Setup Echo with Datadog tracing
	// ここでspanが作成され、以降のハンドラやミドルウェアで利用可能に
	e := echo.New()

	// Disable Echo's default logger since we use our own
	e.HideBanner = true
	e.HidePort = true

	// Apply middlewares in order
	// 1. Logger middleware - sets logger in context
	if logger != nil {
		e.Use(middleware.EchoLoggerMiddleware(logger.(*logrus.Logger)))
	}

	// 2. RepoLocator middleware - sets repository locator in context
	if repoLocator != nil {
		e.Use(middleware.EchoRepoLocatorMiddleware(repoLocator.(*appcontext.RepoLocator)))
	}

	// 3. Datadog tracing middleware
	e.Use(echotrace.Middleware(echotrace.WithService(os.Getenv("DD_SERVICE"))))

	// 4. Recovery middleware AFTER tracing so span is available
	e.Use(middleware.EchoRecoveryMiddleware())

	// 5. CORS middleware with Datadog tracing
	e.Use(middleware.EchoCORSMiddleware())

	// Health endpoints
	e.GET("/", healthHandler.HealthCheck)
	e.GET("/health", healthHandler.HealthCheck)

	// User endpoints
	e.POST("/api/users", userHandler.CreateUser)
	e.GET("/api/users", userHandler.GetAllUsers)
	e.GET("/api/users/:id", userHandler.GetUser)

	// Test endpoints for Datadog demonstration
	e.GET("/api/slow", testHandler.SlowEndpoint)
	e.GET("/api/error", testHandler.ErrorEndpoint)
	e.GET("/api/expected-error", testHandler.ExpectedErrorEndpoint)
	e.GET("/api/unexpected-error", testHandler.UnexpectedErrorEndpoint)
	e.GET("/api/warn", testHandler.WarnEndpoint)
	e.GET("/api/panic", testHandler.PanicEndpoint)

	return e
}
