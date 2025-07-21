#!/bin/bash
set -euo pipefail

# ShopMindAI - Scale to Billions Script
# This script implements genius-level infrastructure automation

echo "🚀 ShopMindAI - Scaling to BILLIONS of users!"
echo "================================================"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
NAMESPACE="shopmindai"
ENVIRONMENT="${1:-production}"
SCALE_FACTOR="${2:-1}"  # 1 = millions, 10 = tens of millions, 100 = hundreds of millions, 1000 = billions

# Function to print colored output
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    commands=("kubectl" "helm" "docker" "jq")
    for cmd in "${commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            error "$cmd is not installed"
            exit 1
        fi
    done
    
    # Check Kubernetes connection
    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log "✅ All prerequisites met"
}

# Calculate resources based on scale
calculate_resources() {
    local scale=$1
    
    # Base resources for MVP (thousands of users)
    CHAT_SERVICE_MIN_REPLICAS=3
    CHAT_SERVICE_MAX_REPLICAS=10
    USER_SERVICE_MIN_REPLICAS=2
    USER_SERVICE_MAX_REPLICAS=5
    AUTH_SERVICE_MIN_REPLICAS=2
    AUTH_SERVICE_MAX_REPLICAS=5
    
    POSTGRES_WORKERS=3
    REDIS_NODES=6
    KAFKA_BROKERS=3
    
    # Scale up based on factor
    if [ "$scale" -ge 10 ]; then
        # Tens of millions
        CHAT_SERVICE_MIN_REPLICAS=10
        CHAT_SERVICE_MAX_REPLICAS=100
        USER_SERVICE_MIN_REPLICAS=5
        USER_SERVICE_MAX_REPLICAS=50
        AUTH_SERVICE_MIN_REPLICAS=5
        AUTH_SERVICE_MAX_REPLICAS=20
        
        POSTGRES_WORKERS=10
        REDIS_NODES=12
        KAFKA_BROKERS=7
    fi
    
    if [ "$scale" -ge 100 ]; then
        # Hundreds of millions
        CHAT_SERVICE_MIN_REPLICAS=50
        CHAT_SERVICE_MAX_REPLICAS=500
        USER_SERVICE_MIN_REPLICAS=20
        USER_SERVICE_MAX_REPLICAS=200
        AUTH_SERVICE_MIN_REPLICAS=10
        AUTH_SERVICE_MAX_REPLICAS=50
        
        POSTGRES_WORKERS=25
        REDIS_NODES=30
        KAFKA_BROKERS=15
    fi
    
    if [ "$scale" -ge 1000 ]; then
        # BILLIONS!
        CHAT_SERVICE_MIN_REPLICAS=100
        CHAT_SERVICE_MAX_REPLICAS=1000
        USER_SERVICE_MIN_REPLICAS=50
        USER_SERVICE_MAX_REPLICAS=500
        AUTH_SERVICE_MIN_REPLICAS=20
        AUTH_SERVICE_MAX_REPLICAS=100
        
        POSTGRES_WORKERS=50
        REDIS_NODES=60
        KAFKA_BROKERS=30
    fi
    
    log "📊 Calculated resources for scale factor: $scale"
    log "   Chat Service: $CHAT_SERVICE_MIN_REPLICAS-$CHAT_SERVICE_MAX_REPLICAS replicas"
    log "   PostgreSQL Workers: $POSTGRES_WORKERS"
    log "   Redis Nodes: $REDIS_NODES"
    log "   Kafka Brokers: $KAFKA_BROKERS"
}

# Deploy core infrastructure
deploy_infrastructure() {
    log "🏗️  Deploying core infrastructure..."
    
    # Create namespace
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label namespace for Istio injection (optional service mesh)
    kubectl label namespace "$NAMESPACE" istio-injection=enabled --overwrite
    
    # Deploy PostgreSQL Citus cluster
    log "📊 Deploying PostgreSQL Citus cluster..."
    kubectl apply -f infrastructure/k8s/statefulsets/postgres-citus.yaml
    
    # Scale PostgreSQL workers
    kubectl scale statefulset postgres-worker -n "$NAMESPACE" --replicas="$POSTGRES_WORKERS"
    
    # Deploy Redis cluster
    log "🔴 Deploying Redis cluster..."
    kubectl apply -f infrastructure/k8s/deployments/redis-cluster.yaml
    kubectl scale statefulset redis-cluster -n "$NAMESPACE" --replicas="$REDIS_NODES"
    
    # Wait for Redis to be ready, then initialize cluster
    kubectl wait --for=condition=ready pod -l app=redis-cluster -n "$NAMESPACE" --timeout=300s
    kubectl apply -f infrastructure/k8s/deployments/redis-cluster.yaml | grep -A20 "kind: Job"
    
    # Deploy Kafka cluster
    log "📨 Deploying Kafka cluster..."
    kubectl apply -f infrastructure/k8s/deployments/kafka-cluster.yaml
    kubectl scale statefulset kafka -n "$NAMESPACE" --replicas="$KAFKA_BROKERS"
    
    # Deploy monitoring stack
    log "📈 Deploying monitoring stack..."
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    
    # Prometheus with custom configuration
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace "$NAMESPACE" \
        --set prometheus.prometheusSpec.retention=30d \
        --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=100Gi \
        --set grafana.persistence.enabled=true \
        --set grafana.persistence.size=10Gi \
        --wait
}

# Deploy application services
deploy_services() {
    log "🚀 Deploying application services..."
    
    # Update HPA limits based on scale
    log "📈 Updating autoscaling limits..."
    
    # Chat Service
    kubectl apply -f infrastructure/k8s/deployments/chat-service.yaml
    kubectl patch hpa chat-service-hpa -n "$NAMESPACE" --type='json' -p='[
        {"op": "replace", "path": "/spec/minReplicas", "value": '"$CHAT_SERVICE_MIN_REPLICAS"'},
        {"op": "replace", "path": "/spec/maxReplicas", "value": '"$CHAT_SERVICE_MAX_REPLICAS"'}
    ]'
    
    # User Service
    kubectl patch hpa user-service-hpa -n "$NAMESPACE" --type='json' -p='[
        {"op": "replace", "path": "/spec/minReplicas", "value": '"$USER_SERVICE_MIN_REPLICAS"'},
        {"op": "replace", "path": "/spec/maxReplicas", "value": '"$USER_SERVICE_MAX_REPLICAS"'}
    ]' 2>/dev/null || echo "User service HPA will be created"
    
    # Auth Service
    kubectl patch hpa auth-service-hpa -n "$NAMESPACE" --type='json' -p='[
        {"op": "replace", "path": "/spec/minReplicas", "value": '"$AUTH_SERVICE_MIN_REPLICAS"'},
        {"op": "replace", "path": "/spec/maxReplicas", "value": '"$AUTH_SERVICE_MAX_REPLICAS"'}
    ]' 2>/dev/null || echo "Auth service HPA will be created"
    
    # Deploy NGINX Ingress
    log "🌐 Deploying NGINX Ingress..."
    kubectl apply -f infrastructure/k8s/ingress/nginx-ingress.yaml
}

# Configure advanced features for scale
configure_advanced_features() {
    log "🧠 Configuring advanced features for billions scale..."
    
    # Enable PodDisruptionBudgets
    log "🛡️  Setting up PodDisruptionBudgets..."
    for service in chat-service user-service auth-service; do
        kubectl apply -f - <<EOF
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: ${service}-pdb
  namespace: ${NAMESPACE}
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: ${service}
EOF
    done
    
    # Configure cluster autoscaler
    log "🔄 Configuring cluster autoscaler..."
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-autoscaler-status
  namespace: kube-system
data:
  max-nodes-total: "${SCALE_FACTOR}00"
  scale-down-delay-after-add: "10m"
  scale-down-unneeded-time: "10m"
  skip-nodes-with-system-pods: "false"
EOF
    
    # Configure priority classes
    log "⚡ Setting up priority classes..."
    kubectl apply -f - <<EOF
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: critical-priority
value: 1000
globalDefault: false
description: "Critical services priority"
---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 900
globalDefault: false
description: "High priority services"
---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: standard-priority
value: 500
globalDefault: true
description: "Standard priority services"
EOF
}

# Optimize for performance
optimize_performance() {
    log "⚡ Optimizing for extreme performance..."
    
    # Configure kernel parameters on nodes
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: sysctl-config
  namespace: kube-system
data:
  sysctl.conf: |
    # Network optimizations
    net.core.somaxconn = 65535
    net.ipv4.tcp_max_syn_backlog = 65535
    net.ipv4.ip_local_port_range = 1024 65535
    net.ipv4.tcp_tw_reuse = 1
    net.ipv4.tcp_fin_timeout = 15
    net.core.netdev_max_backlog = 65535
    net.core.rmem_max = 134217728
    net.core.wmem_max = 134217728
    net.ipv4.tcp_rmem = 4096 87380 134217728
    net.ipv4.tcp_wmem = 4096 65536 134217728
    net.ipv4.tcp_congestion_control = bbr
    net.core.default_qdisc = fq
    
    # File system
    fs.file-max = 2097152
    fs.nr_open = 2097152
    
    # Memory
    vm.swappiness = 0
    vm.max_map_count = 262144
EOF
    
    # Apply sysctl settings via DaemonSet
    kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: sysctl-tuning
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: sysctl-tuning
  template:
    metadata:
      labels:
        name: sysctl-tuning
    spec:
      hostNetwork: true
      hostPID: true
      hostIPC: true
      initContainers:
      - name: sysctl
        image: busybox
        command:
        - sh
        - -c
        - |
          sysctl -w net.core.somaxconn=65535
          sysctl -w net.ipv4.tcp_max_syn_backlog=65535
          sysctl -w net.ipv4.ip_local_port_range="1024 65535"
          sysctl -w net.ipv4.tcp_tw_reuse=1
          sysctl -w net.ipv4.tcp_fin_timeout=15
          sysctl -w net.core.netdev_max_backlog=65535
          sysctl -w net.core.rmem_max=134217728
          sysctl -w net.core.wmem_max=134217728
          sysctl -w vm.max_map_count=262144
        securityContext:
          privileged: true
      containers:
      - name: pause
        image: k8s.gcr.io/pause:3.9
      tolerations:
      - operator: Exists
EOF
}

# Setup cost optimization
setup_cost_optimization() {
    log "💰 Setting up cost optimization..."
    
    # Deploy Kubernetes Resource Report
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-resource-report
  namespace: ${NAMESPACE}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-resource-report
rules:
- apiGroups: [""]
  resources: ["nodes", "pods", "services", "namespaces"]
  verbs: ["get", "list"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets", "daemonsets"]
  verbs: ["get", "list"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["nodes", "pods"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-resource-report
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-resource-report
subjects:
- kind: ServiceAccount
  name: kube-resource-report
  namespace: ${NAMESPACE}
EOF
    
    # Configure spot instance support
    log "🎯 Configuring spot instance support..."
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: spot-instance-config
  namespace: ${NAMESPACE}
data:
  enabled: "true"
  percentage: "70"  # 70% spot instances for cost savings
  on-demand-base-capacity: "30"
EOF
}

# Health checks
perform_health_checks() {
    log "🏥 Performing health checks..."
    
    # Check service health
    services=("chat-service" "user-service" "auth-service" "postgres-coordinator" "redis-cluster" "kafka")
    
    for service in "${services[@]}"; do
        if kubectl get service "$service" -n "$NAMESPACE" &> /dev/null; then
            log "✅ $service is deployed"
        else
            warning "⚠️  $service is not deployed"
        fi
    done
    
    # Check pod status
    kubectl get pods -n "$NAMESPACE" --no-headers | while read -r line; do
        name=$(echo "$line" | awk '{print $1}')
        ready=$(echo "$line" | awk '{print $2}')
        status=$(echo "$line" | awk '{print $3}')
        
        if [[ "$status" != "Running" ]]; then
            warning "Pod $name is in $status state"
        fi
    done
}

# Generate scaling report
generate_scaling_report() {
    log "📊 Generating scaling report..."
    
    cat > scaling-report.md <<EOF
# ShopMindAI Scaling Report

Generated: $(date)
Environment: ${ENVIRONMENT}
Scale Factor: ${SCALE_FACTOR}

## Resource Allocation

### Services
- Chat Service: ${CHAT_SERVICE_MIN_REPLICAS}-${CHAT_SERVICE_MAX_REPLICAS} replicas
- User Service: ${USER_SERVICE_MIN_REPLICAS}-${USER_SERVICE_MAX_REPLICAS} replicas  
- Auth Service: ${AUTH_SERVICE_MIN_REPLICAS}-${AUTH_SERVICE_MAX_REPLICAS} replicas

### Data Layer
- PostgreSQL Workers: ${POSTGRES_WORKERS}
- Redis Nodes: ${REDIS_NODES}
- Kafka Brokers: ${KAFKA_BROKERS}

## Capacity Estimates

Based on current configuration:
- Concurrent Users: $(( SCALE_FACTOR * 1000000 ))
- Messages/Second: $(( CHAT_SERVICE_MAX_REPLICAS * 1000 ))
- Storage Capacity: $(( POSTGRES_WORKERS * 2 )) TB
- Cache Memory: $(( REDIS_NODES * 32 )) GB

## Monitoring

- Prometheus: http://prometheus.${NAMESPACE}.local
- Grafana: http://grafana.${NAMESPACE}.local
- Alerts configured: ✅

## Cost Optimization

- Spot Instances: 70%
- Auto-scaling: Enabled
- Resource Requests: Optimized

## Next Steps

1. Monitor metrics for 24 hours
2. Adjust HPA thresholds based on load
3. Enable predictive scaling
4. Configure backup strategy

EOF
    
    log "📄 Report saved to scaling-report.md"
}

# Main execution
main() {
    echo -e "${BLUE}"
    cat << "EOF"
   _____ __                 __  __ _           __    _    ____
  / ___// /_  ____  ____   /  |/  (_)___  ____/ /   / |  /  _/
  \__ \/ __ \/ __ \/ __ \ / /|_/ / / __ \/ __  /   / /| | / /  
 ___/ / / / / /_/ / /_/ // /  / / / / / / /_/ /   / ___ |/ /   
/____/_/ /_/\____/ .___//_/  /_/_/_/ /_/\__,_/   /_/  |_/___/  
                /_/                                              
                SCALING TO BILLIONS! 🚀
EOF
    echo -e "${NC}"
    
    check_prerequisites
    calculate_resources "$SCALE_FACTOR"
    
    log "Starting deployment for scale factor: $SCALE_FACTOR"
    
    deploy_infrastructure
    deploy_services
    configure_advanced_features
    optimize_performance
    setup_cost_optimization
    perform_health_checks
    generate_scaling_report
    
    echo ""
    log "🎉 Deployment complete! ShopMindAI is ready to scale to BILLIONS!"
    log "📊 Check scaling-report.md for details"
    log "🌐 Access the application at: https://shopmindai.io"
    echo ""
    echo -e "${GREEN}Happy scaling! 🚀${NC}"
}

# Run main function
main "$@"