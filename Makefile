.PHONY: build test clean run-unagnt run-unagntd generate-operator generate-operator-check generate-crds generate-crds-check build-operator

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

generate-operator-check: generate-operator
	@git diff --exit-code k8s/operator/api/v1/zz_generated.deepcopy.go || (echo "Operator codegen is stale. Run: make generate-operator and commit." && exit 1)

generate-crds: generate-operator
	cd k8s/operator && controller-gen crd:allowDangerousTypes=true paths="./api/v1/..." output:crd:dir=config/crd/bases
	cd k8s/operator && bash hack/fix-crd-group.sh
	@mkdir -p k8s/crds
	@rm -f k8s/crds/*.yaml
	@cp k8s/operator/config/crd/bases/*.yaml k8s/crds/

generate-crds-check: generate-crds
	@git diff --exit-code k8s/operator/config/crd/bases/ k8s/crds/ || (echo "CRD generation is stale. Run: make generate-crds and commit." && exit 1)

build-operator: generate-crds
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o bin/unagnt-operator ./k8s/operator/cmd

showcase-deploy: build-unagnt
	@echo "Deploying showcase enterprise-compliance-bot..."
	@if command -v helm >/dev/null 2>&1 && command -v kubectl >/dev/null 2>&1; then \
		helm upgrade --install unagnt ./k8s/helm -f ./k8s/helm/values.yaml 2>/dev/null || true; \
		kubectl apply -f ./showcase/enterprise-compliance-bot/k8s/; \
		echo "Showcase deployed. Run: unagnt run --config showcase/enterprise-compliance-bot/agent.yaml --goal '...'"; \
	else \
		echo "Helm/kubectl not found. Use: unagnt run --config showcase/enterprise-compliance-bot/agent.yaml"; \
	fi
