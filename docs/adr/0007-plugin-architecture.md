# ADR 0007: Plugin Architecture

**Status**: Accepted  
**Date**: 2026-02-26  
**Decision Makers**: Unagnt Core Team

## Context

The runtime needs to support custom tools beyond the built-in set. Requirements:
- Allow users to create custom tools without modifying core
- Support multiple programming languages
- Ensure security and isolation for untrusted code
- Enable hot-reloading for development
- Provide discovery mechanism

## Decision

We will support two plugin types:

1. **Go Plugins** (`.so` files): For trusted, high-performance tools
2. **WASM Plugins** (`.wasm` files): For untrusted, sandboxed tools

### Plugin Interface

All plugins must implement the `tool.Tool` interface:

```go
type Tool interface {
    Name() string
    Version() string
    Description() string
    InputSchema() ([]byte, error)
    Permissions() []Permission
    Execute(ctx context.Context, input json.RawMessage) (map[string]any, error)
}
```

### Discovery Mechanism

Plugins are discovered via `plugin.yaml` manifest files:

```yaml
name: tool_name
version: 1.0.0
type: goplugin  # or wasm
binary: ./tool.so
description: Tool description
author: Author name
permissions:
  - http:external
  - fs:read
```

### Loading Strategy

- **Static**: Load at startup from configured directories
- **Dynamic**: Load on-demand when agent requests tool
- **Hot-reload**: Watch for file changes in development mode

## Alternatives Considered

### 1. Only Go Plugins

**Pros:**
- Native performance
- Simple implementation
- Direct access to Go ecosystem

**Cons:**
- No sandboxing
- Security risk for untrusted code
- Platform-dependent (Linux-only for now)

**Decision**: Rejected. Need sandboxing for user-submitted tools.

### 2. Only WASM

**Pros:**
- Strong sandboxing
- Cross-platform
- Language-agnostic (Rust, Go, C++)

**Cons:**
- Performance overhead
- Limited ecosystem
- Complex memory management

**Decision**: Rejected. Performance matters for trusted tools.

### 3. Subprocess-Based Plugins

**Pros:**
- Process isolation
- Language-agnostic
- Simple IPC

**Cons:**
- High overhead (process spawn per call)
- Complex lifecycle management
- IPC serialization costs

**Decision**: Rejected. Too slow for frequent tool calls.

### 4. HTTP Microservices

**Pros:**
- Maximum isolation
- Language-agnostic
- Easy to scale

**Cons:**
- Network latency on every call
- Complex deployment
- Not suitable for fine-grained tools

**Decision**: Rejected for plugins, but supported via HTTP tool for external services.

## Consequences

### Positive

- **Flexibility**: Two plugin types for different trust/performance needs
- **Security**: WASM sandboxing protects against malicious code
- **Performance**: Go plugins offer zero-overhead native execution
- **Ecosystem**: Users can write plugins in Rust, C++, Go, etc. (via WASM)

### Negative

- **Complexity**: Two plugin systems to maintain
- **Platform limits**: Go plugins only work on Linux/macOS
- **Version coupling**: Go plugins must match runtime's Go version
- **WASM overhead**: 10-30% performance penalty

### Neutral

- **Hot reload**: Useful in development, risky in production
- **Discovery**: Scanning adds startup time
- **Memory**: Each WASM instance requires its own runtime

## Implementation Notes

### Go Plugin Loading

```go
p, err := plugin.Open("./tool.so")
sym, err := p.Lookup("NewTool")
newToolFn := sym.(func() Tool)
tool := newToolFn()
```

### WASM Plugin Loading

```go
rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig())
wasi_snapshot_preview1.Instantiate(ctx, rt)
mod, err := rt.InstantiateWithConfig(ctx, wasmBytes, wazero.NewModuleConfig())
```

### Plugin Lifecycle

1. **Discovery**: Scan directories for `plugin.yaml`
2. **Validation**: Check manifest and binary
3. **Loading**: Load into memory (lazy or eager)
4. **Registration**: Add to tool registry
5. **Execution**: Call via standard `tool.Tool` interface
6. **Unloading**: Remove from memory (hot reload)

## Migration Path

1. **v0.4**: Basic Go and WASM plugin support
2. **v0.5**: Plugin registry and marketplace
3. **v0.6**: Advanced sandboxing and resource limits
4. **v0.7**: Plugin analytics and monitoring

## Testing Strategy

- **Unit tests**: Test each plugin type independently
- **Integration tests**: Test discovery and loading
- **Security tests**: Verify WASM sandboxing
- **Performance tests**: Benchmark overhead

## References

- Go plugins: https://pkg.go.dev/plugin
- WASM: https://github.com/tetratelabs/wazero
- Implementation: `pkg/tool/plugin.go`, `pkg/tool/wasm.go`, `pkg/tool/discovery.go`
