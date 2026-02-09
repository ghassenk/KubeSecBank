# KubeSec Bank

A microservices-based banking application built for learning Kubernetes and container security.

## Architecture

```
┌──────────┐     ┌──────────────┐     ┌─────────────────┐
│  Client   │────▶│  API Gateway │────▶│ Account Service │──▶ PostgreSQL
└──────────┘     │  (Ingress)   │     └─────────────────┘
                 │              │     ┌─────────────────┐
                 │              │────▶│  Auth Service    │──▶ PostgreSQL + Redis
                 │              │     └─────────────────┘
                 │              │     ┌─────────────────────┐
                 │              │────▶│ Transaction Service  │──▶ PostgreSQL
                 └──────────────┘     └─────────────────────┘
                                              │
                                              ▼
                                          NATS (async)
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| Account Service | 8081 | User registration, KYC, account management |
| Auth Service | 8082 | Authentication, JWT, MFA, session management |
| Transaction Service | 8083 | Transfers, transaction history |

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- kubectl
- Kind or Minikube (for local K8s)
- Helm 3

### Local Development

```bash
# Start infrastructure (PostgreSQL, Redis, NATS)
docker-compose up -d

# Run a service locally
cd services/account-service
go run cmd/main.go
```

### Kubernetes Deployment

```bash
# Create a local cluster
kind create cluster --name kubesec-bank

# Deploy with Helm
helm install kubesec-bank deploy/helm/kubesec-bank/

# Or with plain manifests
kubectl apply -k deploy/kubernetes/overlays/dev/
```

## Security Focus Areas

- [ ] Network Policies (pod-to-pod isolation)
- [ ] Pod Security Standards (restricted profile)
- [ ] RBAC (least-privilege service accounts)
- [ ] Secrets management (external secrets / Vault)
- [ ] mTLS between services
- [ ] Image scanning in CI (Trivy)
- [ ] Admission control (OPA/Kyverno)
- [ ] Audit logging
- [ ] Runtime monitoring

## Project Structure

```
KubeSec/
├── services/
│   ├── account-service/      # Account management
│   ├── auth-service/         # Authentication & authorization
│   └── transaction-service/  # Financial transactions
├── deploy/
│   ├── kubernetes/           # Raw K8s manifests
│   │   ├── base/             # Base resources
│   │   └── overlays/         # Kustomize overlays (dev/prod)
│   └── helm/                 # Helm chart
├── scripts/                  # Utility scripts
├── docs/                     # Documentation
└── .github/workflows/        # CI/CD pipelines
```

## License

MIT
