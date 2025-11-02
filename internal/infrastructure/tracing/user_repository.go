package tracing

import (
	"context"

	"github.com/kanehiroyuu/datadog-tour/internal/domain"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// UserRepositoryTracer wraps a UserRepository with tracing
type UserRepositoryTracer struct {
	repo domain.UserRepository
}

// NewUserRepositoryTracer creates a new tracing decorator for UserRepository
func NewUserRepositoryTracer(repo domain.UserRepository) domain.UserRepository {
	return &UserRepositoryTracer{
		repo: repo,
	}
}

// Create wraps the Create method with tracing
func (r *UserRepositoryTracer) Create(ctx context.Context, user *domain.User) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.create_user")
	defer span.Finish()

	// Add metadata
	span.SetTag("db.type", "mysql")
	span.SetTag("db.operation", "INSERT")
	span.SetTag("user.name", user.Name)
	span.SetTag("user.email", user.Email)

	err := r.repo.Create(ctx, user)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("query.success", false)
		return err
	}

	span.SetTag("user.id", user.ID)
	span.SetTag("query.success", true)
	return nil
}

// FindByID wraps the FindByID method with tracing
func (r *UserRepositoryTracer) FindByID(ctx context.Context, id int) (*domain.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.find_user_by_id")
	defer span.Finish()

	// Add metadata
	span.SetTag("db.type", "mysql")
	span.SetTag("db.operation", "SELECT")
	span.SetTag("user.id", id)

	user, err := r.repo.FindByID(ctx, id)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("query.success", false)
		return nil, err
	}

	// Add result metadata
	span.SetTag("user.name", user.Name)
	span.SetTag("user.email", user.Email)
	span.SetTag("query.success", true)
	return user, nil
}

// FindAll wraps the FindAll method with tracing
func (r *UserRepositoryTracer) FindAll(ctx context.Context) ([]*domain.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "mysql.find_all_users")
	defer span.Finish()

	// Add metadata
	span.SetTag("db.type", "mysql")
	span.SetTag("db.operation", "SELECT")

	users, err := r.repo.FindAll(ctx)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("query.success", false)
		return nil, err
	}

	span.SetTag("users.count", len(users))
	span.SetTag("query.success", true)
	return users, nil
}
