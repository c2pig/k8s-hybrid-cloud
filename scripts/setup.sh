#!/bin/bash
# XYZ Platform - Setup Script
# This script installs the complete platform on a k3s cluster
#
# Usage: ./scripts/setup.sh
#
# Prerequisites:
#   - k3s/kind/minikube running
#   - kubectl configured
#   - helm 3.x installed
#   - istioctl installed

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     XYZ Job Recruitment Platform - Kubernetes Foundation       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}▶ Checking prerequisites...${NC}"
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}✗ kubectl not found. Please install kubectl.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ kubectl found${NC}"
    
    # Check helm
    if ! command -v helm &> /dev/null; then
        echo -e "${RED}✗ helm not found. Please install helm 3.x.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ helm found${NC}"
    
    # Check istioctl
    if ! command -v istioctl &> /dev/null; then
        echo -e "${YELLOW}⚠ istioctl not found. Installing via brew...${NC}"
        if command -v brew &> /dev/null; then
            brew install istioctl
        else
            echo -e "${RED}✗ Please install istioctl manually: https://istio.io/latest/docs/setup/getting-started/${NC}"
            exit 1
        fi
    fi
    echo -e "${GREEN}✓ istioctl found${NC}"
    
    # Check cluster connection
    if ! kubectl cluster-info &> /dev/null; then
        echo -e "${RED}✗ Cannot connect to Kubernetes cluster${NC}"
        echo -e "${YELLOW}Please ensure your cluster is running:${NC}"
        echo -e "  - Rancher Desktop: Enable Kubernetes in settings"
        echo -e "  - Colima: colima start --kubernetes"
        echo -e "  - k3s: sudo systemctl start k3s"
        exit 1
    fi
    echo -e "${GREEN}✓ Kubernetes cluster connected${NC}"
    echo ""
}

# Function to wait for deployment
wait_for_deployment() {
    local namespace=$1
    local deployment=$2
    local timeout=${3:-300}
    
    echo -e "${YELLOW}  Waiting for $deployment in $namespace...${NC}"
    kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment -n $namespace 2>/dev/null || true
}

# Function to wait for pods
wait_for_pods() {
    local namespace=$1
    local label=$2
    local timeout=${3:-300}
    
    echo -e "${YELLOW}  Waiting for pods with label $label in $namespace...${NC}"
    kubectl wait --for=condition=ready pod -l $label -n $namespace --timeout=${timeout}s 2>/dev/null || true
}

# Install CRDs
install_crds() {
    echo -e "${BLUE}▶ Installing Custom Resource Definitions...${NC}"
    kubectl apply -f "$PROJECT_ROOT/crds/"
    echo -e "${GREEN}✓ CRDs installed${NC}"
    echo ""
}

# Install Istio
install_istio() {
    echo -e "${BLUE}▶ Installing Istio Service Mesh...${NC}"
    
    # Install minimal profile for laptop (saves resources)
    istioctl install --set profile=minimal -y
    
    # Enable sidecar injection for default namespace
    kubectl label namespace default istio-injection=enabled --overwrite 2>/dev/null || true
    
    echo -e "${GREEN}✓ Istio installed${NC}"
    echo ""
}

# Install ArgoCD
install_argocd() {
    echo -e "${BLUE}▶ Installing ArgoCD...${NC}"
    
    # Create namespace
    kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
    
    # Install ArgoCD
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
    
    # Wait for ArgoCD to be ready
    wait_for_deployment argocd argocd-server
    
    # Apply custom configuration
    kubectl apply -f "$PROJECT_ROOT/platform/argocd/"
    
    # Get initial admin password
    echo -e "${YELLOW}  ArgoCD admin password:${NC}"
    kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
    echo ""
    
    echo -e "${GREEN}✓ ArgoCD installed${NC}"
    echo ""
}

# Install Kyverno
install_kyverno() {
    echo -e "${BLUE}▶ Installing Kyverno Policy Engine...${NC}"
    
    # Add Kyverno helm repo
    helm repo add kyverno https://kyverno.github.io/kyverno/ 2>/dev/null || true
    helm repo update
    
    # Install Kyverno
    helm upgrade --install kyverno kyverno/kyverno \
        --namespace kyverno \
        --create-namespace \
        --set replicaCount=1 \
        --wait
    
    # Wait for Kyverno to be ready
    wait_for_deployment kyverno kyverno-admission-controller
    
    # Apply policies
    echo -e "${YELLOW}  Applying security policies...${NC}"
    kubectl apply -f "$PROJECT_ROOT/platform/kyverno/policies/"
    
    echo -e "${GREEN}✓ Kyverno installed${NC}"
    echo ""
}

# Install CloudNativePG
install_cnpg() {
    echo -e "${BLUE}▶ Installing CloudNativePG (PostgreSQL Operator)...${NC}"
    
    # Add CNPG helm repo
    helm repo add cnpg https://cloudnative-pg.github.io/charts 2>/dev/null || true
    helm repo update
    
    # Install CloudNativePG
    helm upgrade --install cnpg cnpg/cloudnative-pg \
        --namespace cnpg-system \
        --create-namespace \
        --wait
    
    echo -e "${GREEN}✓ CloudNativePG installed${NC}"
    echo ""
}

# Install Dex
install_dex() {
    echo -e "${BLUE}▶ Installing Dex (Identity Provider)...${NC}"
    
    kubectl apply -f "$PROJECT_ROOT/platform/dex/"
    
    wait_for_deployment dex dex
    
    echo -e "${GREEN}✓ Dex installed${NC}"
    echo -e "${YELLOW}  Mock users available:${NC}"
    echo -e "    - admin@xyz.local (password: admin123)"
    echo -e "    - john.doe@xyz.local (candidate domain)"
    echo -e "    - jane.smith@xyz.local (hirer domain)"
    echo -e "    - bob.ml@xyz.local (ai domain)"
    echo ""
}

# Install Observability Stack
install_observability() {
    echo -e "${BLUE}▶ Installing Observability Stack...${NC}"
    
    # Create namespace
    kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
    
    # Add Prometheus helm repo
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
    helm repo update
    
    # Install kube-prometheus-stack (Prometheus + Grafana)
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --set prometheus.prometheusSpec.retention=24h \
        --set prometheus.prometheusSpec.resources.requests.memory=256Mi \
        --set prometheus.prometheusSpec.resources.requests.cpu=100m \
        --set grafana.adminPassword=admin \
        --set grafana.resources.requests.memory=128Mi \
        --set grafana.resources.requests.cpu=50m \
        --wait || true
    
    # Install Kiali (Service Mesh visualization)
    kubectl apply -f "$PROJECT_ROOT/platform/observability/" 2>/dev/null || true
    
    echo -e "${GREEN}✓ Observability stack installed${NC}"
    echo -e "${YELLOW}  Grafana credentials: admin / admin${NC}"
    echo ""
}

# Create tenant namespaces
create_tenants() {
    echo -e "${BLUE}▶ Creating Tenant Namespaces...${NC}"
    
    tenants=("candidate" "hirer" "sales" "marketing" "operations" "data-service" "ai" "analytics")
    
    for tenant in "${tenants[@]}"; do
        echo -e "${YELLOW}  Creating tenant: $tenant${NC}"
        
        # Create namespace with labels
        cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: $tenant
  labels:
    platform.xyz.com/tenant: "$tenant"
    istio-injection: enabled
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/warn: restricted
EOF
        
        # Apply tenant configuration if exists
        if [ -d "$PROJECT_ROOT/tenants/$tenant" ]; then
            kubectl apply -f "$PROJECT_ROOT/tenants/$tenant/" 2>/dev/null || true
        fi
    done
    
    echo -e "${GREEN}✓ Tenant namespaces created${NC}"
    echo ""
}

# Setup port forwarding
setup_port_forwards() {
    echo -e "${BLUE}▶ Setting up port forwards...${NC}"
    
    # Kill any existing port forwards
    pkill -f "kubectl port-forward" 2>/dev/null || true
    
    # ArgoCD
    kubectl port-forward svc/argocd-server -n argocd 8080:443 &>/dev/null &
    echo -e "${GREEN}  ✓ ArgoCD: https://localhost:8080${NC}"
    
    # Grafana
    kubectl port-forward svc/prometheus-grafana -n monitoring 3000:80 &>/dev/null &
    echo -e "${GREEN}  ✓ Grafana: http://localhost:3000${NC}"
    
    # Dex
    kubectl port-forward svc/dex -n dex 5556:5556 &>/dev/null &
    echo -e "${GREEN}  ✓ Dex: http://localhost:5556${NC}"
    
    echo ""
}

# Print summary
print_summary() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    Setup Complete!                             ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}Platform Components:${NC}"
    echo -e "  • ArgoCD (GitOps):        https://localhost:8080"
    echo -e "  • Grafana (Monitoring):   http://localhost:3000"
    echo -e "  • Dex (SSO):              http://localhost:5556"
    echo ""
    echo -e "${GREEN}Tenant Namespaces:${NC}"
    echo -e "  • candidate, hirer, sales, marketing"
    echo -e "  • operations, data-service, ai, analytics"
    echo ""
    echo -e "${GREEN}Quick Commands:${NC}"
    echo -e "  # List all tenants"
    echo -e "  kubectl get tenants"
    echo ""
    echo -e "  # List webservices"
    echo -e "  kubectl get webservices -A"
    echo ""
    echo -e "  # Deploy sample app"
    echo -e "  kubectl apply -f tenants/candidate/"
    echo ""
    echo -e "  # View ArgoCD password"
    echo -e "  kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
    echo ""
    echo -e "${YELLOW}To stop port forwards: pkill -f 'kubectl port-forward'${NC}"
    echo ""
}

# Main installation flow
main() {
    cd "$PROJECT_ROOT"
    
    check_prerequisites
    install_crds
    install_istio
    install_kyverno
    install_cnpg
    install_dex
    install_argocd
    install_observability
    create_tenants
    setup_port_forwards
    print_summary
}

# Run main
main "$@"

