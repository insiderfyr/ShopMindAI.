package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/chatgpt-clone/user-management-service/internal/domain"
	"github.com/chatgpt-clone/user-management-service/internal/events"
	"github.com/chatgpt-clone/user-management-service/internal/handlers"
	"github.com/chatgpt-clone/user-management-service/internal/repository"
	"github.com/chatgpt-clone/user-management-service/internal/service"
	"github.com/chatgpt-clone/user-management-service/pkg/api/proto"
	
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	serviceName = "user-management-service"
	version     = "1.0.0"
)

func main() {
	// Initialize logger
	logger := initLogger()
	logger.Info("Starting User Management Service", 
		"version", version,
		"service", serviceName)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", "error", err)
	}

	// Initialize database with retry
	db, err := initDatabaseWithRetry(cfg.DatabaseURL, 5, time.Second*5)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize Redis
	redisClient := initRedis(cfg.RedisURL)
	defer redisClient.Close()

	// Initialize Kafka producer
	kafkaProducer := initKafkaProducer(cfg.KafkaBrokers)
	defer kafkaProducer.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, redisClient)

	// Initialize event publisher
	eventPublisher := events.NewKafkaEventPublisher(kafkaProducer, logger)

	// Initialize services
	userService := service.NewUserService(userRepo, eventPublisher, logger)

	// Initialize handlers
	httpHandler := handlers.NewHTTPHandler(userService, logger)
	grpcHandler := handlers.NewGRPCHandler(userService, logger)

	// Setup HTTP server
	httpServer := setupHTTPServer(httpHandler, cfg.HTTPPort)

	// Setup gRPC server
	grpcServer := setupGRPCServer(grpcHandler, cfg.GRPCPort)

	// Setup metrics server
	metricsServer := setupMetricsServer(cfg.MetricsPort)

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start servers
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	go func() {
		logger.Info("Starting gRPC server", "port", cfg.GRPCPort)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			logger.Fatal("Failed to listen", "error", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	go func() {
		logger.Info("Starting metrics server", "port", cfg.MetricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Metrics server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	logger.Info("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown servers
	httpServer.Shutdown(ctx)
	grpcServer.GracefulStop()
	metricsServer.Shutdown(ctx)

	logger.Info("Server shutdown complete")
}

// Configuration structure
type Config struct {
	HTTPPort     int
	GRPCPort     int
	MetricsPort  int
	DatabaseURL  string
	RedisURL     string
	KafkaBrokers []string
	LogLevel     string
	Environment  string
}

// Load configuration from environment and files
func loadConfig() (*Config, error) {
	return &Config{
		HTTPPort:     getEnvAsInt("HTTP_PORT", 8080),
		GRPCPort:     getEnvAsInt("GRPC_PORT", 50051),
		MetricsPort:  getEnvAsInt("METRICS_PORT", 9090),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://user:pass@localhost/userdb"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		KafkaBrokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}, nil
}

// Setup HTTP server with middleware
func setupHTTPServer(handler *handlers.HTTPHandler, port int) *http.Server {
	router := gin.New()
	
	// Middleware
	router.Use(gin.Recovery())
	router.Use(handlers.LoggerMiddleware())
	router.Use(handlers.TracingMiddleware())
	router.Use(handlers.MetricsMiddleware())
	router.Use(handlers.RateLimitMiddleware())
	router.Use(handlers.CORSMiddleware())

	// Health checks
	router.GET("/health", handler.Health)
	router.GET("/ready", handler.Ready)

	// API routes
	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", handler.CreateUser)
			users.GET("/:id", handler.GetUser)
			users.PUT("/:id", handler.UpdateUser)
			users.DELETE("/:id", handler.DeleteUser)
			users.GET("", handler.ListUsers)
			users.POST("/:id/preferences", handler.UpdatePreferences)
		}
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Setup gRPC server with interceptors
func setupGRPCServer(handler *handlers.GRPCHandler, port int) *grpc.Server {
	// Interceptors
	opts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
			otelgrpc.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			otelgrpc.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	}

	server := grpc.NewServer(opts...)
	
	// Register services
	proto.RegisterUserServiceServer(server, handler)
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	
	// Register metrics
	grpc_prometheus.Register(server)

	return server
}

// Setup metrics server
func setupMetricsServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}

// Initialize database with retry logic
func initDatabaseWithRetry(dsn string, maxRetries int, retryInterval time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			PrepareStmt: true,
			Logger:      newGormLogger(),
		})
		
		if err == nil {
			// Test connection
			sqlDB, _ := db.DB()
			if err = sqlDB.Ping(); err == nil {
				// Configure connection pool
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetMaxOpenConns(100)
				sqlDB.SetConnMaxLifetime(time.Hour)
				return db, nil
			}
		}

		log.Printf("Database connection attempt %d failed: %v", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

// Initialize Redis client
func initRedis(url string) *redis.Client {
	opt, _ := redis.ParseURL(url)
	client := redis.NewClient(opt)
	
	// Configure connection pool
	client.Options().PoolSize = 100
	client.Options().MinIdleConns = 10
	
	return client
}

// Initialize Kafka producer
func initKafkaProducer(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        "user-events",
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		Compression:  kafka.Snappy,
	}
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}