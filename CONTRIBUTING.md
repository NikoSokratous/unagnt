# Contributing to Agent Runtime

Thank you for your interest in contributing to Agent Runtime!

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting PRs](#submitting-prs)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Be respectful, professional, and inclusive.

## Getting Started

### Prerequisites

- Go 1.22+
- Git
- Make
- Docker (for integration tests)
- Python 3.8+ (for Python SDK development)

### First-Time Setup

```bash
# Clone repository
git clone https://github.com/NikoSokratous/unagnt.git
cd Unagnt

# Install dependencies
go mod download

# Build binaries
make build

# Run tests
make test
```

## Development Setup

### Development Environment

```bash
# Install linters
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install pre-commit hooks (optional)
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### IDE Setup

Recommended: VS Code or GoLand

**VS Code Extensions:**
- Go (golang.go)
- YAML (redhat.vscode-yaml)
- REST Client (humao.rest-client)

### Running Locally

```bash
# Terminal 1: Start unagntd
make run-unagntd

# Terminal 2: Test with unagnt
export OPENAI_API_KEY=sk-...
make run-unagnt
```

## Making Changes

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation
- `refactor/description` - Code refactoring
- `test/description` - Test improvements

### Commit Messages

Follow conventional commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `test` - Tests
- `refactor` - Code refactoring
- `perf` - Performance improvement
- `chore` - Maintenance

**Examples:**
```
feat(tool): add github_api tool for repository operations

fix(policy): handle nil input in CEL evaluation

docs(guide): add tool development examples

test(runtime): increase state machine test coverage to 90%
```

## Testing

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./pkg/runtime -v

# With coverage
make coverage

# Integration tests (requires Docker)
go test ./test/integration -v

# Skip integration tests
go test ./... -short
```

### Writing Tests

**Unit Test Template:**

```go
func TestFeatureName(t *testing.T) {
    // Arrange
    input := "test"
    
    // Act
    result := doSomething(input)
    
    // Assert
    if result != "expected" {
        t.Errorf("got %v, want %v", result, "expected")
    }
}
```

**Table-Driven Tests:**

```go
func TestMultipleScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "a", "A"},
        {"case2", "b", "B"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := transform(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Test Coverage Requirements

- **Core packages** (runtime, policy, tool): 70%+ coverage
- **Critical paths**: 90%+ coverage
- **New features**: Must include tests

## Code Style

### Go Conventions

Follow [Effective Go](https://go.dev/doc/effective_go):

- Use `gofmt` (automatic via CI)
- Use meaningful variable names
- Keep functions small (<50 lines)
- Document exported functions
- Handle all errors (no `_` without comment)

### Linting

```bash
golangci-lint run
```

Fix all linter warnings before submitting.

### Documentation

- Add godoc comments for exported types/functions
- Update README.md if adding features
- Create ADRs for architectural decisions

## Submitting PRs

### Before Submitting

1. **Run tests**: `make test`
2. **Run linters**: `golangci-lint run`
3. **Update docs**: If changing APIs or adding features
4. **Rebase on main**: `git rebase origin/main`

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] No linter warnings
```

### PR Review Process

1. **Automated Checks**: CI must pass (tests, lint, build)
2. **Code Review**: At least one maintainer approval
3. **Documentation**: Verify docs are updated
4. **Testing**: Adequate test coverage

### Review Timelines

- Small PRs (< 100 lines): 1-2 days
- Medium PRs (100-500 lines): 3-5 days
- Large PRs (> 500 lines): Consider splitting

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **Major** (v1.0.0): Breaking changes
- **Minor** (v0.3.0): New features, backward compatible
- **Patch** (v0.2.1): Bug fixes

### Release Checklist

1. Update version in code
2. Update CHANGELOG.md
3. Run full test suite
4. Tag release: `git tag v0.3.0`
5. Push tag: `git push origin v0.3.0`
6. GitHub Actions creates release
7. Announce in Discord/Twitter

## Project Structure

```
Unagnt/
├── cmd/           # CLI and daemon entrypoints
├── pkg/           # Public APIs
├── internal/      # Private implementation
├── sdk/           # Client libraries (Go, Python)
├── test/          # Integration tests
├── docs/          # Documentation
├── examples/      # Example applications
├── deploy/        # Deployment configs
└── .github/       # CI/CD workflows
```

## Areas Needing Help

### High Priority

- [ ] Additional tool implementations (file ops, git, browser)
- [ ] Advanced semantic memory adapters (Qdrant, Weaviate)
- [ ] Web UI for run visualization
- [ ] Additional LLM providers (Cohere, AI21, etc.)

### Medium Priority

- [ ] Performance optimizations
- [ ] More comprehensive tests
- [ ] Example applications
- [ ] Documentation improvements

### Good First Issues

Look for `good-first-issue` label on GitHub.

## Communication

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas
- **Discord**: Real-time chat (coming soon)
- **Email**: security@Unagnt.dev (security issues only)

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md
- GitHub releases
- Project README

## Questions?

- Check [existing issues](https://github.com/NikoSokratous/unagnt/issues)
- Read the [documentation](../guides/)
- Ask in GitHub Discussions

Thank you for contributing to Agent Runtime!
