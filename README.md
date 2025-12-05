# XYZ Job Recruitment Platform - Kubernetes Foundation POC

A comprehensive Internal Developer Platform (IDP) demonstrating multi-tenancy, GitOps, service mesh, and self-service infrastructure on Kubernetes.

## Features

- **Multi-tenancy**: Namespace-based isolation for 8 business domains
- **GitOps**: ArgoCD for declarative deployments
- **Service Mesh**: Istio for mTLS, traffic management, observability
- **Policy Engine**: Kyverno for security and compliance
- **Self-Service CRDs**: Tenant, Webservice, Database, Cache, Bucket, Worker, AIService
- **SSO**: Dex with mock users (Okta-ready)
- **Observability**: Prometheus, Grafana, Kiali

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         DEVELOPER PORTAL (Backstage)                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                         GITOPS LAYER (ArgoCD)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                         PLATFORM OPERATORS                                   │
│  Tenant │ Webservice │ Database │ Cache │ DomainIntegration │ AIService     │
├─────────────────────────────────────────────────────────────────────────────┤
│                         SERVICE MESH (Istio)                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                         POLICY ENGINE (Kyverno)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  TENANTS: candidate │ hirer │ sales │ marketing │ ops │ data │ ai │ analytics│
├─────────────────────────────────────────────────────────────────────────────┤
│                         KUBERNETES (k3s)                                     │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Business Domains

| Domain | Purpose | Key Services |
|--------|---------|--------------|
| **candidate** | Job seekers, applications, profiles | candidate-api, profile-service |
| **hirer** | Employers, job postings | job-posting-api, hirer-api |
| **sales** | B2B sales, subscriptions | crm-api, billing-service |
| **marketing** | Campaigns, analytics | campaign-manager, email-service |
| **operations** | Internal tools, support | support-portal, admin-tools |
| **data-service** | Data lake, warehouse, ETL | warehouse-api, datalake-api |
| **ai** | MLOps, model serving, GenAI | prediction-api, job-matching-model |
| **analytics** | BI tools, dashboards | analytics-api |

## Prerequisites

- **macOS/Linux** with 16GB+ RAM (32GB recommended)
- **Docker Desktop**, **Rancher Desktop**, or **Colima**
- **kubectl** (v1.28+)
- **Helm** (v3.x)
- **istioctl** (v1.20+)

### Install Prerequisites (macOS)

```bash
# Install Homebrew if not installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install required tools
brew install kubectl helm istioctl

# Install Rancher Desktop (includes k3s)
brew install --cask rancher-desktop

# OR install Colima (lighter alternative)
brew install colima
colima start --kubernetes --cpu 4 --memory 8
```

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/xyz-company/k8s-hybrid-cloud.git
cd k8s-hybrid-cloud

# 2. Start your Kubernetes cluster
# Rancher Desktop: Enable Kubernetes in settings
# OR Colima: colima start --kubernetes

# 3. Run setup script
chmod +x scripts/setup.sh
./scripts/setup.sh

# 4. Access services (port forwards started automatically)
# ArgoCD:    https://localhost:8080 (admin / see output)
# Grafana:   http://localhost:3000 (admin / admin)
# Dex:       http://localhost:5556
```

## Project Structure

```
.
├── crds/                    # Custom Resource Definitions
│   ├── tenant.yaml          # Multi-tenancy
│   ├── webservice.yaml      # HTTP services
│   ├── database.yaml        # PostgreSQL (CloudNativePG)
│   ├── cache.yaml           # Redis
│   ├── bucket.yaml          # Object storage (S3/MinIO)
│   ├── worker.yaml          # Background job processors
│   ├── scheduled-task.yaml  # Cron jobs
│   ├── routing.yaml         # Ingress/Gateway
│   ├── domain-integration.yaml  # Cross-domain connectivity
│   └── ai-service.yaml      # ML model serving
│
├── platform/                # Platform components
│   ├── argocd/              # GitOps configuration
│   ├── istio/               # Service mesh
│   ├── kyverno/             # Policy engine
│   │   └── policies/        # Security policies
│   ├── cnpg/                # PostgreSQL operator
│   ├── dex/                 # Identity provider
│   └── observability/       # Prometheus, Grafana, Kiali
│
├── operators/               # Custom operators
│   └── tenant-operator/     # Creates namespace, quota, RBAC
│
├── tenants/                 # Tenant configurations
│   ├── candidate/           # Job seekers domain
│   ├── hirer/               # Employers domain
│   ├── sales/               # Sales domain
│   ├── marketing/           # Marketing domain
│   ├── operations/          # Ops domain
│   ├── data-service/        # Data platform domain
│   ├── ai/                  # AI/ML domain
│   └── analytics/           # Analytics domain
│
├── examples/                # Demo applications
│   ├── candidate-api/       # Go HTTP service
│   └── hirer-api/           # Go HTTP service with cross-domain call
│
└── scripts/                 # Setup scripts
    ├── setup.sh             # Install everything
    └── teardown.sh          # Uninstall everything
```

## Custom Resources

### Tenant

```yaml
apiVersion: platform.xyz.com/v1alpha1
kind: Tenant
metadata:
  name: candidate
spec:
  owner: candidate-team
  costCenter: CC-CANDIDATE
  quota:
    cpu: "20"
    memory: "40Gi"
    pods: 200
  allowedIntegrations:
    - hirer
    - ai
```

### Webservice

```yaml
apiVersion: platform.xyz.com/v1alpha1
kind: Webservice
metadata:
  name: candidate-api
  namespace: candidate
spec:
  image: xyz.azurecr.io/candidate-api:v1.0.0
  port: 8080
  replicas:
    min: 2
    max: 10
  resources:
    cpu: 500m
    memory: 512Mi
  routing:
    public: true
    domain: api.xyz.com
    path: /v1/candidates
  highAvailability: true
```

### Database (CloudNativePG)

```yaml
apiVersion: platform.xyz.com/v1alpha1
kind: Database
metadata:
  name: candidate-db
  namespace: candidate
spec:
  type: postgresql
  version: "15"
  size: medium
  storage: 50Gi
  replicas: 3
  backups:
    enabled: true
    schedule: "0 2 * * *"
```

### Worker (Background Jobs)

```yaml
apiVersion: platform.xyz.com/v1alpha1
kind: Worker
metadata:
  name: email-processor
  namespace: marketing
spec:
  image: xyz.azurecr.io/email-processor:v1.0.0
  replicas: 3
  queue:
    type: redis
    name: email-queue
  scaling:
    metric: queueLength
    threshold: 100
```

### AIService

```yaml
apiVersion: ai.xyz.com/v1alpha1
kind: AIService
metadata:
  name: job-matching-model
  namespace: ai
spec:
  model:
    name: job-matching-v2
    registry: s3://xyz-models/
  serving:
    framework: kserve
    gpu: false
  api:
    type: rest
    rateLimit: 200/min
```

## Security

### Policies (Kyverno)

| Policy | Purpose | Action |
|--------|---------|--------|
| `restrict-image-registries` | Only allow trusted registries | Enforce |
| `disallow-privileged-containers` | Block privileged mode | Enforce |
| `require-resource-limits` | Require CPU/memory limits | Enforce |
| `require-run-as-non-root` | No root containers | Audit → Enforce |
| `disallow-latest-tag` | Require explicit versions | Audit |
| `require-probes` | Health checks required | Audit |
| `disallow-nodeport` | Block NodePort services | Enforce |
| `add-default-network-policy` | Auto-create deny ingress | Generate |

### Service Mesh (Istio)

- **mTLS**: Automatic encryption between all services
- **AuthorizationPolicies**: Fine-grained access control
- **NetworkPolicies**: Default deny ingress

### SSO (Dex)

Mock users for development:

| User | Role | Password |
|------|------|----------|
| `admin@xyz.local` | Platform Admin | `admin123` |
| `john.doe@xyz.local` | Candidate Developer | `password123` |
| `jane.smith@xyz.local` | Hirer Developer | `password123` |
| `bob.ml@xyz.local` | AI Developer | `password123` |

To switch to Okta, update `platform/dex/install.yaml` with your Okta connector.

## Common Commands

```bash
# List all tenants
kubectl get tenants

# List webservices across all namespaces
kubectl get webservices -A

# List databases
kubectl get databases -A

# View ArgoCD applications
kubectl get applications -n argocd

# Check Kyverno policy reports
kubectl get policyreports -A

# View Istio service mesh
istioctl dashboard kiali

# Port forward services
kubectl port-forward svc/argocd-server -n argocd 8080:443
kubectl port-forward svc/prometheus-grafana -n monitoring 3000:80

# View ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d
```

## Deploying Applications

### Method 1: Direct kubectl

```bash
# Apply tenant and resources
kubectl apply -f tenants/candidate/

# Deploy example app
kubectl apply -f examples/candidate-api/k8s/
```

### Method 2: ArgoCD (GitOps)

```bash
# Create ArgoCD Application
kubectl apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: candidate-api
  namespace: argocd
spec:
  project: candidate
  source:
    repoURL: https://github.com/xyz-company/k8s-hybrid-cloud
    path: tenants/candidate
    targetRevision: HEAD
  destination:
    server: https://kubernetes.default.svc
    namespace: candidate
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
EOF
```

## Cross-Domain Integration

Example: Hirer API calling Candidate API

```yaml
apiVersion: platform.xyz.com/v1alpha1
kind: DomainIntegration
metadata:
  name: connect-to-candidates
  namespace: hirer
spec:
  targetDomain: candidate
  targetService: candidate-api
  permissions:
    - read:candidates
```

This creates:
- Istio AuthorizationPolicy to allow traffic
- NetworkPolicy exception
- Service discovery entry

## Troubleshooting

### Pods not starting

```bash
# Check events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Check pod status
kubectl describe pod <pod-name> -n <namespace>

# Check Kyverno policy violations
kubectl get policyreports -n <namespace>
```

### Istio issues

```bash
# Check sidecar injection
kubectl get pods -n <namespace> -o jsonpath='{.items[*].spec.containers[*].name}'

# Verify mTLS
istioctl authn tls-check <pod-name>.<namespace>

# Debug Istio proxy
istioctl proxy-status
```

### ArgoCD sync issues

```bash
# Check application status
argocd app get <app-name>

# Force sync
argocd app sync <app-name> --force
```

## Cleanup

```bash
# Remove everything but keep CRDs
./scripts/teardown.sh --keep-crds

# Remove everything including Istio
./scripts/teardown.sh

# Remove everything including k3s cluster
./scripts/teardown.sh --all
```

## Next Steps

1. **Production**: Replace Dex with Okta/Azure AD
2. **Multi-cloud**: Add Rancher or Cluster API
3. **Storage**: Add Longhorn for distributed storage
4. **Security**: Add Trivy for vulnerability scanning
5. **CI/CD**: Integrate with GitHub Actions or Buildkite

## Documentation

| Document | Description |
|----------|-------------|
| [Quick Start](docs/QUICKSTART.md) | 5-minute setup guide |
| [Getting Started](docs/GETTING_STARTED.md) | Detailed installation guide |
| [CRD Reference](docs/CRD_REFERENCE.md) | Complete CRD documentation |

## Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Istio Documentation](https://istio.io/docs/)
- [Kyverno Documentation](https://kyverno.io/docs/)
- [CloudNativePG](https://cloudnative-pg.io/)

## License

MIT
