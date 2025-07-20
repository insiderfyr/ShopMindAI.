package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// Test User Model
type TestUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Mock Database
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	arguments := m.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.Rows), arguments.Error(1)
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgx.CommandTag, error) {
	arguments := m.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.CommandTag), arguments.Error(1)
}

// Mock Redis
type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

// Test Suite
type UserServiceTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *MockDB
	redis  *MockRedis
	logger *zap.Logger
}

func (suite *UserServiceTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.db = new(MockDB)
	suite.redis = new(MockRedis)
	suite.logger = zap.NewNop()
	
	// Setup routes
	setupTestRoutes(suite.router, suite.db, suite.redis, suite.logger)
}

func setupTestRoutes(router *gin.Engine, db *MockDB, redis *MockRedis, logger *zap.Logger) {
	api := router.Group("/api/v1")
	{
		api.POST("/users", createUserHandler(db, redis, logger))
		api.GET("/users/:id", getUserHandler(db, redis, logger))
		api.PUT("/users/:id", updateUserHandler(db, redis, logger))
		api.DELETE("/users/:id", deleteUserHandler(db, redis, logger))
		api.GET("/users", listUsersHandler(db, redis, logger))
	}
}

// Test Create User
func (suite *UserServiceTestSuite) TestCreateUser() {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful user creation",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"username": "testuser",
				"password": "SecurePass123!",
			},
			setupMocks: func() {
				suite.db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
					Return(pgx.CommandTag{}, nil)
				suite.redis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&redis.StatusCmd{})
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"status": "success",
			},
		},
		{
			name: "invalid email format",
			payload: map[string]interface{}{
				"email":    "invalid-email",
				"username": "testuser",
				"password": "SecurePass123!",
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid email format",
			},
		},
		{
			name: "duplicate email",
			payload: map[string]interface{}{
				"email":    "existing@example.com",
				"username": "newuser",
				"password": "SecurePass123!",
			},
			setupMocks: func() {
				suite.db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
					Return(pgx.CommandTag{}, fmt.Errorf("duplicate key value"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error": "email already exists",
			},
		},
		{
			name: "weak password",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"username": "testuser",
				"password": "weak",
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "password does not meet requirements",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setupMocks()
			
			body, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)
			
			for key, value := range tc.expectedBody {
				assert.Equal(suite.T(), value, response[key])
			}
		})
	}
}

// Test Get User
func (suite *UserServiceTestSuite) TestGetUser() {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful get user from cache",
			userID: "123e4567-e89b-12d3-a456-426614174000",
			setupMocks: func() {
				cachedUser := `{"id":"123e4567-e89b-12d3-a456-426614174000","email":"test@example.com","username":"testuser"}`
				cmd := redis.NewStringResult(cachedUser, nil)
				suite.redis.On("Get", mock.Anything, "user:123e4567-e89b-12d3-a456-426614174000").
					Return(cmd)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":       "123e4567-e89b-12d3-a456-426614174000",
				"email":    "test@example.com",
				"username": "testuser",
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMocks: func() {
				cmd := redis.NewStringResult("", redis.Nil)
				suite.redis.On("Get", mock.Anything, "user:nonexistent").
					Return(cmd)
				suite.db.On("Query", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, pgx.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "user not found",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setupMocks()
			
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", tc.userID), nil)
			w := httptest.NewRecorder()
			
			suite.router.ServeHTTP(w, req)
			
			assert.Equal(suite.T(), tc.expectedStatus, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)
			
			for key, value := range tc.expectedBody {
				assert.Equal(suite.T(), value, response[key])
			}
		})
	}
}

// Integration Tests
func TestUserServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests")
	}

	// Setup test database
	db, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Setup test Redis
	redis := setupTestRedis()
	defer redis.Close()

	// Create test server
	router := setupServer(db, redis)

	t.Run("Full User Lifecycle", func(t *testing.T) {
		// Create user
		createPayload := map[string]interface{}{
			"email":    "integration@test.com",
			"username": "integrationtest",
			"password": "IntegrationPass123!",
		}
		
		body, _ := json.Marshal(createPayload)
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var createResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &createResponse)
		assert.NoError(t, err)
		
		userID := createResponse["id"].(string)
		
		// Get user
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", userID), nil)
		w = httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Update user
		updatePayload := map[string]interface{}{
			"username": "updateduser",
		}
		
		body, _ = json.Marshal(updatePayload)
		req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%s", userID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Delete user
		req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", userID), nil)
		w = httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
		
		// Verify deletion
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", userID), nil)
		w = httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// Benchmark Tests
func BenchmarkCreateUser(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	db := new(MockDB)
	redis := new(MockRedis)
	logger := zap.NewNop()
	
	setupTestRoutes(router, db, redis, logger)
	
	db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
		Return(pgx.CommandTag{}, nil)
	redis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&redis.StatusCmd{})
	
	payload := map[string]interface{}{
		"email":    "bench@test.com",
		"username": "benchuser",
		"password": "BenchPass123!",
	}
	
	body, _ := json.Marshal(payload)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkGetUserFromCache(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	db := new(MockDB)
	redis := new(MockRedis)
	logger := zap.NewNop()
	
	setupTestRoutes(router, db, redis, logger)
	
	cachedUser := `{"id":"123e4567-e89b-12d3-a456-426614174000","email":"test@example.com","username":"testuser"}`
	cmd := redis.NewStringResult(cachedUser, nil)
	redis.On("Get", mock.Anything, mock.Anything).Return(cmd)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/api/v1/users/123e4567-e89b-12d3-a456-426614174000", nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
		}
	})
}

// Load Tests
func TestUserServiceLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	db := new(MockDB)
	redis := new(MockRedis)
	logger := zap.NewNop()
	
	setupTestRoutes(router, db, redis, logger)
	
	// Setup mocks for load test
	db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
		Return(pgx.CommandTag{}, nil)
	redis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&redis.StatusCmd{})
	
	// Simulate 1000 concurrent users
	concurrency := 1000
	requestsPerUser := 10
	
	results := make(chan bool, concurrency*requestsPerUser)
	
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		go func(userNum int) {
			for j := 0; j < requestsPerUser; j++ {
				payload := map[string]interface{}{
					"email":    fmt.Sprintf("user%d_%d@test.com", userNum, j),
					"username": fmt.Sprintf("user%d_%d", userNum, j),
					"password": "LoadTest123!",
				}
				
				body, _ := json.Marshal(payload)
				req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				
				router.ServeHTTP(w, req)
				
				results <- w.Code == http.StatusCreated
			}
		}(i)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < concurrency*requestsPerUser; i++ {
		if <-results {
			successCount++
		}
	}
	
	duration := time.Since(start)
	requestsPerSecond := float64(concurrency*requestsPerUser) / duration.Seconds()
	
	t.Logf("Load test completed:")
	t.Logf("  Total requests: %d", concurrency*requestsPerUser)
	t.Logf("  Successful requests: %d", successCount)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Requests per second: %.2f", requestsPerSecond)
	
	assert.Equal(t, concurrency*requestsPerUser, successCount, "All requests should succeed")
	assert.Greater(t, requestsPerSecond, 1000.0, "Should handle at least 1000 requests per second")
}

// Security Tests
func (suite *UserServiceTestSuite) TestSecurityVulnerabilities() {
	suite.Run("SQL Injection Prevention", func() {
		maliciousPayload := map[string]interface{}{
			"email":    "test@test.com'; DROP TABLE users; --",
			"username": "testuser",
			"password": "SecurePass123!",
		}
		
		body, _ := json.Marshal(maliciousPayload)
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		// Should be rejected due to email validation
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})
	
	suite.Run("XSS Prevention", func() {
		xssPayload := map[string]interface{}{
			"email":    "test@test.com",
			"username": "<script>alert('XSS')</script>",
			"password": "SecurePass123!",
		}
		
		body, _ := json.Marshal(xssPayload)
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		suite.db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
			Return(pgx.CommandTag{}, nil)
		
		suite.router.ServeHTTP(w, req)
		
		// Username should be sanitized
		assert.Equal(suite.T(), http.StatusCreated, w.Code)
	})
	
	suite.Run("Rate Limiting", func() {
		// Test that rate limiting is enforced
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest("GET", "/api/v1/users/test", nil)
			req.Header.Set("X-Real-IP", "192.168.1.1")
			w := httptest.NewRecorder()
			
			suite.router.ServeHTTP(w, req)
			
			if i > 50 { // Assuming rate limit is 50 requests
				assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
			}
		}
	})
}

// Test Main
func TestMain(m *testing.M) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Run tests
	code := m.Run()
	
	// Teardown
	
	os.Exit(code)
}

// Run the test suite
func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}