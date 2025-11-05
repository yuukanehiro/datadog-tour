package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kanehiroyuu/datadog-tour/internal/domain/entities"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// UserRepository implements entities.UserRepository for MySQL
type UserRepository struct {
	db *LoggingDB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB, logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		db: NewLoggingDB(db, logger),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.create_user")
	defer span.Finish()

	query := "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)"

	// SQL automatically logged by LoggingDB
	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = int(id)
	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id int) (*entities.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.find_user_by_id")
	defer span.Finish()

	query := "SELECT id, name, email, created_at FROM users WHERE id = ?"

	var user entities.User

	// SQL automatically logged by LoggingDB
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// FindAll retrieves all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*entities.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.find_all_users")
	defer span.Finish()

	query := "SELECT id, name, email, created_at FROM users ORDER BY created_at DESC LIMIT 100"

	// SQL automatically logged by LoggingDB
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
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

	return users, nil
}

// TestPanic deliberately triggers a panic to test recovery middleware
// This method is for testing purposes only
func (r *UserRepository) TestPanic(ctx context.Context) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.test_panic")
	defer span.Finish()

	// Execute a real query first to demonstrate SQL logging before panic
	query := "SELECT COUNT(*) FROM users"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// Deliberately trigger a panic
	panic("Deliberate panic in repository layer for testing recovery middleware")
}
