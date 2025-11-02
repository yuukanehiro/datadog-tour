package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Response represents API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RespondJSONWithTrace sends a JSON response with trace headers
func RespondJSONWithTrace(ctx context.Context, w http.ResponseWriter, status int, data interface{}) {
	// Extract trace information from context
	span, ok := tracer.SpanFromContext(ctx)
	if ok {
		spanContext := span.Context()
		traceID := spanContext.TraceID()
		spanID := spanContext.SpanID()

		// Add trace headers to response
		w.Header().Set("X-Datadog-Trace-Id", fmt.Sprintf("%d", traceID))
		w.Header().Set("X-Datadog-Span-Id", fmt.Sprintf("%d", spanID))
		w.Header().Set("X-Datadog-Parent-Id", fmt.Sprintf("%d", spanID))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, Response{
		Success: false,
		Message: message,
	})
}

// RespondErrorWithTrace sends an error response with trace information
func RespondErrorWithTrace(ctx context.Context, w http.ResponseWriter, status int, message string) {
	RespondJSONWithTrace(ctx, w, status, Response{
		Success: false,
		Message: message,
	})
}

// RespondSuccess sends a success response
func RespondSuccess(w http.ResponseWriter, status int, data interface{}, message string) {
	RespondJSON(w, status, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// RespondSuccessWithTrace sends a success response with trace information
func RespondSuccessWithTrace(ctx context.Context, w http.ResponseWriter, status int, data interface{}, message string) {
	RespondJSONWithTrace(ctx, w, status, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}
