package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/kanehiroyuu/datadog-tour/internal/domain/entities"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// UserRepository implements entities.UserRepository for MySQL (without tracing)
type UserRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB, logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.create_user")
	defer span.Finish()

	query := "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)"

	r.logWithTrace(ctx, fmt.Sprintf("SQL: %s [params: name=%s, email=%s, created_at=%v]", query, user.Name, user.Email, user.CreatedAt), logrus.Fields{
		"query":           query,
		"user.name":       user.Name,
		"user.email":      user.Email,
		"user.created_at": user.CreatedAt,
	})

	startTime := time.Now()
	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.CreatedAt)
	duration := time.Since(startTime)
	if err != nil {
		r.logErrorWithTrace(ctx, fmt.Sprintf("SQL Error: %s [params: name=%s, email=%s] [duration: %v]", query, user.Name, user.Email, duration), err, logrus.Fields{
			"query":           query,
			"user.name":       user.Name,
			"user.email":      user.Email,
			"sql.duration_ms": duration.Milliseconds(),
		})
		return fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logErrorWithTrace(ctx, "Failed to get last insert ID", err, nil)
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = int(id)

	r.logWithTrace(ctx, fmt.Sprintf("SQL result: User created with ID=%d [duration: %v]", user.ID, duration), logrus.Fields{
		"user.id":         user.ID,
		"sql.duration_ms": duration.Milliseconds(),
	})

	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id int) (*entities.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.find_user_by_id")
	defer span.Finish()

	query := "SELECT id, name, email, created_at FROM users WHERE id = ?"

	r.logWithTrace(ctx, fmt.Sprintf("SQL: %s [params: id=%d]", query, id), logrus.Fields{
		"query":   query,
		"user.id": id,
	})

	var user entities.User

	startTime := time.Now()
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)
	duration := time.Since(startTime)

	if err == sql.ErrNoRows {
		r.logWithTrace(ctx, fmt.Sprintf("SQL result: User not found (id=%d) [duration: %v]", id, duration), logrus.Fields{
			"user.id":         id,
			"sql.duration_ms": duration.Milliseconds(),
		})
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		r.logErrorWithTrace(ctx, fmt.Sprintf("SQL Error: %s [params: id=%d] [duration: %v]", query, id, duration), err, logrus.Fields{
			"query":           query,
			"user.id":         id,
			"sql.duration_ms": duration.Milliseconds(),
		})
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	r.logWithTrace(ctx, fmt.Sprintf("SQL result: Found user ID=%d, name=%s [duration: %v]", user.ID, user.Name, duration), logrus.Fields{
		"user.id":         user.ID,
		"user.name":       user.Name,
		"sql.duration_ms": duration.Milliseconds(),
	})

	return &user, nil
}

// FindAll retrieves all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*entities.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.find_all_users")
	defer span.Finish()

	query := "SELECT id, name, email, created_at FROM users ORDER BY created_at DESC LIMIT 100"

	r.logWithTrace(ctx, fmt.Sprintf("SQL: %s", query), logrus.Fields{
		"query": query,
	})

	startTime := time.Now()
	rows, err := r.db.QueryContext(ctx, query)
	duration := time.Since(startTime)
	if err != nil {
		r.logErrorWithTrace(ctx, fmt.Sprintf("SQL Error: %s [duration: %v]", query, duration), err, logrus.Fields{
			"query":           query,
			"sql.duration_ms": duration.Milliseconds(),
		})
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			continue
		}
		users = append(users, &user)
	}

	r.logWithTrace(ctx, fmt.Sprintf("SQL result: Retrieved %d users [duration: %v]", len(users), duration), logrus.Fields{
		"users.count":     len(users),
		"sql.duration_ms": duration.Milliseconds(),
	})

	return users, nil
}

// logWithTrace logs a message with trace information
func (r *UserRepository) logWithTrace(ctx context.Context, message string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["component"] = "mysql"
	logging.LogWithTrace(ctx, r.logger, "repository", message, fields)
}

// logErrorWithTrace logs an error with trace information
func (r *UserRepository) logErrorWithTrace(ctx context.Context, message string, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["component"] = "mysql"
	logging.LogErrorWithTrace(ctx, r.logger, "repository", message, err, fields)
}
