package middleware

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// EchoCORSMiddleware creates a CORS middleware with Datadog tracing
func EchoCORSMiddleware() echo.MiddlewareFunc {
	// Use Echo's built-in CORS with default config
	corsHandler := echomiddleware.CORSWithConfig(echomiddleware.DefaultCORSConfig)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create span for this middleware
			span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "middleware.cors")
			defer span.Finish()

			// Update request context
			c.SetRequest(c.Request().WithContext(ctx))

			// Execute CORS logic
			return corsHandler(next)(c)
		}
	}
}
