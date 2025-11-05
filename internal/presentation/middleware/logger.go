package middleware

import (
	"github.com/labstack/echo/v4"
	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/sirupsen/logrus"
)

// EchoLoggerMiddleware sets logger in context for Echo
func EchoLoggerMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := appcontext.SetLogger(c.Request().Context(), logger)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
