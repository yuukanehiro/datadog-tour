package handler

import (
	"net/http"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.health_check")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("http.user_agent", r.UserAgent())
	span.SetTag("health.status", "healthy")

	logging.LogWithTrace(ctx, logger, "handler", "Health check endpoint called", nil)

	RespondSuccessWithTrace(ctx, w, http.StatusOK, nil, "Service is healthy")
}
