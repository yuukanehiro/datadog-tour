package logging

import (
	"context"
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// LogWithTrace logs a message with trace information and caller details
func LogWithTrace(ctx context.Context, logger *logrus.Logger, layer, message string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}

	// Get caller information (skip 2 frames to get the actual caller)
	_, file, line, ok := runtime.Caller(2)
	var formattedMessage string
	if ok {
		fields["file"] = file
		fields["line"] = line
		// Format message with layer, file, line, and message
		formattedMessage = fmt.Sprintf("[%s] %s:%d | %s", layer, file, line, message)
	} else {
		// Fallback if caller info is not available
		formattedMessage = fmt.Sprintf("[%s] %s", layer, message)
	}

	// Extract trace information from context
	if span, ok := tracer.SpanFromContext(ctx); ok {
		spanContext := span.Context()
		fields["dd.trace_id"] = spanContext.TraceID()
		fields["dd.span_id"] = spanContext.SpanID()
	}

	fields["layer"] = layer

	logger.WithFields(fields).Info(formattedMessage)
}

// LogErrorWithTrace logs an error with trace information and caller details
func LogErrorWithTrace(ctx context.Context, logger *logrus.Logger, layer, message string, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}

	// Get caller information (skip 2 frames to get the actual caller)
	_, file, line, ok := runtime.Caller(2)
	var formattedMessage string
	if ok {
		fields["file"] = file
		fields["line"] = line
		// Format message with layer, file, line, and message
		formattedMessage = fmt.Sprintf("[%s] %s:%d | %s", layer, file, line, message)
	} else {
		// Fallback if caller info is not available
		formattedMessage = fmt.Sprintf("[%s] %s", layer, message)
	}

	// Extract trace information from context
	if span, ok := tracer.SpanFromContext(ctx); ok {
		spanContext := span.Context()
		fields["dd.trace_id"] = spanContext.TraceID()
		fields["dd.span_id"] = spanContext.SpanID()
	}

	fields["layer"] = layer

	logger.WithFields(fields).WithError(err).Error(formattedMessage)
}
