.PHONY: build test clean run-unagnt run-unagntd generate-operator

BINARY_UNAGNT := bin/unagnt
BINARY_UNAGNTD := bin/unagntd
GO := go
GOFLAGS := -v

all: build

build: build-unagnt build-unagntd

build-unagnt:
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_UNAGNT) ./cmd/unagnt

build-unagntd:
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_UNAGNTD) ./cmd/unagntd

test:
	$(GO) test ./... -race -coverprofile=coverage.out

test-short:
	$(GO) test ./... -short

coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

run-unagnt: build-unagnt
	./$(BINARY_UNAGNT) $(ARGS)

run-unagntd: build-unagntd
	./$(BINARY_UNAGNTD) $(ARGS)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint: fmt vet

generate-operator:
	cd k8s/operator && controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/v1/..."

showcase-deploy: build-unagnt
	@echo "Deploying showcase enterprise-compliance-bot..."
	@if command -v helm >/dev/null 2>&1 && command -v kubectl >/dev/null 2>&1; then \
		helm upgrade --install unagnt ./k8s/helm -f ./k8s/helm/values.yaml 2>/dev/null || true; \
		kubectl apply -f ./showcase/enterprise-compliance-bot/k8s/; \
		echo "Showcase deployed. Run: unagnt run --config showcase/enterprise-compliance-bot/agent.yaml --goal '...'"; \
	else \
		echo "Helm/kubectl not found. Use: unagnt run --config showcase/enterprise-compliance-bot/agent.yaml"; \
	fi
