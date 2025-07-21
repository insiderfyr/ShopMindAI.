#!/bin/bash

# ShopMindAI Health Check Script
# Verifică starea tuturor serviciilor backend

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🏥 ShopMindAI Backend Health Check"
echo "=================================="

# Function to check service health
check_service() {
    local name=$1
    local url=$2
    
    if curl -s -f -o /dev/null "$url"; then
        echo -e "${GREEN}✓${NC} $name: UP"
        return 0
    else
        echo -e "${RED}✗${NC} $name: DOWN"
        return 1
    fi
}

# Infrastructure services
echo -e "\n${YELLOW}Infrastructure Services:${NC}"
check_service "PostgreSQL" "http://localhost:5432" || true
check_service "Redis" "http://localhost:6379" || true
check_service "Kafka" "http://localhost:9092" || true
check_service "Consul" "http://localhost:8500/v1/status/leader"
check_service "Keycloak" "http://localhost:8080/auth/realms/master"

# Microservices
echo -e "\n${YELLOW}Microservices:${NC}"
check_service "User Service" "http://localhost:8081/health"
check_service "Chat Service" "http://localhost:8082/health"
check_service "Auth Service" "http://localhost:8083/health"

# API Gateway
echo -e "\n${YELLOW}API Gateway:${NC}"
check_service "Traefik" "http://localhost:8090/ping"

# Monitoring
echo -e "\n${YELLOW}Monitoring:${NC}"
check_service "Prometheus" "http://localhost:9090/-/healthy"
check_service "Grafana" "http://localhost:3001/api/health"

echo -e "\n=================================="
echo "Health check complete!" 

# ShopMindAI Health Check Script
# Verifică starea tuturor serviciilor backend

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🏥 ShopMindAI Backend Health Check"
echo "=================================="

# Function to check service health
check_service() {
    local name=$1
    local url=$2
    
    if curl -s -f -o /dev/null "$url"; then
        echo -e "${GREEN}✓${NC} $name: UP"
        return 0
    else
        echo -e "${RED}✗${NC} $name: DOWN"
        return 1
    fi
}

# Infrastructure services
echo -e "\n${YELLOW}Infrastructure Services:${NC}"
check_service "PostgreSQL" "http://localhost:5432" || true
check_service "Redis" "http://localhost:6379" || true
check_service "Kafka" "http://localhost:9092" || true
check_service "Consul" "http://localhost:8500/v1/status/leader"
check_service "Keycloak" "http://localhost:8080/auth/realms/master"

# Microservices
echo -e "\n${YELLOW}Microservices:${NC}"
check_service "User Service" "http://localhost:8081/health"
check_service "Chat Service" "http://localhost:8082/health"
check_service "Auth Service" "http://localhost:8083/health"

# API Gateway
echo -e "\n${YELLOW}API Gateway:${NC}"
check_service "Traefik" "http://localhost:8090/ping"

# Monitoring
echo -e "\n${YELLOW}Monitoring:${NC}"
check_service "Prometheus" "http://localhost:9090/-/healthy"
check_service "Grafana" "http://localhost:3001/api/health"

echo -e "\n=================================="
echo "Health check complete!" 