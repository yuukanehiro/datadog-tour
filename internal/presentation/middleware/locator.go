package middleware

import (
	"github.com/labstack/echo/v4"
	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
)

// EchoRepoLocatorMiddleware sets repository locator in context for Echo
func EchoRepoLocatorMiddleware(locator *appcontext.RepoLocator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := appcontext.SetRepoLocator(c.Request().Context(), locator)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
