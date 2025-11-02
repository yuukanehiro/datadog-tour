package tracing

import (
	"context"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// SpanFunc is a function that executes within a traced span
type SpanFunc func(ctx context.Context, span tracer.Span) error

// TraceOperation creates a span and executes the given function
func TraceOperation(ctx context.Context, operationName string, tags map[string]interface{}, fn SpanFunc) error {
	span, ctx := tracer.StartSpanFromContext(ctx, operationName)
	defer span.Finish()

	// Set all provided tags
	for key, value := range tags {
		span.SetTag(key, value)
	}

	// Execute the function
	err := fn(ctx, span)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
	}

	return err
}

// AddSpanTags adds multiple tags to a span
func AddSpanTags(span tracer.Span, tags map[string]interface{}) {
	for key, value := range tags {
		span.SetTag(key, value)
	}
}

// AddSpanError adds error information to a span
func AddSpanError(span tracer.Span, err error) {
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
	}
}

// AddSpanSuccess adds success information to a span
func AddSpanSuccess(span tracer.Span, tags map[string]interface{}) {
	span.SetTag("success", true)
	for key, value := range tags {
		span.SetTag(key, value)
	}
}
