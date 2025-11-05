package logging

import (
	"context"
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// prepareLogFields prepares common log fields with caller info and trace information
func prepareLogFields(ctx context.Context, layer string, fields logrus.Fields, skipFrames int) (logrus.Fields, string) {
	if fields == nil {
		fields = logrus.Fields{}
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(skipFrames)
	var callerInfo string
	if ok {
		fields["file"] = file
		fields["line"] = line
		callerInfo = fmt.Sprintf("%s:%d", file, line)
	}

	// Extract trace information from context
	if span, ok := tracer.SpanFromContext(ctx); ok {
		spanContext := span.Context()
		fields["dd.trace_id"] = spanContext.TraceID()
		fields["dd.span_id"] = spanContext.SpanID()
	}

	fields["layer"] = layer

	return fields, callerInfo
}

// formatLogMessage formats a log message with layer and caller information
func formatLogMessage(layer, callerInfo, message string) string {
	if callerInfo != "" {
		return fmt.Sprintf("[%s] %s | %s", layer, callerInfo, message)
	}
	return fmt.Sprintf("[%s] %s", layer, message)
}

// LogWithTrace logs an informational message with trace information
// Level: INFO
func LogWithTrace(ctx context.Context, logger logrus.FieldLogger, layer, message string, fields logrus.Fields) {
	fields, callerInfo := prepareLogFields(ctx, layer, fields, 2)
	formattedMessage := formatLogMessage(layer, callerInfo, message)
	logger.WithFields(fields).Info(formattedMessage)
}

// LogWarnWithTrace logs a warning with trace information
// Level: WARN
// Use for: Performance warnings, deprecated features, non-critical issues
func LogWarnWithTrace(ctx context.Context, logger logrus.FieldLogger, layer, message string, fields logrus.Fields) {
	fields, callerInfo := prepareLogFields(ctx, layer, fields, 2)
	formattedMessage := formatLogMessage(layer, callerInfo, message)
	logger.WithFields(fields).Warn(formattedMessage)
}

// LogErrorWithTrace logs a system error with trace information
// Level: ERROR (triggers alerts)
// Use for: Database errors, external API failures, system-level errors
//
// Example:
//
//	LogErrorWithTrace(ctx, logger, "usecase", "Database connection failed", err, nil)
func LogErrorWithTrace(ctx context.Context, logger logrus.FieldLogger, layer, message string, err error, fields logrus.Fields) {
	fields, callerInfo := prepareLogFields(ctx, layer, fields, 2)

	// Add error.notify field to logs
	fields["error.notify"] = true

	if span, ok := tracer.SpanFromContext(ctx); ok {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("error.notify", true)

		if errorType, ok := fields["error.type"]; ok {
			span.SetTag("error.type", errorType)
		} else {
			span.SetTag("error.type", "system_error")
		}
	}

	formattedMessage := formatLogMessage(layer, callerInfo, message)
	logger.WithFields(fields).WithError(err).Error(formattedMessage)
}

// LogErrorWithTraceNotNotify logs an error that should not trigger alerts
// Level: ERROR (does NOT trigger alerts due to error.notify=false)
// Use for: Validation errors, duplicate entries, not found, permission denied
//
// Example:
//
//	LogErrorWithTraceNotNotify(ctx, logger, "usecase", "User already exists", err, logrus.Fields{
//	    "error.type": "validation_error",
//	})
func LogErrorWithTraceNotNotify(ctx context.Context, logger logrus.FieldLogger, layer, message string, err error, fields logrus.Fields) {
	fields, callerInfo := prepareLogFields(ctx, layer, fields, 2)

	// Add error.notify field to logs
	fields["error.notify"] = false

	if span, ok := tracer.SpanFromContext(ctx); ok {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("error.notify", false)

		if errorType, ok := fields["error.type"]; ok {
			span.SetTag("error.type", errorType)
		} else {
			span.SetTag("error.type", "expected_error")
		}
	}

	formattedMessage := formatLogMessage(layer, callerInfo, message)
	logger.WithFields(fields).WithError(err).Error(formattedMessage)
}
