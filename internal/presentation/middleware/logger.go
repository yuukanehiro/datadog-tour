package middleware

import (
	"net/http"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware sets logger in context
func LoggerMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := appcontext.SetLogger(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
