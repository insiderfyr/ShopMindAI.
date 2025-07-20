package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "github.com/chatgpt-clone/services/user-service/proto"
)

// User model
type User struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	Email         string    `gorm:"uniqueIndex" json:"email"`
	Username      string    `gorm:"uniqueIndex" json:"username"`
	DisplayName   string    `json:"display_name"`
	Avatar        string    `json:"avatar"`
	Preferences   string    `json:"preferences"` // JSON string
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
}

// UserService implements the gRPC service
type UserService struct {
	pb.UnimplementedUserServiceServer
	db *gorm.DB
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	user := &User{
		ID:          uuid.New().String(),
		Email:       req.Email,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	if err := s.db.Create(user).Error; err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, err
	}

	return userToProto(user), nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	var user User
	if err := s.db.First(&user, "id = ?", req.Id).Error; err != nil {
		return nil, err
	}

	return userToProto(&user), nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	var user User
	if err := s.db.First(&user, "id = ?", req.Id).Error; err != nil {
		return nil, err
	}

	// Update fields
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Preferences != "" {
		user.Preferences = req.Preferences
	}
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return userToProto(&user), nil
}

// ListUsers lists all users with pagination
func (s *UserService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 20
	}
	offset := int(req.Offset)

	var users []User
	var total int64

	s.db.Model(&User{}).Count(&total)
	if err := s.db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}

	protoUsers := make([]*pb.User, len(users))
	for i, user := range users {
		protoUsers[i] = userToProto(&user)
	}

	return &pb.ListUsersResponse{
		Users: protoUsers,
		Total: int32(total),
	}, nil
}

// HTTP handlers for REST API
func setupHTTPServer(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "user-service",
		})
	})

	// User endpoints
	api := r.Group("/api/v1")
	{
		// Create user
		api.POST("/users", func(c *gin.Context) {
			var user User
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			user.ID = uuid.New().String()
			user.CreatedAt = time.Now()
			user.UpdatedAt = time.Now()
			user.IsActive = true

			if err := db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}

			c.JSON(http.StatusCreated, user)
		})

		// Get user
		api.GET("/users/:id", func(c *gin.Context) {
			var user User
			if err := db.First(&user, "id = ?", c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}

			c.JSON(http.StatusOK, user)
		})

		// Update user
		api.PUT("/users/:id", func(c *gin.Context) {
			var user User
			if err := db.First(&user, "id = ?", c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}

			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			user.UpdatedAt = time.Now()
			db.Save(&user)

			c.JSON(http.StatusOK, user)
		})

		// List users
		api.GET("/users", func(c *gin.Context) {
			var users []User
			db.Find(&users)
			c.JSON(http.StatusOK, users)
		})
	}

	return r
}

// Helper function to convert model to proto
func userToProto(user *User) *pb.User {
	return &pb.User{
		Id:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		DisplayName:   user.DisplayName,
		Avatar:        user.Avatar,
		Preferences:   user.Preferences,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Unix(),
		UpdatedAt:     user.UpdatedAt.Unix(),
	}
}

func main() {
	// Load environment variables
	godotenv.Load()

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/chatgpt_users?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatal("Failed to listen:", err)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterUserServiceServer(grpcServer, &UserService{db: db})

		log.Println("User Service gRPC server starting on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve:", err)
		}
	}()

	// Start HTTP server
	r := setupHTTPServer(db)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("User Service HTTP server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}