package logging

import (
	"context"
	"fmt"
	"runtime"
	"time"

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
func LogSQL(ctx context.Context, logger logrus.FieldLogger, query string, args []interface{}, duration time.Duration, rowsAffected int64, err error) {
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

	fields := logrus.Fields{
		"component":       "sql",
		"sql.query":       query,
		"sql.args":        args,
		"sql.duration_ms": durationMsFloat,
	}

	if rowsAffected >= 0 {
		fields["sql.rows_affected"] = rowsAffected
	}

	// Add trace information
	if span, ok := tracer.SpanFromContext(ctx); ok {
		spanContext := span.Context()
		fields["dd.trace_id"] = spanContext.TraceID()
		fields["dd.span_id"] = spanContext.SpanID()
	}

	if err != nil {
		fields["sql.error"] = err.Error()
		logger.WithFields(fields).WithError(err).Error(message)
	} else {
		logger.WithFields(fields).Info(message)
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
