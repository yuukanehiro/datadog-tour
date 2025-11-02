package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/kanehiroyuu/datadog-tour/internal/domain"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// UserUseCase implements user business logic
type UserUseCase struct {
	userRepo  domain.UserRepository
	cacheRepo domain.CacheRepository
	logger    *logrus.Logger
}

// NewUserUseCase creates a new UserUseCase
func NewUserUseCase(userRepo domain.UserRepository, cacheRepo domain.CacheRepository, logger *logrus.Logger) *UserUseCase {
	return &UserUseCase{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
		logger:    logger,
	}
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "usecase.create_user")
	defer span.Finish()

	// Add input parameters to span
	span.SetTag("user.name", name)
	span.SetTag("user.email", email)

	uc.logWithTrace(ctx, "Creating user", logrus.Fields{
		"user.name": name,
		"user.email": email,
	})

	user := &domain.User{
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		uc.logErrorWithTrace(ctx, "Failed to create user in repository", err, nil)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Add created user ID to span
	span.SetTag("user.id", user.ID)

	uc.logWithTrace(ctx, "User created, setting cache", logrus.Fields{
		"user.id": user.ID,
	})

	// Cache the user
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	userData, _ := json.Marshal(user)
	if err := uc.cacheRepo.Set(ctx, cacheKey, string(userData)); err != nil {
		// Log error but don't fail the request
		span.SetTag("cache.error", err.Error())
		span.SetTag("cache.set", false)
		uc.logErrorWithTrace(ctx, "Failed to set user cache", err, logrus.Fields{
			"cache.key": cacheKey,
		})
	} else {
		span.SetTag("cache.set", true)
		uc.logWithTrace(ctx, "User cached successfully", logrus.Fields{
			"cache.key": cacheKey,
		})
	}

	return user, nil
}

// GetUser retrieves a user by ID with caching
func (uc *UserUseCase) GetUser(ctx context.Context, id int) (*domain.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "usecase.get_user")
	defer span.Finish()

	span.SetTag("user.id", id)

	uc.logWithTrace(ctx, "Getting user by ID", logrus.Fields{
		"user.id": id,
	})

	// Try cache first
	cacheKey := fmt.Sprintf("user:%d", id)
	cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)

	if err == nil && cachedData != "" {
		span.SetTag("cache.hit", true)
		span.SetTag("data.source", "cache")
		var user domain.User
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			span.SetTag("user.name", user.Name)
			span.SetTag("user.email", user.Email)
			uc.logWithTrace(ctx, "User found in cache", logrus.Fields{
				"user.id": user.ID,
				"cache.key": cacheKey,
			})
			return &user, nil
		}
	}

	span.SetTag("cache.hit", false)
	span.SetTag("data.source", "database")

	uc.logWithTrace(ctx, "Cache miss, fetching from database", logrus.Fields{
		"user.id": id,
	})

	// Get from repository
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		uc.logErrorWithTrace(ctx, "Failed to get user from repository", err, logrus.Fields{
			"user.id": id,
		})
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Add user metadata to span
	span.SetTag("user.name", user.Name)
	span.SetTag("user.email", user.Email)

	uc.logWithTrace(ctx, "User found in database, setting cache", logrus.Fields{
		"user.id": user.ID,
		"cache.key": cacheKey,
	})

	// Cache the result
	userData, _ := json.Marshal(user)
	if err := uc.cacheRepo.Set(ctx, cacheKey, string(userData)); err != nil {
		span.SetTag("cache.set", false)
		span.SetTag("cache.error", err.Error())
		uc.logErrorWithTrace(ctx, "Failed to set user cache", err, logrus.Fields{
			"cache.key": cacheKey,
		})
	} else {
		span.SetTag("cache.set", true)
	}

	return user, nil
}

// GetAllUsers retrieves all users
func (uc *UserUseCase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "usecase.get_all_users")
	defer span.Finish()

	span.SetTag("data.source", "database")

	uc.logWithTrace(ctx, "Fetching all users", nil)

	users, err := uc.userRepo.FindAll(ctx)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		uc.logErrorWithTrace(ctx, "Failed to fetch users from repository", err, nil)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	span.SetTag("users.count", len(users))
	span.SetTag("query.success", true)

	uc.logWithTrace(ctx, "Users fetched successfully", logrus.Fields{
		"users.count": len(users),
	})

	return users, nil
}

// logWithTrace logs a message with trace information
func (uc *UserUseCase) logWithTrace(ctx context.Context, message string, fields logrus.Fields) {
	logging.LogWithTrace(ctx, uc.logger, "usecase", message, fields)
}

// logErrorWithTrace logs an error with trace information
func (uc *UserUseCase) logErrorWithTrace(ctx context.Context, message string, err error, fields logrus.Fields) {
	logging.LogErrorWithTrace(ctx, uc.logger, "usecase", message, err, fields)
}
