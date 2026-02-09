.PHONY: all build test lint clean docker-build docker-push run-local

SERVICES := account-service auth-service transaction-service
REGISTRY ?= ghcr.io/ghassenk/kubesecbank
TAG ?= latest

all: lint test build

## Build
build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc && go build -o ../../bin/$$svc ./cmd/main.go && cd ../..; \
	done

## Test
test:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test ./... -v -coverprofile=../../coverage-$$svc.txt && cd ../..; \
	done

## Lint
lint:
	@for svc in $(SERVICES); do \
		echo "Linting $$svc..."; \
		cd services/$$svc && golangci-lint run ./... && cd ../..; \
	done

## Docker
docker-build:
	@for svc in $(SERVICES); do \
		echo "Building Docker image for $$svc..."; \
		docker build -t $(REGISTRY)/$$svc:$(TAG) -f services/$$svc/Dockerfile services/$$svc; \
	done

docker-push:
	@for svc in $(SERVICES); do \
		docker push $(REGISTRY)/$$svc:$(TAG); \
	done

## Local development
run-local:
	docker-compose up -d

stop-local:
	docker-compose down

## Kubernetes (Kind)
cluster-create:
	kind create cluster --name kubesec-bank --config scripts/kind-config.yaml

cluster-delete:
	kind delete cluster --name kubesec-bank

deploy-dev:
	kubectl apply -k deploy/kubernetes/overlays/dev/

## Security scanning
scan-images:
	@for svc in $(SERVICES); do \
		echo "Scanning $$svc..."; \
		trivy image $(REGISTRY)/$$svc:$(TAG); \
	done

scan-manifests:
	trivy config deploy/kubernetes/

## Clean
clean:
	rm -rf bin/ coverage-*.txt
