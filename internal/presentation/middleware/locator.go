package middleware

import (
	"github.com/labstack/echo/v4"
	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// EchoRepoLocatorMiddleware sets repository locator in context for Echo
func EchoRepoLocatorMiddleware(locator *appcontext.RepoLocator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create span for this middleware
			span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "middleware.repo_locator")
			defer span.Finish()

			// Set repository locator in context
			ctx = appcontext.SetRepoLocator(ctx, locator)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
