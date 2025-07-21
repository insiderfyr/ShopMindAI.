#!/bin/bash

# ChatGPT Clone - Deployment Script
# =================================

set -e

echo "🚀 ChatGPT Clone - Enterprise Deployment Script"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}📋 Checking prerequisites...${NC}"
    
    local missing_deps=0
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed${NC}"
        echo "   Please install Docker from: https://docs.docker.com/get-docker/"
        missing_deps=1
    else
        echo -e "${GREEN}✅ Docker found$(NC}"
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed${NC}"
        echo "   Please install Docker Compose from: https://docs.docker.com/compose/install/"
        missing_deps=1
    else
        echo -e "${GREEN}✅ Docker Compose found${NC}"
    fi
    
    # Check kubectl (for production)
    if ! command -v kubectl &> /dev/null; then
        echo -e "${YELLOW}⚠️  kubectl not found (needed for Kubernetes deployment)${NC}"
    else
        echo -e "${GREEN}✅ kubectl found${NC}"
    fi
    
    if [ $missing_deps -eq 1 ]; then
        echo ""
        echo -e "${RED}Please install missing dependencies before continuing.${NC}"
        exit 1
    fi
    
    echo ""
}

# Deploy local development
deploy_local() {
    echo -e "${BLUE}🏠 Deploying local development environment...${NC}"
    echo ""
    
    # Create necessary directories
    mkdir -p infrastructure/docker/ssl
    
    # Start services
    echo -e "${YELLOW}Starting all services...${NC}"
    
    # Use docker compose or docker-compose
    if docker compose version &> /dev/null; then
        docker compose up -d
    else
        docker-compose up -d
    fi
    
    echo ""
    echo -e "${GREEN}✅ Local deployment complete!${NC}"
    echo ""
    echo -e "${BLUE}📌 Access points:${NC}"
    echo "   • Frontend (ChatGPT UI): http://localhost:3000"
    echo "   • API Gateway: http://localhost:8000"
    echo "   • User Service: http://localhost:8080"
    echo "   • Chat Service: http://localhost:8081"
    echo "   • Auth Service: http://localhost:8082"
    echo "   • Keycloak Admin: http://localhost:8180 (admin/admin123)"
    echo "   • Consul UI: http://localhost:8500"
    echo "   • Traefik Dashboard: http://localhost:8090"
    echo "   • Prometheus: http://localhost:9090"
    echo "   • Grafana: http://localhost:3001 (admin/admin123)"
    echo ""
    echo -e "${YELLOW}💡 Tips:${NC}"
    echo "   • Check logs: docker-compose logs -f [service-name]"
    echo "   • Stop all: docker-compose down"
    echo "   • Reset data: docker-compose down -v"
}

# Deploy to Kubernetes
deploy_kubernetes() {
    echo -e "${BLUE}☸️  Deploying to Kubernetes...${NC}"
    echo ""
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}❌ kubectl is required for Kubernetes deployment${NC}"
        exit 1
    fi
    
    # Create namespace
    echo -e "${YELLOW}Creating namespace...${NC}"
    kubectl apply -f infrastructure/kubernetes/base/namespace.yaml
    
    # Install Helm chart
    if command -v helm &> /dev/null; then
        echo -e "${YELLOW}Installing Helm chart...${NC}"
        helm upgrade --install chatgpt-clone infrastructure/helm/chatgpt-clone/ \
            --namespace chatgpt-clone \
            --create-namespace
    else
        echo -e "${YELLOW}Applying Kubernetes manifests...${NC}"
        kubectl apply -k infrastructure/kubernetes/overlays/production/
    fi
    
    echo ""
    echo -e "${GREEN}✅ Kubernetes deployment initiated!${NC}"
    echo ""
    echo -e "${BLUE}📌 Check deployment status:${NC}"
    echo "   kubectl get pods -n chatgpt-clone"
    echo "   kubectl get svc -n chatgpt-clone"
}

# Build images
build_images() {
    echo -e "${BLUE}🔨 Building Docker images...${NC}"
    echo ""
    
    # Build services
    services=("user-service" "chat-service" "auth-service")
    
    for service in "${services[@]}"; do
        echo -e "${YELLOW}Building $service...${NC}"
        docker build -t chatgpt-clone/$service:latest ./services/$service/
    done
    
    # Build frontend
    echo -e "${YELLOW}Building frontend...${NC}"
    docker build -t chatgpt-clone/web:latest ./apps/web/
    
    echo ""
    echo -e "${GREEN}✅ All images built successfully!${NC}"
}

# Main menu
show_menu() {
    echo -e "${BLUE}Select deployment option:${NC}"
    echo "1) Local Development (Docker Compose)"
    echo "2) Build Docker Images"
    echo "3) Deploy to Kubernetes"
    echo "4) Full Setup (Build + Local Deploy)"
    echo "5) Exit"
    echo ""
    read -p "Enter choice [1-5]: " choice
}

# Main execution
main() {
    check_prerequisites
    
    while true; do
        show_menu
        
        case $choice in
            1)
                deploy_local
                break
                ;;
            2)
                build_images
                break
                ;;
            3)
                deploy_kubernetes
                break
                ;;
            4)
                build_images
                deploy_local
                break
                ;;
            5)
                echo -e "${YELLOW}Exiting...${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Invalid option. Please try again.${NC}"
                ;;
        esac
    done
}

# Run main function
main