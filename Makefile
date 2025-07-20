.PHONY: help install dev build test deploy clean

# Colors
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

help: ## Show this help message
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "  ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## Development
install: ## Install all dependencies
	npm install
	cd services/user-service && go mod download
	cd services/chat-service && go mod download
	cd services/auth-service && go mod download

dev: ## Start development environment
	docker compose up -d postgres redis kafka zookeeper keycloak
	npm run dev

dev-full: ## Start full development environment with all services
	docker compose up -d
	npm run dev

## Building
build: ## Build all services
	npm run build

build-docker: ## Build all Docker images
	docker compose build --parallel

build-go: ## Build all Go services
	cd services/user-service && go build -o bin/server cmd/server/main.go
	cd services/chat-service && go build -o bin/server cmd/server/main.go
	cd services/auth-service && go build -o bin/server cmd/server/main.go

## Testing
test: ## Run all tests
	npm run test
	cd services && go test -v ./...

test-unit: ## Run unit tests only
	npm run test:unit
	cd services && go test -v -short ./...

test-e2e: ## Run E2E tests
	npm run test:e2e

test-integration: ## Run integration tests
	cd services && go test -v -tags=integration ./...

test-load: ## Run load tests
	k6 run tests/load/chat-load-test.js

## Linting & Formatting
lint: ## Run linters
	npm run lint
	cd services && golangci-lint run ./...

format: ## Format code
	npm run format
	cd services && go fmt ./...

## Database
db-migrate: ## Run database migrations
	cd services/user-service && go run cmd/migrate/main.go up
	cd services/chat-service && go run cmd/migrate/main.go up

db-rollback: ## Rollback database migrations
	cd services/user-service && go run cmd/migrate/main.go down
	cd services/chat-service && go run cmd/migrate/main.go down

db-seed: ## Seed database with test data
	cd services/user-service && go run cmd/seed/main.go

## Docker & Kubernetes
docker-up: ## Start Docker Compose stack
	docker compose up -d

docker-down: ## Stop Docker Compose stack
	docker compose down

docker-logs: ## Show Docker logs
	docker compose logs -f

k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -k infrastructure/k8s/overlays/development

k8s-delete: ## Delete from Kubernetes
	kubectl delete -k infrastructure/k8s/overlays/development

## Monitoring
monitoring-up: ## Start monitoring stack
	docker compose up -d prometheus grafana loki

grafana-open: ## Open Grafana dashboard
	open http://localhost:3001

prometheus-open: ## Open Prometheus
	open http://localhost:9090

## Security
security-scan: ## Run security scans
	trivy fs .
	cd apps/web && npm audit
	cd services && gosec ./...

## Utilities
clean: ## Clean build artifacts
	rm -rf apps/web/.next
	rm -rf apps/web/node_modules
	rm -rf services/*/bin
	docker compose down -v

reset: clean ## Reset everything (including volumes)
	docker system prune -af --volumes

logs: ## Show all logs
	docker compose logs -f

ps: ## Show running containers
	docker compose ps

## Git hooks
hooks: ## Install git hooks
	npm run prepare

## Generate
gen-proto: ## Generate protobuf files
	cd services/chat-service && \
		protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		proto/*.proto

gen-mocks: ## Generate mocks for testing
	cd services && go generate ./...

## Release
release: ## Create a new release
	npm run release

## Quick commands
up: docker-up ## Alias for docker-up
down: docker-down ## Alias for docker-down
logs: docker-logs ## Alias for docker-logs