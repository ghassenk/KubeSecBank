# KubeSec Bank

A microservices-based banking application built for learning Kubernetes and container security.

## Architecture

```
┌──────────┐     ┌──────────────┐     ┌─────────────────┐
│  Client  │────▶│  API Gateway │────▶│ Account Service │──▶ PostgreSQL
└──────────┘     │  (Ingress)   │     └─────────────────┘
                 │              │     ┌─────────────────┐
                 │              │────▶│  Auth Service   │──▶ PostgreSQL + Redis
                 │              │     └─────────────────┘
                 │              │     ┌─────────────────────┐
                 │              │────▶│ Transaction Service │──▶ PostgreSQL
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

- JDK 21+
- Maven 3.9+ (or use included Maven Wrapper `./mvnw`)
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
./mvnw spring-boot:run
```

### Building

```bash
# Build all services (Maven)
make build

# Build Docker images
make docker-build
```

### Testing

```bash
# Run unit tests for all services
make test

# Test a single service
cd services/account-service
./mvnw test
```

### Kubernetes Deployment (Kind)

Kind runs a local Kubernetes cluster inside Docker. The cluster container will appear in Docker Desktop.

```bash
# 1. Create a local Kind cluster
make cluster-create

# 2. Build Docker images
make docker-build

# 3. Load images into Kind
make kind-load

# 4. Deploy with Kustomize (dev overlay)
make deploy-dev

# 5. Verify pods are running
kubectl get pods -n kubesec-bank -w

# 6. Port-forward to test a service
kubectl port-forward -n kubesec-bank svc/account-service 8081:8081
curl http://localhost:8081/health
```

Or deploy with Helm instead of Kustomize:

```bash
helm install kubesec-bank deploy/helm/kubesec-bank/
```

To tear down:

```bash
make cluster-delete
```

### Makefile Targets

#### Quick summary of Makefile targets


```bash
make docker-build # Build all Docker images
make cluster-create # Create Kind cluster
make deploy-dev	# Deploy to Kind with dev overlay
make cluster-delete	# Tear down the cluster
```

| Target | Description |
|--------|-------------|
| `make build` | Build all services (Maven, skip tests) |
| `make test` | Run unit tests for all services |
| `make lint` | Run Checkstyle on all services |
| `make docker-build` | Build Docker images for all services |
| `make docker-push` | Push Docker images to registry |
| `make run-local` | Start infrastructure with docker-compose |
| `make stop-local` | Stop docker-compose |
| `make cluster-create` | Create a local Kind cluster |
| `make kind-load` | Load Docker images into Kind cluster |
| `make cluster-delete` | Delete the Kind cluster |
| `make deploy-dev` | Deploy to Kubernetes (dev overlay) |
| `make scan-images` | Run Trivy image scans |
| `make scan-manifests` | Run Trivy on K8s manifests |
| `make clean` | Remove build artifacts |

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
