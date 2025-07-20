module github.com/shopmindai/shopmindai/services/auth-service

go 1.21

require (
	github.com/Nerzal/gocloak/v13 v13.9.0
	github.com/gin-gonic/gin v1.9.1
	github.com/redis/go-redis/v9 v9.4.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/viper v1.18.2
	github.com/prometheus/client_golang v1.18.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/casbin/casbin/v2 v2.82.0
	github.com/casbin/redis-adapter/v3 v3.0.1
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/stretchr/testify v1.8.4
	github.com/segmentio/kafka-go v0.4.47
)