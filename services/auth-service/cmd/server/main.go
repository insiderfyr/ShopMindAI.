clpackage main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/casbin/casbin/v2"
	redisadapter "github.com/casbin/redis-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/shopmindai/shopmindai/services/auth-service/internal/config"
	"github.com/shopmindai/shopmindai/services/auth-service/internal/handlers"
	"github.com/shopmindai/shopmindai/services/auth-service/internal/middleware"
	"github.com/shopmindai/shopmindai/services/auth-service/internal/service"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Keycloak client
	keycloakClient := gocloak.NewClient(cfg.Keycloak.URL)
	
	// Optionally set custom http client with timeout
	restyClient := keycloakClient.RestyClient()
	restyClient.SetTimeout(30 * time.Second)

	// Initialize Redis adapter for Casbin (distributed RBAC)
	adapter, err := redisadapter.NewAdapter("tcp", cfg.Redis.Addr, redisadapter.WithPassword(cfg.Redis.Password))
	if err != nil {
		logger.Fatalf("Failed to create Redis adapter: %v", err)
	}

	// Initialize Casbin enforcer with RBAC model
	enforcer, err := casbin.NewEnforcer("configs/rbac_model.conf", adapter)
	if err != nil {
		logger.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	// Enable auto-save policy
	enforcer.EnableAutoSave(true)

	// Load policies
	if err := enforcer.LoadPolicy(); err != nil {
		logger.Warnf("Failed to load policies: %v", err)
		// Initialize default policies
		initializeDefaultPolicies(enforcer)
	}

	// Initialize services
	authService := service.NewAuthService(keycloakClient, enforcer, cfg, logger)
	tokenService := service.NewTokenService(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "auth-service",
			"version": cfg.Version,
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, tokenService, logger)

	// Public routes
	public := router.Group("/api/auth")
	{
		public.POST("/login", authHandler.Login)
		public.POST("/register", authHandler.Register)
		public.POST("/refresh", authHandler.RefreshToken)
		public.POST("/logout", authHandler.Logout)
		public.GET("/oauth/:provider", authHandler.OAuthLogin)
		public.POST("/oauth/callback", authHandler.OAuthCallback)
		public.POST("/forgot-password", authHandler.ForgotPassword)
		public.POST("/reset-password", authHandler.ResetPassword)
		public.POST("/verify-email", authHandler.VerifyEmail)
	}

	// Protected routes
	protected := router.Group("/api/auth")
	protected.Use(middleware.AuthMiddleware(tokenService, enforcer))
	{
		protected.GET("/me", authHandler.GetProfile)
		protected.PATCH("/profile", authHandler.UpdateProfile)
		protected.POST("/change-password", authHandler.ChangePassword)
		protected.DELETE("/account", authHandler.DeleteAccount)
		protected.GET("/sessions", authHandler.GetSessions)
		protected.DELETE("/sessions/:sessionId", authHandler.RevokeSession)
	}

	// Admin routes
	admin := router.Group("/api/auth/admin")
	admin.Use(middleware.AuthMiddleware(tokenService, enforcer))
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/users", authHandler.ListUsers)
		admin.GET("/users/:userId", authHandler.GetUser)
		admin.PATCH("/users/:userId", authHandler.UpdateUser)
		admin.DELETE("/users/:userId", authHandler.DeleteUser)
		admin.POST("/users/:userId/roles", authHandler.AssignRole)
		admin.DELETE("/users/:userId/roles/:role", authHandler.RemoveRole)
		admin.GET("/roles", authHandler.ListRoles)
		admin.POST("/roles", authHandler.CreateRole)
		admin.GET("/permissions", authHandler.ListPermissions)
		admin.POST("/permissions", authHandler.CreatePermission)
	}

	// Internal routes (for other services)
	internal := router.Group("/auth")
	internal.Use(middleware.InternalAuth(cfg.InternalToken))
	{
		internal.POST("/verify", authHandler.VerifyToken)
		internal.POST("/users/batch", authHandler.GetUsersBatch)
	}

	// WebSocket support for real-time auth events
	router.GET("/ws/auth", middleware.AuthMiddleware(tokenService, enforcer), authHandler.WebSocketHandler)

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Infof("Starting auth service on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

// Initialize default RBAC policies
func initializeDefaultPolicies(enforcer *casbin.Enforcer) {
	// Define roles
	enforcer.AddGroupingPolicy("admin", "user")
	
	// Admin permissions
	enforcer.AddPolicy("admin", "/api/auth/admin/*", "*")
	enforcer.AddPolicy("admin", "/api/chat/*", "*")
	enforcer.AddPolicy("admin", "/api/users/*", "*")
	
	// User permissions
	enforcer.AddPolicy("user", "/api/auth/me", "GET")
	enforcer.AddPolicy("user", "/api/auth/profile", "PATCH")
	enforcer.AddPolicy("user", "/api/auth/change-password", "POST")
	enforcer.AddPolicy("user", "/api/chat/conversations", "*")
	enforcer.AddPolicy("user", "/api/chat/conversations/*", "*")
	
	// Save policies
	enforcer.SavePolicy()
}