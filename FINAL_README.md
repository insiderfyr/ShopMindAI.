# 🚀 ShopGPT - Enterprise-Grade AI Shopping Assistant

[![CI/CD Pipeline](https://github.com/shopgpt/shopgpt/workflows/ShopGPT%20CI%2FCD%20Pipeline/badge.svg)](https://github.com/shopgpt/shopgpt/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0-blue.svg)](https://www.typescriptlang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Ready-blue.svg)](https://kubernetes.io/)

## 🌟 Overview

ShopGPT is a **production-ready**, **highly scalable** AI-powered shopping assistant that helps users find the best products across multiple online stores. Built with a microservices architecture, it's designed to handle **billions of users** with zero-friction scaling.

### 🎯 Key Features

- **🤖 AI-Powered Chat**: Natural language product search and recommendations
- **🔍 Multi-Store Search**: Aggregate results from Amazon, BestBuy, Walmart, and more
- **💬 Real-time Communication**: WebSocket-based chat with streaming responses
- **📊 Smart Analytics**: Track user behavior and optimize recommendations
- **🔐 Enterprise Security**: OAuth2/OIDC authentication with Keycloak
- **🚀 Infinite Scale**: Kubernetes-native with auto-scaling capabilities
- **📈 Production Monitoring**: Full observability with Prometheus, Grafana, and Loki

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Load Balancer (NGINX)                      │
│                                    ↓                                │
│                         CDN (Cloudflare)                            │
└─────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────┐
│                      API Gateway (Traefik)                          │
│                    [Auth, Rate Limiting, SSL]                       │
└─────────────────────────────────────────────────────────────────────┘
                                    ↓
┌──────────────┬──────────────┬──────────────┬──────────────────────┐
│ User Service │ Chat Service │Search Service│  Analytics Service   │
│   (Go/gRPC)  │  (Go/WS/gRPC)│   (Go/gRPC)  │     (Go/gRPC)        │
└──────────────┴──────────────┴──────────────┴──────────────────────┘
         ↓              ↓              ↓                ↓
┌─────────────────────────────────────────────────────────────────────┐
│                    Data Layer & Infrastructure                      │
│  PostgreSQL  │     Redis      │    Kafka     │   Elasticsearch    │
│   (Citus)    │   (Cluster)    │  (Event Bus) │    (Search)        │
└─────────────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Kubernetes cluster (for production)
- Go 1.21+
- Node.js 20+ & pnpm 8+
- PostgreSQL 15+
- Redis 7+

### 🐳 Local Development

```bash
# Clone the repository
git clone https://github.com/shopgpt/shopgpt.git
cd shopgpt

# Start all services with Docker Compose
docker-compose up -d

# Install dependencies
pnpm install

# Run database migrations
make migrate

# Start development servers
pnpm run dev

# Access the application
open http://localhost:3000
```

### 🧪 Running Tests

```bash
# Run all tests
make test

# Run backend tests
make test-backend

# Run frontend tests
pnpm run test

# Run E2E tests
pnpm run test:e2e

# Run with coverage
make test-coverage
```

### 📦 Building for Production

```bash
# Build all Docker images
make docker-build-all

# Deploy to Kubernetes
make deploy-k8s ENV=production

# Or use Helm
helm install shopgpt ./k8s/helm/shopgpt \
  --namespace shopgpt \
  --values ./k8s/helm/shopgpt/values.production.yaml
```

## 🛠️ Technology Stack

### Backend
- **Languages**: Go 1.21
- **Framework**: Gin (HTTP), gRPC (Inter-service)
- **Database**: PostgreSQL 15 with Citus for sharding
- **Cache**: Redis Cluster
- **Message Queue**: Apache Kafka
- **Search**: Elasticsearch
- **Authentication**: Keycloak (OAuth2/OIDC)

### Frontend
- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript 5
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Real-time**: WebSocket with reconnection
- **Build Tool**: Turborepo + Vite

### Infrastructure
- **Container**: Docker with multi-stage builds
- **Orchestration**: Kubernetes with Helm
- **Service Mesh**: Istio (optional)
- **API Gateway**: Traefik
- **Load Balancer**: NGINX
- **CDN**: Cloudflare
- **Monitoring**: Prometheus + Grafana
- **Logging**: Loki + Promtail
- **Tracing**: Jaeger
- **CI/CD**: GitHub Actions + ArgoCD

## 📁 Project Structure

```
shopgpt/
├── apps/                      # Frontend applications
│   ├── web/                   # Main web application (Next.js)
│   ├── admin/                 # Admin dashboard
│   └── mobile/                # Mobile app (React Native)
├── services/                  # Backend microservices
│   ├── user-service/          # User management
│   ├── chat-service/          # WebSocket chat handler
│   ├── auth-service/          # Authentication & authorization
│   ├── search-service/        # Product search aggregator
│   ├── analytics-service/     # Analytics & tracking
│   └── notification-service/  # Email/Push notifications
├── packages/                  # Shared packages
│   ├── ui/                    # Shared UI components
│   ├── config/                # Shared configurations
│   └── types/                 # Shared TypeScript types
├── infrastructure/            # Infrastructure as Code
│   ├── terraform/             # Cloud infrastructure
│   ├── k8s/                   # Kubernetes manifests
│   ├── helm/                  # Helm charts
│   └── monitoring/            # Monitoring stack
├── .github/                   # GitHub Actions workflows
├── docs/                      # Documentation
└── scripts/                   # Utility scripts
```

## 🔧 Configuration

### Environment Variables

Create `.env` files for each service:

```bash
# Backend Services
DB_HOST=localhost
DB_PORT=5432
DB_NAME=shopgpt
DB_USER=postgres
DB_PASSWORD=secure_password

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_PREFIX=shopgpt

JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_ANALYTICS_ID=UA-XXXXXXXXX
```

## 📊 Monitoring & Observability

### Metrics
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3001` (admin/admin)

### Logs
- **Loki**: Centralized logging
- **Grafana**: Log visualization

### Traces
- **Jaeger**: `http://localhost:16686`

### Alerts
- Configured in Prometheus with Alertmanager
- Slack/PagerDuty integrations

## 🔒 Security

- **Authentication**: OAuth2/OIDC via Keycloak
- **Authorization**: RBAC with Casbin
- **API Security**: Rate limiting, CORS, CSRF protection
- **Data Encryption**: TLS 1.3, encrypted at rest
- **Secrets Management**: HashiCorp Vault
- **Security Scanning**: Trivy, Snyk, OWASP dependency check

## 🚀 Deployment

### Kubernetes Deployment

```bash
# Create namespace
kubectl create namespace shopgpt

# Install with Helm
helm install shopgpt ./k8s/helm/shopgpt \
  --namespace shopgpt \
  --values values.production.yaml

# Check deployment status
kubectl get pods -n shopgpt
```

### Scaling

```yaml
# Horizontal Pod Autoscaler
kubectl autoscale deployment chat-service \
  --min=3 --max=100 \
  --cpu-percent=70 \
  -n shopgpt

# Vertical Pod Autoscaler
kubectl apply -f k8s/vpa/
```

## 📈 Performance

- **Response Time**: <100ms (p95)
- **Throughput**: 100K+ requests/second
- **WebSocket Connections**: 1M+ concurrent
- **Search Latency**: <200ms across all stores
- **Availability**: 99.99% SLA

## 🧪 Testing Strategy

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: All API endpoints
- **E2E Tests**: Critical user journeys
- **Performance Tests**: k6 load testing
- **Security Tests**: OWASP ZAP scanning
- **Chaos Engineering**: Litmus chaos tests

## 📚 Documentation

- [API Documentation](./docs/api/)
- [Architecture Decisions](./docs/architecture/)
- [Development Guide](./docs/development/)
- [Deployment Guide](./docs/deployment/)
- [Security Guide](./docs/security/)
- [Performance Tuning](./docs/performance/)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

```bash
# Fork the repository
# Create your feature branch
git checkout -b feature/amazing-feature

# Commit your changes
git commit -m 'Add some amazing feature'

# Push to the branch
git push origin feature/amazing-feature

# Open a Pull Request
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Acknowledgments

- Built with ❤️ by the ShopGPT team
- Inspired by modern cloud-native architectures
- Special thanks to all contributors

## 📞 Support

- **Documentation**: https://docs.shopgpt.com
- **Community**: https://community.shopgpt.com
- **Email**: support@shopgpt.com
- **Discord**: https://discord.gg/shopgpt

---

<p align="center">
  Made with ❤️ by ShopGPT Team
</p>