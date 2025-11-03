package middleware

import (
	"net/http"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
)

// InteractorMiddleware sets interactor in context
func InteractorMiddleware(interactor any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := appcontext.SetInteractor(r.Context(), interactor)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
