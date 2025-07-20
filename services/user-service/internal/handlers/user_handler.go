package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopgpt/user-service/internal/models"
	"github.com/shopgpt/user-service/internal/repository"
	"github.com/shopgpt/user-service/pkg/logger"
	"github.com/shopgpt/user-service/pkg/validator"
	"go.uber.org/zap"
)

type UserHandler struct {
	repo   repository.UserRepository
	logger *zap.Logger
	cache  repository.CacheRepository
}

func NewUserHandler(repo repository.UserRepository, cache repository.CacheRepository, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		repo:   repo,
		logger: logger,
		cache:  cache,
	}
}

// CreateUser handles user registration
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	// Validate request
	if err := validator.ValidateStruct(req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Validation failed", err)
		return
	}
	
	// Check if user exists
	existing, _ := h.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		h.respondError(w, http.StatusConflict, "User already exists", nil)
		return
	}
	
	// Create user
	user := &models.User{
		Email:     req.Email,
		Username:  req.Username,
		FullName:  req.FullName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := user.SetPassword(req.Password); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}
	
	if err := h.repo.Create(ctx, user); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}
	
	// Invalidate cache
	h.cache.Delete(ctx, "users:all")
	
	h.logger.Info("User created", zap.String("user_id", user.ID), zap.String("email", user.Email))
	h.respondJSON(w, http.StatusCreated, user)
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mux.Vars(r)["id"]
	
	// Check cache first
	cacheKey := "user:" + userID
	var user models.User
	if err := h.cache.Get(ctx, cacheKey, &user); err == nil {
		h.respondJSON(w, http.StatusOK, user)
		return
	}
	
	// Get from database
	dbUser, err := h.repo.GetByID(ctx, userID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "User not found", err)
		return
	}
	
	// Cache for 5 minutes
	h.cache.Set(ctx, cacheKey, dbUser, 5*time.Minute)
	
	h.respondJSON(w, http.StatusOK, dbUser)
}

// UpdateUser updates user information
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mux.Vars(r)["id"]
	
	// Get authenticated user ID from context
	authUserID, ok := ctx.Value("user_id").(string)
	if !ok || authUserID != userID {
		h.respondError(w, http.StatusForbidden, "Unauthorized", nil)
		return
	}
	
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	user, err := h.repo.GetByID(ctx, userID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "User not found", err)
		return
	}
	
	// Update fields
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	
	user.UpdatedAt = time.Now()
	
	if err := h.repo.Update(ctx, user); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to update user", err)
		return
	}
	
	// Invalidate cache
	h.cache.Delete(ctx, "user:"+userID)
	h.cache.Delete(ctx, "users:all")
	
	h.logger.Info("User updated", zap.String("user_id", user.ID))
	h.respondJSON(w, http.StatusOK, user)
}

// DeleteUser soft deletes a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mux.Vars(r)["id"]
	
	// Get authenticated user ID from context
	authUserID, ok := ctx.Value("user_id").(string)
	if !ok || authUserID != userID {
		h.respondError(w, http.StatusForbidden, "Unauthorized", nil)
		return
	}
	
	if err := h.repo.Delete(ctx, userID); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}
	
	// Invalidate cache
	h.cache.Delete(ctx, "user:"+userID)
	h.cache.Delete(ctx, "users:all")
	
	h.logger.Info("User deleted", zap.String("user_id", userID))
	h.respondJSON(w, http.StatusNoContent, nil)
}

// ListUsers returns paginated list of users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse query parameters
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	
	// Set defaults
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "20"
	}
	
	// Check cache for non-search queries
	cacheKey := "users:page:" + page + ":limit:" + limit
	if search == "" {
		var users []models.User
		if err := h.cache.Get(ctx, cacheKey, &users); err == nil {
			h.respondJSON(w, http.StatusOK, map[string]interface{}{
				"users": users,
				"page":  page,
				"limit": limit,
			})
			return
		}
	}
	
	// Get from database
	users, total, err := h.repo.List(ctx, page, limit, search)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list users", err)
		return
	}
	
	// Cache non-search results
	if search == "" {
		h.cache.Set(ctx, cacheKey, users, 2*time.Minute)
	}
	
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	
	user, err := h.repo.GetByID(ctx, userID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "User not found", err)
		return
	}
	
	h.respondJSON(w, http.StatusOK, user)
}

// UpdatePassword updates user password
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	
	var req models.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	user, err := h.repo.GetByID(ctx, userID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "User not found", err)
		return
	}
	
	// Verify old password
	if !user.CheckPassword(req.OldPassword) {
		h.respondError(w, http.StatusBadRequest, "Invalid old password", nil)
		return
	}
	
	// Set new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}
	
	user.UpdatedAt = time.Now()
	
	if err := h.repo.Update(ctx, user); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to update password", err)
		return
	}
	
	h.logger.Info("Password updated", zap.String("user_id", user.ID))
	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// Helper methods
func (h *UserHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *UserHandler) respondError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		h.logger.Error(message, zap.Error(err))
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}