#!/bin/bash

# ShopMindAI Developer Setup Script
# Quick setup for local development

set -e

echo "🚀 ShopMindAI Backend - Developer Setup"
echo "======================================="
echo ""

# Check dependencies
echo "📋 Checking dependencies..."

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 is not installed. Please install it first."
        exit 1
    else
        echo "✅ $1 is installed"
    fi
}

check_command docker
check_command docker-compose
check_command go
check_command make
check_command curl

echo ""
echo "📦 Setting up environment..."

# Copy env file if not exists
if [ ! -f .env ]; then
    cp .env.example .env
    echo "✅ Created .env file from template"
else
    echo "ℹ️  .env file already exists"
fi

# Create necessary directories
mkdir -p logs data/postgres data/redis data/kafka monitoring/grafana monitoring/prometheus

echo ""
echo "🐳 Starting infrastructure..."
make docker-up

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 15

echo ""
echo "🏥 Running health check..."
./scripts/health-check.sh

echo ""
echo "📊 Creating Grafana dashboards..."
# Import Grafana dashboard
curl -X POST http://admin:admin123@localhost:3001/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @monitoring/grafana-dashboard.json || echo "⚠️  Grafana not ready yet"

echo ""
echo "✅ Setup complete!"
echo ""
echo "🎯 Quick Start Commands:"
echo "  make dev          - Start development environment"
echo "  make test         - Run tests"
echo "  make logs         - View logs"
echo "  make health       - Check service health"
echo "  make monitor      - Open monitoring dashboards"
echo ""
echo "📚 Documentation:"
echo "  README.md         - Architecture and setup"
echo "  API_DOCUMENTATION.md - API reference"
echo ""
echo "🔗 Service URLs:"
echo "  API Gateway:  http://localhost:8080"
echo "  Consul UI:    http://localhost:8500"
echo "  Prometheus:   http://localhost:9090"
echo "  Grafana:      http://localhost:3001 (admin/admin123)"
echo ""
echo "Happy coding! 🎉" 

# ShopMindAI Developer Setup Script
# Quick setup for local development

set -e

echo "🚀 ShopMindAI Backend - Developer Setup"
echo "======================================="
echo ""

# Check dependencies
echo "📋 Checking dependencies..."

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 is not installed. Please install it first."
        exit 1
    else
        echo "✅ $1 is installed"
    fi
}

check_command docker
check_command docker-compose
check_command go
check_command make
check_command curl

echo ""
echo "📦 Setting up environment..."

# Copy env file if not exists
if [ ! -f .env ]; then
    cp .env.example .env
    echo "✅ Created .env file from template"
else
    echo "ℹ️  .env file already exists"
fi

# Create necessary directories
mkdir -p logs data/postgres data/redis data/kafka monitoring/grafana monitoring/prometheus

echo ""
echo "🐳 Starting infrastructure..."
make docker-up

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 15

echo ""
echo "🏥 Running health check..."
./scripts/health-check.sh

echo ""
echo "📊 Creating Grafana dashboards..."
# Import Grafana dashboard
curl -X POST http://admin:admin123@localhost:3001/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @monitoring/grafana-dashboard.json || echo "⚠️  Grafana not ready yet"

echo ""
echo "✅ Setup complete!"
echo ""
echo "🎯 Quick Start Commands:"
echo "  make dev          - Start development environment"
echo "  make test         - Run tests"
echo "  make logs         - View logs"
echo "  make health       - Check service health"
echo "  make monitor      - Open monitoring dashboards"
echo ""
echo "📚 Documentation:"
echo "  README.md         - Architecture and setup"
echo "  API_DOCUMENTATION.md - API reference"
echo ""
echo "🔗 Service URLs:"
echo "  API Gateway:  http://localhost:8080"
echo "  Consul UI:    http://localhost:8500"
echo "  Prometheus:   http://localhost:9090"
echo "  Grafana:      http://localhost:3001 (admin/admin123)"
echo ""
echo "Happy coding! 🎉" 