.PHONY: help build test clean docker-up docker-down migrate

# Default target
help:
	@echo "ShopMindAI Backend - Available commands:"
	@echo "  make build          - Build all services"
	@echo "  make test           - Run all tests"
	@echo "  make docker-up      - Start all services with Docker"
	@echo "  make docker-down    - Stop all services"
	@echo "  make migrate        - Run database migrations"
	@echo "  make dev            - Start development environment"
	@echo "  make logs           - View service logs"
	@echo "  make clean          - Clean build artifacts"

# Build all services
build:
	@echo "Building User Service..."
	cd services/user-service && go build -o bin/user-service ./cmd/server
	@echo "Building Chat Service..."
	cd services/chat-service && go build -o bin/chat-service ./cmd/server
	@echo "Building Auth Service..."
	cd services/auth-service && go build -o bin/auth-service ./cmd/server
	@echo "Build complete!"

# Run tests
test:
	@echo "Running User Service tests..."
	cd services/user-service && go test ./...
	@echo "Running Chat Service tests..."
	cd services/chat-service && go test ./...
	@echo "Running Auth Service tests..."
	cd services/auth-service && go test ./...

# Docker commands
docker-up:
	docker-compose up -d
	@echo "Waiting for services to start..."
	@sleep 10
	@echo "Services running at:"
	@echo "- API Gateway: http://localhost:8080"
	@echo "- User Service: http://localhost:8081"
	@echo "- Chat Service: http://localhost:8082"
	@echo "- Auth Service: http://localhost:8083"
	@echo "- Consul UI: http://localhost:8500"
	@echo "- Prometheus: http://localhost:9090"
	@echo "- Grafana: http://localhost:3001 (admin/admin123)"

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v
	docker system prune -f

# Database migrations
migrate:
	@echo "Running database migrations..."
	docker exec -it shopmindai-postgres psql -U postgres -d shopmindai -f /docker-entrypoint-initdb.d/init-db.sql

# Development environment
dev:
	@echo "Starting development infrastructure..."
	docker-compose up -d postgres redis kafka zookeeper consul keycloak
	@echo "Waiting for infrastructure..."
	@sleep 15
	@echo "Infrastructure ready!"
	@echo "Starting microservices..."
	@make -j3 run-user-service run-chat-service run-auth-service

# Run individual services
run-user-service:
	cd services/user-service && go run ./cmd/server/main.go

run-chat-service:
	cd services/chat-service && go run ./cmd/server/main.go

run-auth-service:
	cd services/auth-service && go run ./cmd/server/main.go

# View logs
logs:
	docker-compose logs -f

logs-user:
	docker-compose logs -f user-service

logs-chat:
	docker-compose logs -f chat-service

logs-auth:
	docker-compose logs -f auth-service

# Health checks
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8081/health | jq . || echo "User Service: DOWN"
	@curl -s http://localhost:8082/health | jq . || echo "Chat Service: DOWN"
	@curl -s http://localhost:8083/health | jq . || echo "Auth Service: DOWN"

# Load testing
load-test:
	@echo "Running load tests..."
	k6 run tests/load/api-load-test.js

# Clean build artifacts
clean:
	rm -rf services/*/bin
	rm -rf services/*/vendor
	find . -name "*.test" -delete
	find . -name "*.out" -delete

# Database management
db-shell:
	docker exec -it shopmindai-postgres psql -U postgres -d shopmindai

db-backup:
	docker exec shopmindai-postgres pg_dump -U postgres shopmindai > backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore:
	@read -p "Enter backup filename: " backup; \
	docker exec -i shopmindai-postgres psql -U postgres shopmindai < $$backup

# Monitoring
monitor:
	@echo "Opening monitoring dashboards..."
	@open http://localhost:9090  # Prometheus
	@open http://localhost:3001  # Grafana
	@open http://localhost:8500  # Consul

# API documentation
docs:
	@echo "Generating API documentation..."
	swag init -g ./services/user-service/cmd/server/main.go -o ./docs/user-service
	swag init -g ./services/chat-service/cmd/server/main.go -o ./docs/chat-service
	@echo "API docs generated in ./docs/"

# Production build
prod-build:
	@echo "Building production images..."
	docker build -t shopmindai/user-service:latest -f services/user-service/Dockerfile services/user-service
	docker build -t shopmindai/chat-service:latest -f services/chat-service/Dockerfile services/chat-service
	docker build -t shopmindai/auth-service:latest -f services/auth-service/Dockerfile services/auth-service

# Kubernetes deployment
k8s-deploy:
	kubectl apply -f infrastructure/k8s/base/namespace.yaml
	kubectl apply -f infrastructure/k8s/deployments/
	kubectl apply -f infrastructure/k8s/services/
	kubectl apply -f infrastructure/k8s/ingress/

k8s-status:
	kubectl get all -n shopmindai

# Performance profiling
profile:
	go tool pprof http://localhost:8081/debug/pprof/profile?seconds=30
	go tool pprof http://localhost:8082/debug/pprof/profile?seconds=30
	go tool pprof http://localhost:8082/debug/pprof/profile?seconds=30