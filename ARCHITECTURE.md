# ChatGPT Clone - Enterprise Microservices Architecture

## ✅ RESPECTĂM ARHITECTURA ORIGINALĂ

```
+------------------------------------+
|          Internet / Users          |
+------------------------------------+
                  |
                  v
+------------------------------------+
| Load Balancer (NGINX open-source)  |
|   + CDN (Cloudflare Free Tier)     |
+------------------------------------+
                  |
                  v
+------------------------------------+
| API Gateway (Traefik open-source)  |
|   - Auth: JWT/OAuth2 via Keycloak  |
|   - Rate Limiting                  |
|   - Service Discovery (Consul OSS) |
+------------------------------------+
                  |
                  +-----------------+-----------------+
                  |                 |                 |
                  v                 v                 v
+-------------+   +-------------+   +-------------+   +-------------+
| User Service|   | Chat Service|   | Auth Service|   | Monitoring  |
| (Go)        |   | (Go)        |   | (Keycloak   |   | (Prometheus)|
| - User Mgmt |   | - Sessions  |   | Integration)|   | + Grafana   |
| - Profiles  |   | - History   |   +-------------+   +-------------+
+-------------+   +-------------+
                  |                 |
                  v                 v
+-------------+   +-------------+   +-------------+
| PostgreSQL  |   | Redis Cache |   | Kafka (Event|
| (w/ Citus)  |   | (Sessions)   |   | Driven)     |
+-------------+   +-------------+   +-------------+
                  |
                  v
+------------------------------------+
|  Kubernetes Cluster (open-source)  |
|  - Pods, Deployments, Helm Charts  |
|  - Auto-scaling (HPA/VPA)          |
+------------------------------------+
                  |
                  v
+------------------------------------+
|  CI/CD Pipeline (GitHub Actions)   |
|  - Build, Test, Deploy w/ ArgoCD   |
+------------------------------------+
```

## 🎯 Ce construim: ChatGPT Clone (NU Shopping Assistant)

- Interfață IDENTICĂ cu ChatGPT
- Conversații AI generale (nu shopping)
- Scalabil la miliarde de utilizatori
- Full microservices architecture

## 📁 Structura Corectă

```
/workspace/
├── apps/
│   └── web/                    # Next.js - UI identică ChatGPT
├── services/
│   ├── user-service/          # Go - User management
│   ├── chat-service/          # Go - Chat sessions, WebSocket
│   ├── auth-service/          # Keycloak integration
│   └── ai-service/            # LLM integration
├── infrastructure/
│   ├── kubernetes/            # K8s manifests
│   ├── helm/                  # Helm charts
│   └── terraform/             # Cloud resources
└── platform/
    ├── api-gateway/           # Traefik config
    ├── service-discovery/     # Consul
    └── monitoring/            # Prometheus + Grafana
```

## 🔧 Microservicii Complete

### 1. User Service (Go)
- User registration/login
- Profile management
- Preferences
- PostgreSQL with Citus sharding

### 2. Chat Service (Go)
- WebSocket real-time
- Conversation management
- Message history
- Redis pub/sub + Kafka events

### 3. Auth Service
- Keycloak SSO
- JWT tokens
- RBAC with Casbin
- OAuth2/OIDC

### 4. AI Service (Go)
- LLM integration (OpenAI/Ollama)
- Prompt management
- Response streaming
- Token management

## 🚀 Scalabilitate pentru Miliarde

1. **Database Sharding**
   - PostgreSQL Citus
   - Shard by user_id
   - Read replicas

2. **Caching Strategy**
   - Redis Cluster
   - Session caching
   - Response caching

3. **Event Driven**
   - Kafka for all events
   - Async processing
   - Event sourcing

4. **Kubernetes**
   - HPA/VPA autoscaling
   - Multi-region
   - Service mesh (Istio)

## 🔄 Event Flow

```
User Message → API Gateway → Chat Service → Kafka
                                   ↓
                            AI Service (LLM)
                                   ↓
                            Response Stream
                                   ↓
                            WebSocket → User
```

## ✅ Tehnologii (Open-Source Only)

- **Frontend**: Next.js 14 (UI identică ChatGPT)
- **Backend**: Go microservices
- **Auth**: Keycloak
- **Database**: PostgreSQL + Citus
- **Cache**: Redis Cluster
- **Queue**: Kafka
- **Orchestration**: Kubernetes
- **Service Mesh**: Istio
- **API Gateway**: Traefik
- **Service Discovery**: Consul
- **Monitoring**: Prometheus + Grafana
- **CI/CD**: GitHub Actions + ArgoCD