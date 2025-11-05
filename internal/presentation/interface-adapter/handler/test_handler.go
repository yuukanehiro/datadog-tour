package handler

import (
	"errors"
	"net/http"
	"time"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/interface-adapter/response"
	"github.com/kanehiroyuu/datadog-tour/internal/usecase"
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

	response.RespondSuccessWithTrace(ctx, w, http.StatusOK, map[string]any{
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

	problem := response.NewInternalErrorProblem(
		"Simulated database connection error for Datadog demonstration",
		r.URL.Path,
		true,
	)
	problem.Extra["error.stack"] = "user_repository.go:42"
	problem.Extra["db.operation"] = "connection_test"
	response.RespondProblemWithTrace(ctx, w, problem)
}

// ExpectedErrorEndpoint handles GET /api/expected-error - demonstrates expected error (no alert)
func (h *TestHandler) ExpectedErrorEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.expected_error_endpoint")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "expected_error_simulation")

	logging.LogWithTrace(ctx, logger, "handler", "Expected error endpoint called", nil)

	// Simulate an expected error (validation, duplicate, etc.)
	err := errors.New("user already exists")

	// Use LogErrorWithTraceNotNotify for expected errors that shouldn't trigger alerts
	logging.LogErrorWithTraceNotNotify(ctx, logger, "handler", "Expected error occurred", err, map[string]any{
		"error.type": "validation_error",
		"user.email": "duplicate@example.com",
	})

	problem := response.NewConflictProblem(
		"User with email 'duplicate@example.com' already exists. This is an expected error that should not trigger alerts.",
		r.URL.Path,
	)
	problem.Extra["user.email"] = "duplicate@example.com"
	problem.Extra["validation.field"] = "email"
	response.RespondProblemWithTrace(ctx, w, problem)
}

// UnexpectedErrorEndpoint handles GET /api/unexpected-error - demonstrates unexpected error (should alert)
func (h *TestHandler) UnexpectedErrorEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.unexpected_error_endpoint")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "unexpected_error_simulation")

	logging.LogWithTrace(ctx, logger, "handler", "Unexpected error endpoint called", nil)

	// Simulate an unexpected error (system error, should trigger alerts)
	err := errors.New("database connection lost")

	// Normal error logging - WILL trigger alerts
	logging.LogErrorWithTrace(ctx, logger, "handler", "Unexpected error occurred", err, map[string]any{
		"error.type": "system_error",
		"db.host":    "mysql.example.com",
	})

	problem := response.NewInternalErrorProblem(
		"Database connection to mysql.example.com was lost unexpectedly. This system error should trigger an alert for immediate investigation.",
		r.URL.Path,
		true,
	)
	problem.Extra["db.host"] = "mysql.example.com"
	problem.Extra["db.port"] = 3306
	problem.Extra["retry.attempted"] = false
	response.RespondProblemWithTrace(ctx, w, problem)
}

// WarnEndpoint handles GET /api/warn - demonstrates warning logs
func (h *TestHandler) WarnEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.warn_endpoint")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "warning_simulation")

	logging.LogWithTrace(ctx, logger, "handler", "Warn endpoint called", nil)

	// Simulate warning scenarios
	logging.LogWarnWithTrace(ctx, logger, "handler", "Performance degradation detected", map[string]any{
		"warn.type":        "performance",
		"response_time_ms": 1500,
		"threshold_ms":     1000,
	})

	response.RespondSuccessWithTrace(ctx, w, http.StatusOK, map[string]any{
		"message": "Warning logged successfully",
		"level":   "warn",
		"type":    "performance_degradation",
	}, "Warning endpoint completed")
}

// PanicEndpoint handles GET /api/panic - demonstrates panic recovery and trace logging
func (h *TestHandler) PanicEndpoint(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.panic_endpoint")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)
	repoLocator := appcontext.GetRepoLocator(ctx)

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("test.type", "panic_simulation")

	logging.LogWithTrace(ctx, logger, "handler", "Panic endpoint called - will trigger panic in repository layer", nil)

	// Create usecase interactor
	interactor := &usecase.UserUseCase{
		Logger: logger,
		RUser:  repoLocator.UserRepo,
		RCache: repoLocator.CacheRepo,
	}

	// Call usecase method that will trigger panic in repository
	// This panic should be caught by RecoveryMiddleware and logged with trace information
	err := interactor.TestPanic(ctx)
	if err != nil {
		logging.LogErrorWithTrace(ctx, logger, "handler", "Test panic returned error", err, nil)
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		problem := response.NewInternalErrorProblem(
			"Test panic failed",
			r.URL.Path,
			true,
		)
		problem.Extra["error"] = err.Error()
		response.RespondProblemWithTrace(ctx, w, problem)
		return
	}

	// This line should never be reached due to panic
	response.RespondSuccessWithTrace(ctx, w, http.StatusOK, map[string]any{
		"message": "This should never be returned",
	}, "Panic test completed")
}
