#!/bin/bash

echo "🚀 ShopMindAI - Deploy Script"
echo "============================="

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    exit 1
fi

# Check Docker Compose
if ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed"
    exit 1
fi

echo "✅ Prerequisites check passed"

# Create necessary directories
mkdir -p infrastructure/docker/ssl

# Start all services
echo "🏗️ Starting all services..."
docker compose up -d

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📌 Access points:"
echo "   • Frontend: http://localhost:3000"
echo "   • API Gateway: http://localhost:8000"
echo "   • Keycloak: http://localhost:8180 (admin/admin123)"
echo "   • Consul: http://localhost:8500"
echo "   • Prometheus: http://localhost:9090"
echo "   • Grafana: http://localhost:3001 (admin/admin123)"
echo ""
echo "💡 Commands:"
echo "   • View logs: docker compose logs -f"
echo "   • Stop: docker compose down"
echo "   • Restart: docker compose restart" 