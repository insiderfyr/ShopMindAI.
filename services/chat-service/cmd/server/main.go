package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/shopmindai/shopmindai/services/chat-service/internal/config"
	"github.com/shopmindai/shopmindai/services/chat-service/internal/repository"
	"github.com/shopmindai/shopmindai/services/chat-service/internal/service"
	"github.com/shopmindai/shopmindai/services/chat-service/internal/transport/grpc"
	"github.com/shopmindai/shopmindai/services/chat-service/internal/transport/websocket"
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

	// Initialize database connection
	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient := repository.NewRedisClient(cfg.Redis)
	defer redisClient.Close()

	// Initialize Kafka
	kafkaProducer, err := repository.NewKafkaProducer(cfg.Kafka)
	if err != nil {
		logger.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Initialize repositories
	chatRepo := repository.NewChatRepository(db)
	sessionRepo := repository.NewSessionRepository(redisClient)
	eventRepo := repository.NewEventRepository(kafkaProducer)

	// Initialize services
	chatService := service.NewChatService(chatRepo, sessionRepo, eventRepo, logger)

	// Setup gRPC server
	grpcServer := grpc.NewServer()
	grpcHandler := grpc.NewChatHandler(chatService, logger)
	grpcHandler.Register(grpcServer)

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		logger.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	go func() {
		logger.Infof("Starting gRPC server on port %d", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Setup HTTP server with Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Setup gRPC-Gateway
	ctx := context.Background()
	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	
	err = grpc.RegisterChatServiceHandlerFromEndpoint(ctx, gwMux, 
		fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort), opts)
	if err != nil {
		logger.Fatalf("Failed to register gRPC gateway: %v", err)
	}

	// Mount gRPC-Gateway
	router.Any("/api/chat/*path", gin.WrapH(gwMux))

	// Setup WebSocket handler
	wsHub := websocket.NewHub(chatService, logger)
	go wsHub.Run()

	router.GET("/ws", func(c *gin.Context) {
		websocket.HandleWebSocket(wsHub, c.Writer, c.Request)
	})

	// Start HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Infof("Starting HTTP server on port %d", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()

	logger.Info("Servers stopped")
}