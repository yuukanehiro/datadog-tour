package response

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

// ProblemDetail represents RFC 9457 Problem Details for HTTP APIs
// See: https://www.rfc-editor.org/rfc/rfc9457.html
type ProblemDetail struct {
	Type     string                 `json:"type"`               // URI reference identifying the problem type
	Title    string                 `json:"title"`              // Short, human-readable summary
	Status   int                    `json:"status"`             // HTTP status code
	Detail   string                 `json:"detail"`             // Human-readable explanation
	Instance string                 `json:"instance"`           // URI reference identifying the specific occurrence
	TraceID  string                 `json:"trace_id,omitempty"` // Datadog trace ID for correlation
	SpanID   string                 `json:"span_id,omitempty"`  // Datadog span ID for correlation
	Notify   *bool                  `json:"notify,omitempty"`   // Whether this error should trigger alerts
	Extra    map[string]interface{} `json:"-"`                  // Additional extension members
}

// ErrorType defines standard error type URIs
const (
	ErrorTypeValidation     = "https://datadog-tour.example.com/errors/validation"
	ErrorTypeNotFound       = "https://datadog-tour.example.com/errors/not-found"
	ErrorTypeConflict       = "https://datadog-tour.example.com/errors/conflict"
	ErrorTypeUnauthorized   = "https://datadog-tour.example.com/errors/unauthorized"
	ErrorTypeForbidden      = "https://datadog-tour.example.com/errors/forbidden"
	ErrorTypeInternal       = "https://datadog-tour.example.com/errors/internal"
	ErrorTypeBadRequest     = "https://datadog-tour.example.com/errors/bad-request"
	ErrorTypeServiceUnavail = "https://datadog-tour.example.com/errors/service-unavailable"
)

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

// RespondSuccessWithTrace sends a success response with trace information
func RespondSuccessWithTrace(ctx context.Context, w http.ResponseWriter, status int, data interface{}, message string) {
	RespondJSONWithTrace(ctx, w, status, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// RespondProblemWithTrace sends an RFC 9457 Problem Details response with trace information
func RespondProblemWithTrace(ctx context.Context, w http.ResponseWriter, problem ProblemDetail) {
	// Extract trace information from context
	span, ok := tracer.SpanFromContext(ctx)
	if ok {
		spanContext := span.Context()
		traceID := spanContext.TraceID()
		spanID := spanContext.SpanID()

		// Add trace IDs to problem detail
		problem.TraceID = fmt.Sprintf("%d", traceID)
		problem.SpanID = fmt.Sprintf("%d", spanID)

		// Add trace headers to response
		w.Header().Set("X-Datadog-Trace-Id", fmt.Sprintf("%d", traceID))
		w.Header().Set("X-Datadog-Span-Id", fmt.Sprintf("%d", spanID))
		w.Header().Set("X-Datadog-Parent-Id", fmt.Sprintf("%d", spanID))
	}

	// Set Content-Type as per RFC 9457
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)

	// Encode problem detail with extra fields
	encoder := json.NewEncoder(w)

	// Marshal main struct
	mainData, _ := json.Marshal(problem)
	var result map[string]interface{}
	json.Unmarshal(mainData, &result)

	// Add extra fields to root level
	for k, v := range problem.Extra {
		result[k] = v
	}

	encoder.Encode(result)
}

// NewProblemDetail creates a new ProblemDetail with common fields set
func NewProblemDetail(errorType, title string, status int, detail, instance string) ProblemDetail {
	return ProblemDetail{
		Type:     errorType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Extra:    make(map[string]interface{}),
	}
}

// NewInternalErrorProblem creates a problem detail for internal server errors
func NewInternalErrorProblem(detail, instance string, notify bool) ProblemDetail {
	problem := NewProblemDetail(
		ErrorTypeInternal,
		"Internal Server Error",
		http.StatusInternalServerError,
		detail,
		instance,
	)
	problem.Notify = &notify
	return problem
}

// NewValidationErrorProblem creates a problem detail for validation errors
func NewValidationErrorProblem(detail, instance string) ProblemDetail {
	notifyFalse := false
	problem := NewProblemDetail(
		ErrorTypeValidation,
		"Validation Error",
		http.StatusBadRequest,
		detail,
		instance,
	)
	problem.Notify = &notifyFalse
	return problem
}

// NewNotFoundProblem creates a problem detail for not found errors
func NewNotFoundProblem(detail, instance string) ProblemDetail {
	notifyFalse := false
	problem := NewProblemDetail(
		ErrorTypeNotFound,
		"Not Found",
		http.StatusNotFound,
		detail,
		instance,
	)
	problem.Notify = &notifyFalse
	return problem
}

// NewConflictProblem creates a problem detail for conflict errors
func NewConflictProblem(detail, instance string) ProblemDetail {
	notifyFalse := false
	problem := NewProblemDetail(
		ErrorTypeConflict,
		"Conflict",
		http.StatusConflict,
		detail,
		instance,
	)
	problem.Notify = &notifyFalse
	return problem
}
