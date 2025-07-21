#!/bin/bash

# ShopMindAI Kubernetes Deployment Script
# Deploy backend to Kubernetes cluster

set -e

# Configuration
NAMESPACE="shopmindai"
REGISTRY="docker.io/shopmindai"
VERSION="${VERSION:-latest}"

echo "🚀 ShopMindAI Kubernetes Deployment"
echo "==================================="
echo "Namespace: $NAMESPACE"
echo "Registry: $REGISTRY"
echo "Version: $VERSION"
echo ""

# Create namespace if not exists
echo "📦 Creating namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Deploy infrastructure
echo "🏗️ Deploying infrastructure..."
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/postgres-cluster.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/redis-cluster.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/kafka-cluster.yaml

# Wait for infrastructure
echo "⏳ Waiting for infrastructure to be ready..."
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=postgres --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=redis --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=kafka --timeout=300s

# Deploy microservices
echo "🎯 Deploying microservices..."
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/user-service.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/chat-service.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/auth-service.yaml

# Deploy ingress
echo "🌐 Configuring ingress..."
kubectl apply -n $NAMESPACE -f infrastructure/k8s/ingress/nginx-ingress.yaml

# Wait for services
echo "⏳ Waiting for services to be ready..."
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=user-service --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=chat-service --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=auth-service --timeout=300s

# Get status
echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Deployment status:"
kubectl get all -n $NAMESPACE

echo ""
echo "🔗 Service URLs:"
kubectl get ingress -n $NAMESPACE

echo ""
echo "📝 To access the services:"
echo "kubectl port-forward -n $NAMESPACE svc/api-gateway 8080:80" 

# ShopMindAI Kubernetes Deployment Script
# Deploy backend to Kubernetes cluster

set -e

# Configuration
NAMESPACE="shopmindai"
REGISTRY="docker.io/shopmindai"
VERSION="${VERSION:-latest}"

echo "🚀 ShopMindAI Kubernetes Deployment"
echo "==================================="
echo "Namespace: $NAMESPACE"
echo "Registry: $REGISTRY"
echo "Version: $VERSION"
echo ""

# Create namespace if not exists
echo "📦 Creating namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Deploy infrastructure
echo "🏗️ Deploying infrastructure..."
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/postgres-cluster.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/redis-cluster.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/kafka-cluster.yaml

# Wait for infrastructure
echo "⏳ Waiting for infrastructure to be ready..."
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=postgres --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=redis --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=kafka --timeout=300s

# Deploy microservices
echo "🎯 Deploying microservices..."
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/user-service.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/chat-service.yaml
kubectl apply -n $NAMESPACE -f infrastructure/k8s/deployments/auth-service.yaml

# Deploy ingress
echo "🌐 Configuring ingress..."
kubectl apply -n $NAMESPACE -f infrastructure/k8s/ingress/nginx-ingress.yaml

# Wait for services
echo "⏳ Waiting for services to be ready..."
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=user-service --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=chat-service --timeout=300s
kubectl wait -n $NAMESPACE --for=condition=ready pod -l app=auth-service --timeout=300s

# Get status
echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Deployment status:"
kubectl get all -n $NAMESPACE

echo ""
echo "🔗 Service URLs:"
kubectl get ingress -n $NAMESPACE

echo ""
echo "📝 To access the services:"
echo "kubectl port-forward -n $NAMESPACE svc/api-gateway 8080:80" 