package router

import (
	"net/http"
	"os"

	gorillatrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"

	"github.com/kanehiroyuu/datadog-tour/internal/presentation/interface-adapter/handler"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/middleware"
)

// Setup configures all routes with Datadog tracing
func Setup(userHandler *handler.UserHandler, healthHandler *handler.HealthHandler, testHandler *handler.TestHandler) *gorillatrace.Router {
	// Setup router with tracing
	// ここでspanが作成され、以降のハンドラやミドルウェアで利用可能に
	router := gorillatrace.NewRouter(gorillatrace.WithServiceName(os.Getenv("DD_SERVICE")))

	// Apply recovery middleware INSIDE the router so span is available
	router.Use(func(next http.Handler) http.Handler {
		return middleware.RecoveryMiddleware()(next)
	})

	// Health endpoints
	router.HandleFunc("/", healthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")

	// User endpoints
	router.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/api/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")

	// Test endpoints for Datadog demonstration
	router.HandleFunc("/api/slow", testHandler.SlowEndpoint).Methods("GET")
	router.HandleFunc("/api/error", testHandler.ErrorEndpoint).Methods("GET")
	router.HandleFunc("/api/expected-error", testHandler.ExpectedErrorEndpoint).Methods("GET")
	router.HandleFunc("/api/unexpected-error", testHandler.UnexpectedErrorEndpoint).Methods("GET")
	router.HandleFunc("/api/warn", testHandler.WarnEndpoint).Methods("GET")
	router.HandleFunc("/api/panic", testHandler.PanicEndpoint).Methods("GET")

	return router
}
