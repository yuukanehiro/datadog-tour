package middleware

import (
	"net/http"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
)

// RepoLocatorMiddleware sets repository locator in context
func RepoLocatorMiddleware(locator *appcontext.RepoLocator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := appcontext.SetRepoLocator(r.Context(), locator)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
