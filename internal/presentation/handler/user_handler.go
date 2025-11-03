package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	appcontext "github.com/kanehiroyuu/datadog-tour/internal/common/context"
	"github.com/kanehiroyuu/datadog-tour/internal/common/logging"
	"github.com/kanehiroyuu/datadog-tour/internal/usecase"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct{}

// NewUserHandler creates a new UserHandler
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	//  各層でtracer.StartSpanFromContext(ctx, "span_name")を呼ぶと、dd-trace-goが自動的に：
	//  - trace-idを生成（または親spanから継承）
	//  - span-idを生成
	//  - span.Finish()が呼ばれた時にDatadog Agentへ送信
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.create_user")
	defer span.Finish() // ここでspanを終了させる, これによりspanのdurationが計測される

	logger := appcontext.GetLogger(ctx)
	repoLocator := appcontext.GetRepoLocator(ctx)

	interactor := &usecase.UserUseCase{
		Logger: logger,
		RUser:  repoLocator.UserRepo,
		RCache: repoLocator.CacheRepo,
	}

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("http.user_agent", r.UserAgent())

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.LogErrorWithTrace(ctx, logger, "handler", "Failed to decode request body", err, nil)
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		RespondErrorWithTrace(ctx, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Add request data to span
	span.SetTag("user.name", req.Name)
	span.SetTag("user.email", req.Email)

	user, err := interactor.CreateUser(ctx, req.Name, req.Email)
	if err != nil {
		logging.LogErrorWithTrace(ctx, logger, "handler", "Failed to create user", err, nil)
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		RespondErrorWithTrace(ctx, w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Add result to span
	span.SetTag("user.id", user.ID)

	logging.LogWithTrace(ctx, logger, "handler", "User created successfully", nil)
	RespondSuccessWithTrace(ctx, w, http.StatusCreated, user, "User created successfully")
}

// GetUser handles GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.get_user")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)
	repoLocator := appcontext.GetRepoLocator(ctx)

	interactor := &usecase.UserUseCase{
		Logger: logger,
		RUser:  repoLocator.UserRepo,
		RCache: repoLocator.CacheRepo,
	}

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("http.user_agent", r.UserAgent())

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", "Invalid user ID")
		RespondErrorWithTrace(ctx, w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	span.SetTag("user.id", id)

	user, err := interactor.GetUser(ctx, id)
	if err != nil {
		logging.LogErrorWithTrace(ctx, logger, "handler", "Failed to get user", err, nil)
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		RespondErrorWithTrace(ctx, w, http.StatusNotFound, "User not found")
		return
	}

	// Add result metadata
	span.SetTag("user.name", user.Name)
	span.SetTag("user.email", user.Email)

	logging.LogWithTrace(ctx, logger, "handler", "User retrieved successfully", nil)
	RespondSuccessWithTrace(ctx, w, http.StatusOK, user, "")
}

// GetAllUsers handles GET /api/users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "handler.get_all_users")
	defer span.Finish()

	logger := appcontext.GetLogger(ctx)
	repoLocator := appcontext.GetRepoLocator(ctx)

	interactor := &usecase.UserUseCase{
		Logger: logger,
		RUser:  repoLocator.UserRepo,
		RCache: repoLocator.CacheRepo,
	}

	// Add request metadata to span
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	span.SetTag("http.user_agent", r.UserAgent())

	users, err := interactor.GetAllUsers(ctx)
	if err != nil {
		logging.LogErrorWithTrace(ctx, logger, "handler", "Failed to get users", err, nil)
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		RespondErrorWithTrace(ctx, w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	// Add result metadata
	span.SetTag("users.count", len(users))

	logging.LogWithTrace(ctx, logger, "handler", "Users retrieved successfully", nil)
	RespondSuccessWithTrace(ctx, w, http.StatusOK, users, "")
}
