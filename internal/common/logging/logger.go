package logging

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// prepareLogAttrs prepares common log attributes with caller info and trace information
func prepareLogAttrs(ctx context.Context, layer string, fields map[string]any, skipFrames int) []any {
	attrs := make([]any, 0, 20)

	// Get caller information
	_, file, line, ok := runtime.Caller(skipFrames)
	if ok {
		attrs = append(attrs, "file", file, "line", line)
	}

	// Extract trace information from context
	// Convert to string format for Datadog log-trace correlation
	if span, ok := tracer.SpanFromContext(ctx); ok {
		spanContext := span.Context()
		attrs = append(attrs,
			"dd.trace_id", strconv.FormatUint(spanContext.TraceID(), 10),
			"dd.span_id", strconv.FormatUint(spanContext.SpanID(), 10),
		)
	}

	attrs = append(attrs, "layer", layer)

	// Add custom fields
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}

	return attrs
}

// formatLogMessage formats a log message with layer
func formatLogMessage(layer, message string) string {
	return fmt.Sprintf("[%s] %s", layer, message)
}

// LogWithTrace logs an informational message with trace information
// Level: INFO
func LogWithTrace(ctx context.Context, logger *slog.Logger, layer, message string, fields map[string]any) {
	attrs := prepareLogAttrs(ctx, layer, fields, 2)
	formattedMessage := formatLogMessage(layer, message)
	logger.InfoContext(ctx, formattedMessage, attrs...)
}

// LogWarnWithTrace logs a warning with trace information
// Level: WARN
// Use for: Performance warnings, deprecated features, non-critical issues
func LogWarnWithTrace(ctx context.Context, logger *slog.Logger, layer, message string, fields map[string]any) {
	attrs := prepareLogAttrs(ctx, layer, fields, 2)
	formattedMessage := formatLogMessage(layer, message)
	logger.WarnContext(ctx, formattedMessage, attrs...)
}

// LogErrorWithTrace logs a system error with trace information
// Level: ERROR (triggers alerts)
// Use for: Database errors, external API failures, system-level errors
//
// Example:
//
//	LogErrorWithTrace(ctx, logger, "usecase", "Database connection failed", err, nil)
func LogErrorWithTrace(ctx context.Context, logger *slog.Logger, layer, message string, err error, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}

	// Add error.notify field to logs
	fields["error.notify"] = true
	fields["error"] = err.Error()

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

	attrs := prepareLogAttrs(ctx, layer, fields, 2)
	formattedMessage := formatLogMessage(layer, message)
	logger.ErrorContext(ctx, formattedMessage, attrs...)
}

// LogErrorWithTraceNotNotify logs an error that should not trigger alerts
// Level: ERROR (does NOT trigger alerts due to error.notify=false)
// Use for: Validation errors, duplicate entries, not found, permission denied
//
// Example:
//
//	LogErrorWithTraceNotNotify(ctx, logger, "usecase", "User already exists", err, map[string]any{
//	    "error.type": "validation_error",
//	})
func LogErrorWithTraceNotNotify(ctx context.Context, logger *slog.Logger, layer, message string, err error, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}

	// Add error.notify field to logs
	fields["error.notify"] = false
	fields["error"] = err.Error()

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

	attrs := prepareLogAttrs(ctx, layer, fields, 2)
	formattedMessage := formatLogMessage(layer, message)
	logger.ErrorContext(ctx, formattedMessage, attrs...)
}

// LogSQL logs SQL execution in GORM format with actual parameter values
// Format: [timestamp]  [duration]  sql with values  [rows affected]
//
// Example:
//
//	LogSQL(ctx, logger, "SELECT * FROM users WHERE id = ?", []interface{}{123}, time.Millisecond*10, 1, nil)
//	// Output: [2024-11-05 15:04:05]  [10.00ms]  SELECT * FROM users WHERE id = 123  [1 rows]
//
//	LogSQL(ctx, logger, "INSERT INTO users (name, email) VALUES (?, ?)", []interface{}{"John", "john@example.com"}, time.Millisecond*5, 1, nil)
//	// Output: [2024-11-05 15:04:05]  [5.00ms]  INSERT INTO users (name, email) VALUES ('John', 'john@example.com')  [1 rows]
func LogSQL(ctx context.Context, logger *slog.Logger, query string, args []interface{}, duration time.Duration, rowsAffected int64, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	durationMs := fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000.0)

	// Replace placeholders with actual values
	formattedQuery := formatSQLWithArgs(query, args)

	var rowsStr string
	if rowsAffected >= 0 {
		rowsStr = fmt.Sprintf("  [%d rows]", rowsAffected)
	} else {
		rowsStr = ""
	}

	message := fmt.Sprintf("[%s]  [%s]  %s%s", timestamp, durationMs, formattedQuery, rowsStr)

	// Convert duration to float64 milliseconds for accurate sub-millisecond values
	durationMsFloat := float64(duration.Microseconds()) / 1000.0

	attrs := []any{
		"component", "sql",
		"sql.query", query,
		"sql.args", args,
		"sql.duration_ms", durationMsFloat,
	}

	if rowsAffected >= 0 {
		attrs = append(attrs, "sql.rows_affected", rowsAffected)
	}

	// Add trace information
	// Convert to string format for Datadog log-trace correlation
	if span, ok := tracer.SpanFromContext(ctx); ok {
		spanContext := span.Context()
		attrs = append(attrs,
			"dd.trace_id", strconv.FormatUint(spanContext.TraceID(), 10),
			"dd.span_id", strconv.FormatUint(spanContext.SpanID(), 10),
		)
	}

	if err != nil {
		attrs = append(attrs, "sql.error", err.Error(), "error", err.Error())
		logger.ErrorContext(ctx, message, attrs...)
	} else {
		logger.InfoContext(ctx, message, attrs...)
	}
}

// formatSQLWithArgs replaces SQL placeholders with actual values for logging
func formatSQLWithArgs(query string, args []interface{}) string {
	if len(args) == 0 {
		return query
	}

	result := query
	for _, arg := range args {
		var value string
		switch v := arg.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case time.Time:
			value = fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
		case nil:
			value = "NULL"
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			value = fmt.Sprintf("%v", v)
		case float32, float64:
			value = fmt.Sprintf("%v", v)
		case bool:
			if v {
				value = "TRUE"
			} else {
				value = "FALSE"
			}
		default:
			value = fmt.Sprintf("'%v'", v)
		}

		// Replace first occurrence of ?
		for i := 0; i < len(result); i++ {
			if result[i] == '?' {
				result = result[:i] + value + result[i+1:]
				break
			}
		}
	}

	return result
}
