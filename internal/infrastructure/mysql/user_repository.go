package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/kanehiroyuu/datadog-tour/internal/domain"
	"github.com/sirupsen/logrus"
)

// UserRepository implements domain.UserRepository for MySQL (without tracing)
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
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)"

	r.logWithTrace(ctx, "Executing SQL query", logrus.Fields{
		"query": query,
		"user.name": user.Name,
		"user.email": user.Email,
	})

	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.CreatedAt)
	if err != nil {
		r.logErrorWithTrace(ctx, "Failed to execute SQL query", err, logrus.Fields{
			"query": query,
		})
		return fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logErrorWithTrace(ctx, "Failed to get last insert ID", err, nil)
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = int(id)

	r.logWithTrace(ctx, "User created in database", logrus.Fields{
		"user.id": user.ID,
	})

	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	query := "SELECT id, name, email, created_at FROM users WHERE id = ?"

	r.logWithTrace(ctx, "Executing SQL query", logrus.Fields{
		"query": query,
		"user.id": id,
	})

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		r.logWithTrace(ctx, "User not found in database", logrus.Fields{
			"user.id": id,
		})
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		r.logErrorWithTrace(ctx, "Failed to execute SQL query", err, logrus.Fields{
			"query": query,
		})
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	r.logWithTrace(ctx, "User found in database", logrus.Fields{
		"user.id": user.ID,
		"user.name": user.Name,
	})

	return &user, nil
}

// FindAll retrieves all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	query := "SELECT id, name, email, created_at FROM users ORDER BY created_at DESC LIMIT 100"

	r.logWithTrace(ctx, "Executing SQL query", logrus.Fields{
		"query": query,
	})

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logErrorWithTrace(ctx, "Failed to execute SQL query", err, logrus.Fields{
			"query": query,
		})
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			continue
		}
		users = append(users, &user)
	}

	r.logWithTrace(ctx, "Users retrieved from database", logrus.Fields{
		"users.count": len(users),
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
