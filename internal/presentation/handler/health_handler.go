package handler

import (
	"context"
	"net/http"

	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger *logrus.Logger
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.health_check")
	defer span.Finish()

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("http.user_agent", r.UserAgent())
	span.SetTag("health.status", "healthy")

	h.logWithTrace(ctx, "Health check endpoint called")

	RespondSuccessWithTrace(ctx, w, http.StatusOK, nil, "Service is healthy")
}

func (h *HealthHandler) logWithTrace(ctx context.Context, message string) {
	logging.LogWithTrace(ctx, h.logger, "handler", message, nil)
}
