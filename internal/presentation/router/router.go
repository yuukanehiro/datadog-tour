package router

import (
	"os"

	gorillatrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"

	"github.com/kanehiroyuu/datadog-tour/internal/presentation/handler"
)

// Setup configures all routes with Datadog tracing
func Setup(userHandler *handler.UserHandler, healthHandler *handler.HealthHandler, testHandler *handler.TestHandler) *gorillatrace.Router {
	// Setup router with tracing
	router := gorillatrace.NewRouter(gorillatrace.WithServiceName(os.Getenv("DD_SERVICE")))

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

	return router
}
