package database

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
)

// LoggingDB wraps sql.DB to automatically log SQL queries in GORM format
type LoggingDB struct {
	*sql.DB
	logger *slog.Logger
}

// NewLoggingDB creates a new LoggingDB wrapper
func NewLoggingDB(db *sql.DB, logger *slog.Logger) *LoggingDB {
	return &LoggingDB{
		DB:     db,
		logger: logger,
	}
}

// ExecContext wraps sql.DB.ExecContext with automatic logging
func (db *LoggingDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	startTime := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(startTime)

	var rowsAffected int64 = -1
	if err == nil && result != nil {
		rowsAffected, _ = result.RowsAffected()
	}

	logging.LogSQL(ctx, db.logger, query, args, duration, rowsAffected, err)

	return result, err
}

// QueryContext wraps sql.DB.QueryContext with automatic logging
func (db *LoggingDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(startTime)

	// For SELECT queries, we don't know rows count until scanning
	logging.LogSQL(ctx, db.logger, query, args, duration, -1, err)

	return rows, err
}

// QueryRowContext wraps sql.DB.QueryRowContext with automatic logging
func (db *LoggingDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	startTime := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(startTime)

	// For QueryRow, we log without error check (error is checked on Scan)
	logging.LogSQL(ctx, db.logger, query, args, duration, -1, nil)

	return row
}
