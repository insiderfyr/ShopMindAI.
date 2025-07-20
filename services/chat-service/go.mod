module github.com/shopmindai/shopmindai/services/chat-service

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/gorilla/websocket v1.5.1
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.4.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/google/uuid v1.5.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/viper v1.18.2
	github.com/prometheus/client_golang v1.18.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/stretchr/testify v1.8.4
	github.com/testcontainers/testcontainers-go v0.27.0
)