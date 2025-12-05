#!/bin/bash
# XYZ Platform - Teardown Script
# This script removes all platform components and tenant resources
#
# Usage: ./scripts/teardown.sh
#
# Options:
#   --keep-crds    Keep CRDs installed
#   --keep-istio   Keep Istio installed
#   --all          Remove everything including k3s cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Parse arguments
KEEP_CRDS=false
KEEP_ISTIO=false
REMOVE_ALL=false

for arg in "$@"; do
    case $arg in
        --keep-crds)
            KEEP_CRDS=true
            ;;
        --keep-istio)
            KEEP_ISTIO=true
            ;;
        --all)
            REMOVE_ALL=true
            ;;
    esac
done

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     XYZ Platform - Teardown                                    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Confirm
read -p "Are you sure you want to remove the platform? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Aborted.${NC}"
    exit 1
fi

# Stop port forwards
echo -e "${YELLOW}▶ Stopping port forwards...${NC}"
pkill -f "kubectl port-forward" 2>/dev/null || true
echo -e "${GREEN}✓ Port forwards stopped${NC}"

# Delete tenant namespaces
echo -e "${YELLOW}▶ Deleting tenant namespaces...${NC}"
tenants=("candidate" "hirer" "sales" "marketing" "operations" "data-service" "ai" "analytics")
for tenant in "${tenants[@]}"; do
    kubectl delete namespace $tenant --ignore-not-found=true 2>/dev/null || true
done
echo -e "${GREEN}✓ Tenant namespaces deleted${NC}"

# Delete observability
echo -e "${YELLOW}▶ Removing observability stack...${NC}"
helm uninstall prometheus -n monitoring 2>/dev/null || true
kubectl delete namespace monitoring --ignore-not-found=true 2>/dev/null || true
echo -e "${GREEN}✓ Observability removed${NC}"

# Delete Dex
echo -e "${YELLOW}▶ Removing Dex...${NC}"
kubectl delete -f "$PROJECT_ROOT/platform/dex/" 2>/dev/null || true
kubectl delete namespace dex --ignore-not-found=true 2>/dev/null || true
echo -e "${GREEN}✓ Dex removed${NC}"

# Delete CloudNativePG
echo -e "${YELLOW}▶ Removing CloudNativePG...${NC}"
helm uninstall cnpg -n cnpg-system 2>/dev/null || true
kubectl delete namespace cnpg-system --ignore-not-found=true 2>/dev/null || true
echo -e "${GREEN}✓ CloudNativePG removed${NC}"

# Delete Kyverno
echo -e "${YELLOW}▶ Removing Kyverno...${NC}"
kubectl delete -f "$PROJECT_ROOT/platform/kyverno/policies/" 2>/dev/null || true
helm uninstall kyverno -n kyverno 2>/dev/null || true
kubectl delete namespace kyverno --ignore-not-found=true 2>/dev/null || true
echo -e "${GREEN}✓ Kyverno removed${NC}"

# Delete ArgoCD
echo -e "${YELLOW}▶ Removing ArgoCD...${NC}"
kubectl delete -f "$PROJECT_ROOT/platform/argocd/" 2>/dev/null || true
kubectl delete -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml 2>/dev/null || true
kubectl delete namespace argocd --ignore-not-found=true 2>/dev/null || true
echo -e "${GREEN}✓ ArgoCD removed${NC}"

# Delete Istio (optional)
if [ "$KEEP_ISTIO" = false ]; then
    echo -e "${YELLOW}▶ Removing Istio...${NC}"
    istioctl uninstall --purge -y 2>/dev/null || true
    kubectl delete namespace istio-system --ignore-not-found=true 2>/dev/null || true
    echo -e "${GREEN}✓ Istio removed${NC}"
else
    echo -e "${YELLOW}⚠ Keeping Istio (--keep-istio flag)${NC}"
fi

# Delete CRDs (optional)
if [ "$KEEP_CRDS" = false ]; then
    echo -e "${YELLOW}▶ Removing CRDs...${NC}"
    kubectl delete -f "$PROJECT_ROOT/crds/" 2>/dev/null || true
    echo -e "${GREEN}✓ CRDs removed${NC}"
else
    echo -e "${YELLOW}⚠ Keeping CRDs (--keep-crds flag)${NC}"
fi

# Remove k3s cluster (if --all)
if [ "$REMOVE_ALL" = true ]; then
    echo -e "${YELLOW}▶ Removing k3s cluster...${NC}"
    if command -v k3s-uninstall.sh &> /dev/null; then
        sudo k3s-uninstall.sh
    fi
    echo -e "${GREEN}✓ k3s removed${NC}"
fi

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    Teardown Complete!                          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}To reinstall: ./scripts/setup.sh${NC}"
