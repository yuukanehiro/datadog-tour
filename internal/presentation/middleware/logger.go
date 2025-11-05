package middleware

import (
	"github.com/labstack/echo/v4"
	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// EchoLoggerMiddleware sets logger in context for Echo
func EchoLoggerMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create span for this middleware
			span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "middleware.logger")
			defer span.Finish()

			// Set logger in context
			ctx = appcontext.SetLogger(ctx, logger)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
