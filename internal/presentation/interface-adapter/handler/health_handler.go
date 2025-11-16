package handler

import (
	"net/http"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/labstack/echo/v4"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(c echo.Context) error {
	span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "handler.health_check")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)
	repoLocator := appcontext.GetRepoLocator(ctx)

	// Add request metadata to span
	span.SetTag("http.method", c.Request().Method)
	span.SetTag("http.url", c.Request().URL.Path)
	span.SetTag("http.user_agent", c.Request().UserAgent())
	span.SetTag("health.status", "healthy")

	// Send custom metric: health check count
	repoLocator.StatsdClient.Incr("api.health.check", nil, 1)

	logging.LogWithTrace(ctx, logger, "handler", "Health check endpoint called", nil)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Service is healthy",
	})
}
