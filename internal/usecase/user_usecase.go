package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/kanehiroyuu/datadog-tour/internal/domain/entities"
	"github.com/kanehiroyuu/datadog-tour/internal/usecase/port"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// UserUseCase implements user business logic
type UserUseCase struct {
	Logger port.Logger
	RUser  port.UserRepository
	RCache port.CacheRepository
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, name, email string) (*entities.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "usecase.create_user")
	defer span.Finish()

	// Add input parameters to span
	span.SetTag("user.name", name)
	span.SetTag("user.email", email)

	logging.LogWithTrace(ctx, uc.Logger, "usecase", "Creating user", logrus.Fields{
		"user.name":  name,
		"user.email": email,
	})

	user := &entities.User{
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}

	if err := uc.RUser.Create(ctx, user); err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		logging.LogErrorWithTrace(ctx, uc.Logger, "usecase", "Failed to create user in repository", err, nil)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Add created user ID to span
	span.SetTag("user.id", user.ID)

	logging.LogWithTrace(ctx, uc.Logger, "usecase", "User created, setting cache", logrus.Fields{
		"user.id": user.ID,
	})

	// Cache the user
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	userData, _ := json.Marshal(user)
	if err := uc.RCache.Set(ctx, cacheKey, string(userData)); err != nil {
		// Log error but don't fail the request
		span.SetTag("cache.error", err.Error())
		span.SetTag("cache.set", false)
		logging.LogErrorWithTrace(ctx, uc.Logger, "usecase", "Failed to set user cache", err, logrus.Fields{
			"cache.key": cacheKey,
		})
	} else {
		span.SetTag("cache.set", true)
		logging.LogWithTrace(ctx, uc.Logger, "usecase", "User cached successfully", logrus.Fields{
			"cache.key": cacheKey,
		})
	}

	return user, nil
}

// GetUser retrieves a user by ID with caching
func (uc *UserUseCase) GetUser(ctx context.Context, id int) (*entities.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "usecase.get_user")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	span.SetTag("user.id", id)

	logging.LogWithTrace(ctx, logger, "usecase", "Getting user by ID", logrus.Fields{
		"user.id": id,
	})

	// Try cache first
	cacheKey := fmt.Sprintf("user:%d", id)
	cachedData, err := uc.RCache.Get(ctx, cacheKey)

	if err == nil && cachedData != "" {
		span.SetTag("cache.hit", true)
		span.SetTag("data.source", "cache")
		var user entities.User
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			span.SetTag("user.name", user.Name)
			span.SetTag("user.email", user.Email)
			logging.LogWithTrace(ctx, logger, "usecase", "User found in cache", logrus.Fields{
				"user.id":   user.ID,
				"cache.key": cacheKey,
			})
			return &user, nil
		}
	}

	span.SetTag("cache.hit", false)
	span.SetTag("data.source", "database")

	logging.LogWithTrace(ctx, logger, "usecase", "Cache miss, fetching from database", logrus.Fields{
		"user.id": id,
	})

	// Get from repository
	user, err := uc.RUser.FindByID(ctx, id)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		logging.LogErrorWithTrace(ctx, logger, "usecase", "Failed to get user from repository", err, logrus.Fields{
			"user.id": id,
		})
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Add user metadata to span
	span.SetTag("user.name", user.Name)
	span.SetTag("user.email", user.Email)

	logging.LogWithTrace(ctx, logger, "usecase", "User found in database, setting cache", logrus.Fields{
		"user.id":   user.ID,
		"cache.key": cacheKey,
	})

	// Cache the result
	userData, _ := json.Marshal(user)
	if err := uc.RCache.Set(ctx, cacheKey, string(userData)); err != nil {
		span.SetTag("cache.set", false)
		span.SetTag("cache.error", err.Error())
		logging.LogErrorWithTrace(ctx, uc.Logger, "usecase", "Failed to set user cache", err, logrus.Fields{
			"cache.key": cacheKey,
		})
	} else {
		span.SetTag("cache.set", true)
	}

	return user, nil
}

// GetAllUsers retrieves all users
func (uc *UserUseCase) GetAllUsers(ctx context.Context) ([]*entities.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "usecase.get_all_users")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	span.SetTag("data.source", "database")

	logging.LogWithTrace(ctx, logger, "usecase", "Fetching all users", nil)

	users, err := uc.RUser.FindAll(ctx)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		logging.LogErrorWithTrace(ctx, logger, "usecase", "Failed to fetch users from repository", err, nil)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	span.SetTag("users.count", len(users))
	span.SetTag("query.success", true)

	logging.LogWithTrace(ctx, logger, "usecase", "Users fetched successfully", logrus.Fields{
		"users.count": len(users),
	})

	return users, nil
}

// TestPanic triggers a panic in repository layer for testing recovery middleware
func (uc *UserUseCase) TestPanic(ctx context.Context) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "usecase.test_panic")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)

	span.SetTag("test.type", "panic_recovery")

	logging.LogWithTrace(ctx, logger, "usecase", "Test panic called - will trigger repository panic", nil)

	// Call repository method that will panic
	err := uc.RUser.TestPanic(ctx)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		logging.LogErrorWithTrace(ctx, logger, "usecase", "Repository returned error", err, nil)
		return fmt.Errorf("test panic failed: %w", err)
	}

	// This line should never be reached due to panic
	return nil
}
