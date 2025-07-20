package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopgpt/user-service/internal/models"
	"github.com/shopgpt/user-service/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockUserRepository, *mocks.MockCacheRepository)
		expectedStatus int
		expectedBody   map[string]interface{}
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful user creation",
			requestBody: models.CreateUserRequest{
				Email:    "test@shopgpt.com",
				Username: "testuser",
				Password: "SecurePass123!",
				FullName: "Test User",
			},
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@shopgpt.com").Return(nil, errors.New("not found"))
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(1).(*models.User)
					user.ID = "user-123"
				})
				cacheRepo.On("Delete", mock.Anything, "users:all").Return(nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "user-123", resp["id"])
				assert.Equal(t, "test@shopgpt.com", resp["email"])
				assert.Equal(t, "testuser", resp["username"])
				assert.NotContains(t, resp, "password")
			},
		},
		{
			name: "user already exists",
			requestBody: models.CreateUserRequest{
				Email:    "existing@shopgpt.com",
				Username: "existing",
				Password: "SecurePass123!",
				FullName: "Existing User",
			},
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				existingUser := &models.User{
					ID:    "existing-123",
					Email: "existing@shopgpt.com",
				}
				userRepo.On("GetByEmail", mock.Anything, "existing@shopgpt.com").Return(existingUser, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error": "User already exists",
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name: "validation failure - missing email",
			requestBody: models.CreateUserRequest{
				Username: "testuser",
				Password: "SecurePass123!",
				FullName: "Test User",
			},
			mockSetup:      func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Validation failed",
			},
		},
		{
			name: "database error",
			requestBody: models.CreateUserRequest{
				Email:    "test@shopgpt.com",
				Username: "testuser",
				Password: "SecurePass123!",
				FullName: "Test User",
			},
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@shopgpt.com").Return(nil, errors.New("not found"))
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to create user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			userRepo := new(mocks.MockUserRepository)
			cacheRepo := new(mocks.MockCacheRepository)
			logger := zaptest.NewLogger(t)
			handler := NewUserHandler(userRepo, cacheRepo, logger)

			if tt.mockSetup != nil {
				tt.mockSetup(userRepo, cacheRepo)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			handler.CreateUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedBody != nil {
				for key, value := range tt.expectedBody {
					assert.Equal(t, value, response[key])
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			userRepo.AssertExpectations(t)
			cacheRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*mocks.MockUserRepository, *mocks.MockCacheRepository)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "get user from cache",
			userID: "user-123",
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				cachedUser := models.User{
					ID:       "user-123",
					Email:    "test@shopgpt.com",
					Username: "testuser",
				}
				cacheRepo.On("Get", mock.Anything, "user:user-123", mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(2).(*models.User)
					*user = cachedUser
				})
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":       "user-123",
				"email":    "test@shopgpt.com",
				"username": "testuser",
			},
		},
		{
			name:   "get user from database",
			userID: "user-456",
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				cacheRepo.On("Get", mock.Anything, "user:user-456", mock.AnythingOfType("*models.User")).Return(errors.New("cache miss"))
				
				dbUser := &models.User{
					ID:       "user-456",
					Email:    "db@shopgpt.com",
					Username: "dbuser",
				}
				userRepo.On("GetByID", mock.Anything, "user-456").Return(dbUser, nil)
				cacheRepo.On("Set", mock.Anything, "user:user-456", dbUser, 5*time.Minute).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":       "user-456",
				"email":    "db@shopgpt.com",
				"username": "dbuser",
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				cacheRepo.On("Get", mock.Anything, "user:nonexistent", mock.AnythingOfType("*models.User")).Return(errors.New("cache miss"))
				userRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "User not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			userRepo := new(mocks.MockUserRepository)
			cacheRepo := new(mocks.MockCacheRepository)
			logger := zaptest.NewLogger(t)
			handler := NewUserHandler(userRepo, cacheRepo, logger)

			if tt.mockSetup != nil {
				tt.mockSetup(userRepo, cacheRepo)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID})
			w := httptest.NewRecorder()

			// Execute
			handler.GetUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, value := range tt.expectedBody {
				assert.Equal(t, value, response[key])
			}

			userRepo.AssertExpectations(t)
			cacheRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		authUserID     string
		requestBody    interface{}
		mockSetup      func(*mocks.MockUserRepository, *mocks.MockCacheRepository)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful update",
			userID:     "user-123",
			authUserID: "user-123",
			requestBody: models.UpdateUserRequest{
				Username: "newusername",
				FullName: "New Name",
				Bio:      "Updated bio",
			},
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				existingUser := &models.User{
					ID:       "user-123",
					Email:    "test@shopgpt.com",
					Username: "oldusername",
					FullName: "Old Name",
				}
				userRepo.On("GetByID", mock.Anything, "user-123").Return(existingUser, nil)
				userRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
				cacheRepo.On("Delete", mock.Anything, "user:user-123").Return(nil)
				cacheRepo.On("Delete", mock.Anything, "users:all").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":       "user-123",
				"username": "newusername",
				"fullName": "New Name",
				"bio":      "Updated bio",
			},
		},
		{
			name:           "unauthorized - different user",
			userID:         "user-123",
			authUserID:     "user-456",
			requestBody:    models.UpdateUserRequest{Username: "hacker"},
			mockSetup:      func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error": "Unauthorized",
			},
		},
		{
			name:       "user not found",
			userID:     "nonexistent",
			authUserID: "nonexistent",
			requestBody: models.UpdateUserRequest{
				Username: "newusername",
			},
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				userRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "User not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			userRepo := new(mocks.MockUserRepository)
			cacheRepo := new(mocks.MockCacheRepository)
			logger := zaptest.NewLogger(t)
			handler := NewUserHandler(userRepo, cacheRepo, logger)

			if tt.mockSetup != nil {
				tt.mockSetup(userRepo, cacheRepo)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID})
			
			// Add auth context
			ctx := context.WithValue(req.Context(), "user_id", tt.authUserID)
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			// Execute
			handler.UpdateUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, value := range tt.expectedBody {
				assert.Equal(t, value, response[key])
			}

			userRepo.AssertExpectations(t)
			cacheRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdatePassword(t *testing.T) {
	tests := []struct {
		name           string
		authUserID     string
		requestBody    interface{}
		mockSetup      func(*mocks.MockUserRepository, *mocks.MockCacheRepository)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful password update",
			authUserID: "user-123",
			requestBody: models.UpdatePasswordRequest{
				OldPassword: "OldPass123!",
				NewPassword: "NewPass456!",
			},
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				user := &models.User{
					ID:       "user-123",
					Email:    "test@shopgpt.com",
					Password: "$2a$10$YPd8.p0B1Zg/0nM9YV3xZu8i8Rt6I0GyLBNNTgW8hRZ3A8YGxU6/m", // bcrypt of "OldPass123!"
				}
				userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
				userRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Password updated successfully",
			},
		},
		{
			name:       "invalid old password",
			authUserID: "user-123",
			requestBody: models.UpdatePasswordRequest{
				OldPassword: "WrongPass123!",
				NewPassword: "NewPass456!",
			},
			mockSetup: func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {
				user := &models.User{
					ID:       "user-123",
					Email:    "test@shopgpt.com",
					Password: "$2a$10$YPd8.p0B1Zg/0nM9YV3xZu8i8Rt6I0GyLBNNTgW8hRZ3A8YGxU6/m",
				}
				userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid old password",
			},
		},
		{
			name:           "unauthenticated request",
			authUserID:     "",
			requestBody:    models.UpdatePasswordRequest{},
			mockSetup:      func(userRepo *mocks.MockUserRepository, cacheRepo *mocks.MockCacheRepository) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "Unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			userRepo := new(mocks.MockUserRepository)
			cacheRepo := new(mocks.MockCacheRepository)
			logger := zaptest.NewLogger(t)
			handler := NewUserHandler(userRepo, cacheRepo, logger)

			if tt.mockSetup != nil {
				tt.mockSetup(userRepo, cacheRepo)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Add auth context if provided
			if tt.authUserID != "" {
				ctx := context.WithValue(req.Context(), "user_id", tt.authUserID)
				req = req.WithContext(ctx)
			}
			
			w := httptest.NewRecorder()

			// Execute
			handler.UpdatePassword(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, value := range tt.expectedBody {
				assert.Equal(t, value, response[key])
			}

			userRepo.AssertExpectations(t)
			cacheRepo.AssertExpectations(t)
		})
	}
}

// Benchmark tests
func BenchmarkUserHandler_CreateUser(b *testing.B) {
	userRepo := new(mocks.MockUserRepository)
	cacheRepo := new(mocks.MockCacheRepository)
	logger := zap.NewNop()
	handler := NewUserHandler(userRepo, cacheRepo, logger)

	userRepo.On("GetByEmail", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
	userRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	cacheRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	reqBody := models.CreateUserRequest{
		Email:    "bench@shopgpt.com",
		Username: "benchuser",
		Password: "BenchPass123!",
		FullName: "Bench User",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.CreateUser(w, req)
	}
}

func BenchmarkUserHandler_GetUser_Cache(b *testing.B) {
	userRepo := new(mocks.MockUserRepository)
	cacheRepo := new(mocks.MockCacheRepository)
	logger := zap.NewNop()
	handler := NewUserHandler(userRepo, cacheRepo, logger)

	cachedUser := models.User{
		ID:       "bench-123",
		Email:    "bench@shopgpt.com",
		Username: "benchuser",
	}
	cacheRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(2).(*models.User)
		*user = cachedUser
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users/bench-123", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "bench-123"})
		w := httptest.NewRecorder()
		handler.GetUser(w, req)
	}
}

func BenchmarkUserHandler_GetUser_DB(b *testing.B) {
	userRepo := new(mocks.MockUserRepository)
	cacheRepo := new(mocks.MockCacheRepository)
	logger := zap.NewNop()
	handler := NewUserHandler(userRepo, cacheRepo, logger)

	dbUser := &models.User{
		ID:       "bench-456",
		Email:    "benchdb@shopgpt.com",
		Username: "benchdbuser",
	}
	cacheRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("cache miss"))
	userRepo.On("GetByID", mock.Anything, mock.Anything).Return(dbUser, nil)
	cacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users/bench-456", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "bench-456"})
		w := httptest.NewRecorder()
		handler.GetUser(w, req)
	}
}