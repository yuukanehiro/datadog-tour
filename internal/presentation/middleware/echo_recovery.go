package middleware

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// EchoRecoveryMiddleware recovers from panics and logs them with trace information
func EchoRecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract span BEFORE any panic can occur
			var traceID, spanID string
			if span, ok := tracer.SpanFromContext(c.Request().Context()); ok {
				spanContext := span.Context()
				traceID = strconv.FormatUint(spanContext.TraceID(), 10)
				spanID = strconv.FormatUint(spanContext.SpanID(), 10)
			}

			defer func() {
				if err := recover(); err != nil {
					// Get logger from context
					logger := appcontext.GetLogger(c.Request().Context())
					if logger == nil {
						// Create fallback logger with JSON formatter
						logger = logrus.New()
					}

					// Ensure logger uses JSON formatter for Datadog log-trace correlation
					logger.SetFormatter(&logrus.JSONFormatter{})
					logger.SetOutput(os.Stdout)
					logger.SetLevel(logrus.InfoLevel)

					// Get stack trace
					stackTrace := string(debug.Stack())

					// Create error from panic
					panicErr := fmt.Errorf("panic recovered: %v", err)

					// Prepare log fields
					logFields := logrus.Fields{
						"panic.value":       err,
						"panic.stack_trace": stackTrace,
						"http.method":       c.Request().Method,
						"http.url":          c.Request().URL.Path,
					}

					// Use trace information extracted before panic
					if traceID != "" && spanID != "" {
						logFields["dd.trace_id"] = traceID
						logFields["dd.span_id"] = spanID
					}

					// Log with trace information
					logging.LogErrorWithTrace(c.Request().Context(), logger, "middleware", "Panic recovered", panicErr, logFields)

					// Set error tag on span
					if span, ok := tracer.SpanFromContext(c.Request().Context()); ok {
						span.SetTag("error", true)
						span.SetTag("error.type", "panic")
						span.SetTag("error.msg", fmt.Sprintf("%v", err))
						span.SetTag("error.stack", stackTrace)
						span.SetTag("error.notify", true)
					}

					// Return 500 Internal Server Error
					c.JSON(500, map[string]string{
						"error": "Internal Server Error",
					})
				}
			}()

			return next(c)
		}
	}
}
