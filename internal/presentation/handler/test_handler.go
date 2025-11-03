package handler

import (
	"errors"
	"net/http"
	"time"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TestHandler handles test endpoints for Datadog demonstrations
type TestHandler struct{}

// NewTestHandler creates a new TestHandler
func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

// SlowEndpoint handles GET /api/slow - demonstrates slow requests
func (h *TestHandler) SlowEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.slow_endpoint")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "slow_request")

	logging.LogWithTrace(ctx, logger, "handler", "Slow endpoint called - simulating 2 second delay", nil)

	// Simulate slow database query
	span.SetTag("operation", "slow_query_simulation")
	time.Sleep(2 * time.Second)

	logging.LogWithTrace(ctx, logger, "handler", "Slow operation completed", nil)

	RespondSuccessWithTrace(ctx, w, http.StatusOK, map[string]interface{}{
		"message": "This endpoint intentionally took 2 seconds to respond",
		"delay":   "2s",
	}, "Slow request completed successfully")
}

// ErrorEndpoint handles GET /api/error - demonstrates error tracing
func (h *TestHandler) ErrorEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.error_endpoint")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "error_simulation")

	logging.LogWithTrace(ctx, logger, "handler", "Error endpoint called - will generate an error", nil)

	// Simulate an error
	err := errors.New("simulated database connection error")

	// Mark span as error
	span.SetTag("error", true)
	span.SetTag("error.msg", err.Error())
	span.SetTag("error.type", "database_error")
	span.SetTag("error.stack", "user_repository.go:42")

	logging.LogErrorWithTrace(ctx, logger, "handler", "Simulated error occurred", err, nil)

	RespondErrorWithTrace(ctx, w, http.StatusInternalServerError, "An intentional error occurred for Datadog demonstration")
}
