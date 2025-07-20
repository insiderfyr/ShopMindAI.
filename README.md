# ChatGPT Clone - Enterprise-Ready Scalable Architecture

[![CI/CD](https://github.com/chatgpt-clone/chatgpt-clone/workflows/CI/badge.svg)](https://github.com/chatgpt-clone/chatgpt-clone/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🚀 Overview

Enterprise-ready ChatGPT clone built for billions of users with microservices architecture, using only open-source technologies.

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│         Load Balancer (NGINX)           │
│      + CDN (Cloudflare Free Tier)       │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│      API Gateway (Traefik)              │
│  • Auth: JWT/OAuth2 via Keycloak       │
│  • Rate Limiting                        │
│  • Service Discovery (Consul)           │
└──────────────────┬──────────────────────┘
                   │
      ┌────────────┼────────────┐
      │            │            │
┌─────▼─────┐ ┌───▼────┐ ┌─────▼─────┐
│User Service│ │Chat Svc│ │Auth Service│
│   (Go)     │ │  (Go)  │ │ (Keycloak) │
└─────┬─────┘ └───┬────┘ └───────────┘
      │           │
┌─────▼───────────▼──────┐
│  PostgreSQL + Citus    │
│  Redis + Kafka         │
└────────────────────────┘
```

## 🛠️ Tech Stack

- **Frontend**: Next.js 14, TailwindCSS (ChatGPT-identical UI)
- **Backend**: Go microservices with gRPC & REST
- **Database**: PostgreSQL with Citus sharding
- **Cache**: Redis Cluster
- **Queue**: Apache Kafka
- **Auth**: Keycloak (OAuth2/OIDC)
- **Service Discovery**: Consul
- **API Gateway**: Traefik
- **Load Balancer**: NGINX
- **Monitoring**: Prometheus + Grafana
- **Container Orchestration**: Kubernetes with Helm
- **CI/CD**: GitHub Actions + ArgoCD

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+
- Go 1.21+

### Local Development

1. **Clone the repository**
```bash
git clone https://github.com/chatgpt-clone/chatgpt-clone.git
cd chatgpt-clone
```

2. **Start all services**
```bash
docker-compose up -d
```

3. **Access the applications**
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8000
- Keycloak: http://localhost:8180 (admin/admin123)
- Consul UI: http://localhost:8500
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin/admin123)

### Production Deployment

1. **Deploy to Kubernetes**
```bash
# Create namespace
kubectl apply -f infrastructure/kubernetes/base/namespace.yaml

# Install with Helm
helm install chatgpt-clone infrastructure/helm/chatgpt-clone/

# Apply Kubernetes manifests
kubectl apply -k infrastructure/kubernetes/overlays/production/
```

2. **Setup monitoring**
```bash
# Prometheus & Grafana are included in Helm chart
kubectl get svc -n chatgpt-clone | grep -E "(prometheus|grafana)"
```

## 📁 Project Structure

```
.
├── apps/
│   └── web/                 # Next.js frontend
├── services/
│   ├── user-service/        # User management (Go)
│   ├── chat-service/        # Chat & WebSocket (Go)
│   └── auth-service/        # Keycloak integration
├── infrastructure/
│   ├── docker/              # Docker configurations
│   ├── kubernetes/          # K8s manifests
│   └── helm/                # Helm charts
├── .github/
│   └── workflows/           # CI/CD pipelines
└── docker-compose.yml       # Local development

```

## 🔧 Development

### Running a specific service
```bash
# User service
cd services/user-service
go run main.go

# Frontend
cd apps/web
npm run dev
```

### Running tests
```bash
# All tests
make test

# Specific service
cd services/user-service
go test ./...
```

## 🚀 Scaling to Billions

1. **Database Sharding**: PostgreSQL Citus shards by user_id
2. **Caching**: Redis cluster with 512MB per instance
3. **Auto-scaling**: HPA scales 3-100 pods per service
4. **CDN**: Cloudflare for static assets
5. **Event-driven**: Kafka for async processing

## 📊 Monitoring

- **Metrics**: Prometheus scrapes all services
- **Visualization**: Grafana dashboards
- **Alerts**: Configured in Prometheus rules
- **Logs**: Structured JSON to stdout

## 🔒 Security

- **Authentication**: Keycloak with OAuth2/OIDC
- **Authorization**: RBAC with Casbin
- **Secrets**: Kubernetes secrets
- **Network**: Service mesh ready (Istio compatible)
- **Scanning**: Trivy in CI/CD pipeline

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
