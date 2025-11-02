package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TestHandler handles test endpoints for Datadog demonstrations
type TestHandler struct {
	logger *logrus.Logger
}

// NewTestHandler creates a new TestHandler
func NewTestHandler(logger *logrus.Logger) *TestHandler {
	return &TestHandler{
		logger: logger,
	}
}

// SlowEndpoint handles GET /api/slow - demonstrates slow requests
func (h *TestHandler) SlowEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.slow_endpoint")
	defer span.Finish()

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "slow_request")

	h.logWithTrace(ctx, "Slow endpoint called - simulating 2 second delay")

	// Simulate slow database query
	span.SetTag("operation", "slow_query_simulation")
	time.Sleep(2 * time.Second)

	h.logWithTrace(ctx, "Slow operation completed")

	RespondSuccessWithTrace(ctx, w, http.StatusOK, map[string]interface{}{
		"message": "This endpoint intentionally took 2 seconds to respond",
		"delay":   "2s",
	}, "Slow request completed successfully")
}

// ErrorEndpoint handles GET /api/error - demonstrates error tracing
func (h *TestHandler) ErrorEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.error_endpoint")
	defer span.Finish()

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "error_simulation")

	h.logWithTrace(ctx, "Error endpoint called - will generate an error")

	// Simulate an error
	err := errors.New("simulated database connection error")

	// Mark span as error
	span.SetTag("error", true)
	span.SetTag("error.msg", err.Error())
	span.SetTag("error.type", "database_error")
	span.SetTag("error.stack", "user_repository.go:42")

	h.logErrorWithTrace(ctx, "Simulated error occurred", err)

	RespondErrorWithTrace(ctx, w, http.StatusInternalServerError, "An intentional error occurred for Datadog demonstration")
}

func (h *TestHandler) logWithTrace(ctx context.Context, message string) {
	logging.LogWithTrace(ctx, h.logger, "handler", message, nil)
}

func (h *TestHandler) logErrorWithTrace(ctx context.Context, message string, err error) {
	logging.LogErrorWithTrace(ctx, h.logger, "handler", message, err, nil)
}
