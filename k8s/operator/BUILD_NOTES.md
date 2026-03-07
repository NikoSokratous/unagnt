# Kubernetes Operator Build Notes

## Status

The Kubernetes operator code is present and uses generated DeepCopy and CRD manifests. Code generation runs automatically when building the operator.

## Required Steps

1. **Install controller-gen** (use v0.14.0 for deterministic output):
   ```bash
   go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0
   ```

2. **Generate DeepCopy methods**:
   ```bash
   make generate-operator
   ```
   Or manually: `cd k8s/operator && controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/v1/..."`

3. **Generate CRD manifests**:
   ```bash
   make generate-crds
   ```
   Generates CRDs to `config/crd/bases/` and copies to `k8s/crds/`. Run this if you change types in `api/v1/types.go`.

4. **Verify codegen is up to date** (for pre-commit):
   ```bash
   make generate-operator-check
   make generate-crds-check
   ```
   CI fails if generated files are stale. Run `make generate-crds` and commit before merging.

5. **Build operator** (runs codegen first):
   ```bash
   make build-operator
   ```
   Produces `bin/unagnt-operator`. No need to run `generate-operator` or `generate-crds` manually.

6. **Registration** in `api/v1/types.go` (already uncommented):
   ```go
   func init() {
       SchemeBuilder.Register(&Agent{}, &AgentList{})
       SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
       SchemeBuilder.Register(&Policy{}, &PolicyList{})
   }
   ```

## Current State

- ✅ Controller logic is complete
- ✅ CRD type definitions are complete
- ✅ Kubernetes dependencies added
- ✅ DeepCopy methods generated (zz_generated.deepcopy.go committed)
- ✅ CRD YAML manifests generated and committed (`config/crd/bases/`, `k8s/crds/`)

## CI and Release Workflow

- **CI check**: The `operator-codegen` job runs on every push/PR. It regenerates code and CRDs, fails if `zz_generated.deepcopy.go` or CRD files differ from the committed version, builds the operator, and validates CRDs with `kubectl apply --dry-run=client`.
- **Before merging**: If you change types in `api/v1/types.go`, run `make generate-crds` and commit the updated `zz_generated.deepcopy.go`, `config/crd/bases/*.yaml`, and `k8s/crds/*.yaml`.
- **Release checklist**: Verify operator codegen passes CI before cutting a release.

## Container image

Build the operator Docker image (uses committed generated code; run `make generate-crds` before building if types changed):

```bash
docker build -f deploy/Dockerfile.operator -t unagnt-operator .
```

## Build Without Operator

To build everything except the operator:

```bash
# Build core packages
go build $(go list ./... | grep -v '/k8s/operator')

# Or use build tags
go build -tags='!operator' ./...
```
