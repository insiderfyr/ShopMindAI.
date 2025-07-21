# ShopMindAI Backend - Microservices Architecture

AI-powered shopping assistant backend with smart recommendations, real-time personalization, and scalable microservices architecture.

## 🏗️ Architecture Overview

```ascii
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                    INTERNET / CLIENTS                                │
└─────────────────────────────────────┬───────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                            LOAD BALANCER (NGINX)                                    │
│                         Ports: 80 (HTTP), 443 (HTTPS)                               │
└─────────────────────────────────────┬───────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                          API GATEWAY (Traefik)                                      │
│                              Port: 8080                                             │
│  • Rate Limiting (100 req/min)                                                      │
│  • JWT Validation                                                                   │
│  • Request Routing                                                                  │
│  • Service Discovery (Consul)                                                       │
└────────────────┬──────────────────┬──────────────────┬─────────────────────────────┘
                 │                  │                  │
      ┌──────────▼─────────┐ ┌─────▼──────────┐ ┌────▼───────────┐
      │   USER SERVICE     │ │  CHAT SERVICE  │ │  AUTH SERVICE  │
      │   Port: 8081       │ │  Port: 8082    │ │  Port: 8083    │
      │   gRPC: 50051      │ │  gRPC: 50052   │ │  (Keycloak)    │
      └──────────┬─────────┘ └─────┬──────────┘ └────────────────┘
                 │                  │
      ┌──────────▼─────────────────▼──────────┐
      │         MESSAGE BROKER (Kafka)        │
      │            Port: 9092                 │
      │  Topics:                              │
      │  • user-events                        │
      │  • chat-messages                      │
      │  • recommendations                     │
      └──────────────────────────────────────┘
                 │
      ┌──────────▼─────────────────────────────┐
      │          DATA LAYER                    │
      ├─────────────────┬──────────────────────┤
      │  PostgreSQL     │    Redis Cache       │
      │  (Citus Shard)  │    Port: 6379        │
      │  Port: 5432     │    • Sessions        │
      │  • User Data    │    • Hot Data        │
      │  • Chat History │    • Rate Limits     │
      └─────────────────┴──────────────────────┘
```

## 🚀 Services

### 1. User Service (Go)
**Base URL:** `http://localhost:8081`
**gRPC:** `localhost:50051`

#### REST Endpoints:

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/api/v1/users` | Create user | `{"email": "user@example.com", "name": "John Doe"}` | `{"id": "uuid", "email": "...", "name": "...", "created_at": "..."}` |
| GET | `/api/v1/users/:id` | Get user by ID | - | `{"id": "uuid", "email": "...", "name": "...", "preferences": {...}}` |
| PUT | `/api/v1/users/:id` | Update user | `{"name": "Jane Doe", "preferences": {...}}` | `{"id": "uuid", "updated_at": "..."}` |
| DELETE | `/api/v1/users/:id` | Delete user | - | `{"message": "User deleted successfully"}` |
| GET | `/api/v1/users/:id/preferences` | Get user preferences | - | `{"categories": ["electronics", "fashion"], "budget": {...}}` |
| POST | `/api/v1/users/:id/preferences` | Update preferences | `{"categories": ["..."], "budget": {"min": 0, "max": 1000}}` | `{"updated": true}` |

#### gRPC Methods:
```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (User);
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteResponse);
  rpc GetUserPreferences(GetUserRequest) returns (UserPreferences);
}
```

### 2. Chat Service (Go)
**Base URL:** `http://localhost:8082`
**gRPC:** `localhost:50052`
**WebSocket:** `ws://localhost:8082/ws`

#### REST Endpoints:

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/api/v1/chat/sessions` | Create chat session | `{"user_id": "uuid"}` | `{"session_id": "uuid", "created_at": "..."}` |
| GET | `/api/v1/chat/sessions/:id` | Get session | - | `{"session_id": "uuid", "messages": [...]}` |
| POST | `/api/v1/chat/messages` | Send message | `{"session_id": "uuid", "content": "Find me a laptop"}` | `{"message_id": "uuid", "response": "..."}` |
| GET | `/api/v1/chat/sessions/:id/messages` | Get messages | - | `{"messages": [{"role": "user", "content": "..."}, ...]}` |
| DELETE | `/api/v1/chat/sessions/:id` | Delete session | - | `{"deleted": true}` |

#### WebSocket Events:
```javascript
// Connect: ws://localhost:8082/ws?session_id=uuid

// Client -> Server
{
  "type": "message",
  "content": "Find me a gaming laptop under $1500"
}

// Server -> Client
{
  "type": "response",
  "content": "I found 3 great gaming laptops...",
  "recommendations": [
    {
      "product_id": "123",
      "name": "ASUS ROG Zephyrus",
      "price": 1299,
      "match_score": 0.95
    }
  ]
}

// Server -> Client (streaming)
{
  "type": "stream",
  "chunk": "Based on your preferences",
  "is_final": false
}
```

### 3. Auth Service (Keycloak Integration)
**Base URL:** `http://localhost:8083`
**Keycloak Admin:** `http://localhost:8080/auth`

#### Endpoints:

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/api/v1/auth/login` | User login | `{"email": "...", "password": "..."}` | `{"access_token": "jwt...", "refresh_token": "..."}` |
| POST | `/api/v1/auth/register` | Register user | `{"email": "...", "password": "...", "name": "..."}` | `{"user_id": "uuid", "email": "..."}` |
| POST | `/api/v1/auth/refresh` | Refresh token | `{"refresh_token": "..."}` | `{"access_token": "jwt...", "refresh_token": "..."}` |
| POST | `/api/v1/auth/logout` | Logout | `{"refresh_token": "..."}` | `{"message": "Logged out successfully"}` |
| GET | `/api/v1/auth/verify` | Verify token | Header: `Authorization: Bearer jwt...` | `{"valid": true, "user_id": "uuid"}` |

## 🔧 Infrastructure Components

### PostgreSQL (Citus)
- **Port:** 5432
- **Database:** shopmindai
- **Sharding:** By user_id
- **Replicas:** 3

### Redis
- **Port:** 6379
- **Use Cases:**
  - Session storage (TTL: 24h)
  - Rate limiting (sliding window)
  - Cache layer (TTL: 5min)
  - Real-time recommendations

### Kafka
- **Port:** 9092
- **Topics:**
  - `user-events`: User activity tracking
  - `chat-messages`: Chat history and analytics
  - `recommendations`: AI recommendations pipeline
  - `product-updates`: Real-time price/stock updates

### Consul (Service Discovery)
- **Port:** 8500
- **UI:** http://localhost:8500
- **Services registered:** All microservices auto-register

### Prometheus + Grafana
- **Prometheus:** http://localhost:9090
- **Grafana:** http://localhost:3000
- **Metrics:**
  - Request latency (p50, p95, p99)
  - Error rates
  - Service health
  - Database connections

## 🚦 API Gateway Features

### Rate Limiting
```yaml
default: 100 requests/minute
authenticated: 1000 requests/minute
premium: 10000 requests/minute
```

### JWT Structure
```json
{
  "sub": "user_uuid",
  "email": "user@example.com",
  "roles": ["user", "premium"],
  "exp": 1234567890,
  "iat": 1234567890
}
```

### CORS Configuration
```yaml
allowed_origins:
  - http://localhost:*
  - https://*.shopmindai.com
allowed_methods: [GET, POST, PUT, DELETE, OPTIONS]
allowed_headers: [Content-Type, Authorization]
```

## 🐳 Docker Deployment

```bash
# Start all services
docker-compose up -d

# Start only backend services
docker-compose up -d postgres redis kafka consul traefik

# Start microservices
docker-compose up -d user-service chat-service auth-service

# View logs
docker-compose logs -f [service-name]

# Scale services
docker-compose up -d --scale user-service=3 --scale chat-service=5
```

## 🔑 Environment Variables

```env
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=shopmindai
POSTGRES_USER=admin
POSTGRES_PASSWORD=secure_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=shopmindai-backend

# Auth
KEYCLOAK_URL=http://localhost:8080/auth
KEYCLOAK_REALM=shopmindai
KEYCLOAK_CLIENT_ID=backend-services
KEYCLOAK_CLIENT_SECRET=secret

# Service Ports
USER_SERVICE_PORT=8081
CHAT_SERVICE_PORT=8082
AUTH_SERVICE_PORT=8083

# gRPC Ports
USER_SERVICE_GRPC_PORT=50051
CHAT_SERVICE_GRPC_PORT=50052
```

## 📊 Monitoring & Health Checks

### Health Check Endpoints
All services expose: `GET /health`

Response:
```json
{
  "status": "healthy",
  "service": "user-service",
  "version": "1.0.0",
  "uptime": "2h 15m 30s",
  "dependencies": {
    "postgres": "healthy",
    "redis": "healthy",
    "kafka": "healthy"
  }
}
```

### Metrics Endpoint
All services expose: `GET /metrics` (Prometheus format)

## 🔒 Security

- **JWT** for authentication
- **mTLS** between services
- **Rate limiting** per IP/user
- **Input validation** on all endpoints
- **SQL injection protection**
- **XSS protection**
- **CORS properly configured**

## 🚀 Quick Start

```bash
# 1. Clone repository
git clone https://github.com/yourusername/shopmindai-backend.git

# 2. Start infrastructure
docker-compose up -d postgres redis kafka consul keycloak

# 3. Run migrations
make migrate

# 4. Start services
docker-compose up -d user-service chat-service auth-service

# 5. Check health
curl http://localhost:8080/health

# 6. View API docs
open http://localhost:8080/swagger
```

## 📈 Performance

- **Latency:** p99 < 100ms
- **Throughput:** 10,000 req/sec per service
- **Availability:** 99.9% uptime
- **Scalability:** Horizontal scaling up to 100 pods

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

MIT License - see LICENSE file for details
