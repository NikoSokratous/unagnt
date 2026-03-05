# Plugin Development Guide

## Overview

The Unagnt supports two types of plugins:

1. **Go Plugins**: Native Go `.so` files for maximum performance
2. **WASM Plugins**: WebAssembly modules for secure sandboxing

## Go Plugins

### Creating a Go Plugin

**Structure:**

```
my-tool-plugin/
├── plugin.yaml       # Manifest
├── tool.go          # Implementation
└── Makefile         # Build script
```

**tool.go:**

```go
package main

import (
    "context"
    "encoding/json"
    
    "github.com/Unagnt/Unagnt/pkg/tool"
)

type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_custom_tool"
}

func (t *MyTool) Version() string {
    return "1.0.0"
}

func (t *MyTool) Description() string {
    return "My custom tool description"
}

func (t *MyTool) InputSchema() ([]byte, error) {
    schema := map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type": "string",
                "description": "Input text",
            },
        },
        "required": []string{"input"},
    }
    return json.Marshal(schema)
}

func (t *MyTool) Permissions() []tool.Permission {
    return []tool.Permission{
        {Scope: "http:external", Required: true},
    }
}

func (t *MyTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
    var params struct {
        Input string `json:"input"`
    }
    
    if err := json.Unmarshal(input, &params); err != nil {
        return nil, err
    }
    
    // Your tool logic here
    result := map[string]any{
        "output": "Processed: " + params.Input,
        "status": "success",
    }
    
    return result, nil
}

// NewTool is the required plugin entry point
func NewTool() tool.Tool {
    return &MyTool{}
}
```

**plugin.yaml:**

```yaml
name: my_custom_tool
version: 1.0.0
type: goplugin
binary: ./my_custom_tool.so
description: Custom tool for specific task
author: Your Name
permissions:
  - http:external
```

**Build:**

```bash
go build -buildmode=plugin -o my_custom_tool.so tool.go
```

### Loading Go Plugins

```go
import "github.com/Unagnt/Unagnt/pkg/tool"

loader := tool.NewPluginLoader()
err := loader.LoadPlugin("./plugins/my_custom_tool.so")
if err != nil {
    log.Fatal(err)
}

// Get plugin
pluginTool, err := loader.GetPlugin("my_custom_tool", "1.0.0")
```

## WASM Plugins

### Creating a WASM Plugin

WASM plugins provide sandboxed execution with limited system access.

**Rust Example (tool.rs):**

```rust
use serde::{Deserialize, Serialize};
use serde_json;

#[derive(Deserialize)]
struct Input {
    text: String,
}

#[derive(Serialize)]
struct Output {
    result: String,
    status: String,
}

#[no_mangle]
pub extern "C" fn execute(input_ptr: *const u8, input_len: usize) -> *const u8 {
    // Read input
    let input_slice = unsafe {
        std::slice::from_raw_parts(input_ptr, input_len)
    };
    
    let input: Input = serde_json::from_slice(input_slice).unwrap();
    
    // Process
    let output = Output {
        result: format!("Processed: {}", input.text),
        status: "success".to_string(),
    };
    
    // Return JSON
    let json = serde_json::to_string(&output).unwrap();
    json.as_ptr()
}
```

**Build:**

```bash
cargo build --target wasm32-wasi --release
cp target/wasm32-wasi/release/tool.wasm my_tool.wasm
```

**plugin.yaml:**

```yaml
name: my_wasm_tool
version: 1.0.0
type: wasm
binary: ./my_tool.wasm
description: WASM-based tool with sandboxing
author: Your Name
permissions:
  - compute:cpu
```

### Loading WASM Plugins

```go
loader := tool.NewWASMLoader()

metadata := tool.WASMToolMetadata{
    Name:    "my_wasm_tool",
    Version: "1.0.0",
    InputSchema: []byte(`{"type":"object","properties":{"text":{"type":"string"}}}`),
    Permissions: []tool.Permission{
        {Scope: "compute:cpu", Required: true},
    },
}

err := loader.LoadWASM(ctx, "./plugins/my_tool.wasm", metadata)
```

## Plugin Discovery

### Auto-Discovery

The runtime can automatically discover plugins in configured directories:

```go
discovery := tool.NewPluginDiscovery([]string{
    "./plugins",
    "/usr/local/share/Unagnt/plugins",
})

manifests, err := discovery.ScanPlugins()
for _, m := range manifests {
    fmt.Printf("Found: %s (%s)\n", m.Name, m.Type)
    discovery.LoadFromManifest(m)
}
```

### Hot Reload

Enable automatic reloading when plugin files change:

```go
err := discovery.WatchForChanges("./plugins", 5*time.Second)
```

## CLI Plugin Management

### Scan for Plugins

```bash
unagnt plugin scan --dirs ./plugins --verbose
```

### List Loaded Plugins

```bash
unagnt plugin list
```

## Manifest Format

The `plugin.yaml` manifest describes plugin metadata:

```yaml
name: tool_name          # Tool identifier
version: 1.0.0           # Semantic version
type: goplugin           # "goplugin" or "wasm"
binary: ./tool.so        # Path to binary (relative to manifest)
description: Tool desc   # Human-readable description
author: Your Name        # Plugin author
permissions:             # Required permissions
  - http:external
  - fs:read
  - db:write
```

## Security Considerations

### Go Plugins

- **Full system access**: Go plugins run in the same process
- **Trust requirement**: Only load plugins from trusted sources
- **Version pinning**: Pin plugin versions in production
- **Code review**: Audit plugin source before deployment

### WASM Plugins

- **Sandboxed execution**: Limited system access via WASI
- **Memory isolation**: Plugins cannot access runtime memory
- **Capability-based**: Explicitly grant file, network, env access
- **Performance overhead**: ~10-30% slower than native

**WASM security model:**

```go
// Plugins can only access explicitly granted capabilities
config := wazero.NewModuleConfig().
    WithStdout(os.Stdout).              // Allow stdout
    WithStderr(os.Stderr).              // Allow stderr
    WithFS(fsys)                        // Mount filesystem
    WithEnv("API_KEY", apiKey)          // Expose env var
```

## Performance

### Benchmark Results

| Plugin Type | Overhead | Isolation | Use Case |
|-------------|----------|-----------|----------|
| Go Plugin   | ~0%      | None      | Trusted, high-performance tools |
| WASM        | ~10-30%  | Strong    | Untrusted, user-submitted tools |

### Optimization Tips

1. **Minimize data copying**: Pass large data by reference
2. **Batch operations**: Group multiple tool calls
3. **Connection pooling**: Reuse HTTP clients, DB connections
4. **Lazy loading**: Only load plugins when needed

## Testing

### Unit Testing Go Plugins

```go
func TestMyTool(t *testing.T) {
    tool := &MyTool{}
    
    input := json.RawMessage(`{"input":"test"}`)
    output, err := tool.Execute(context.Background(), input)
    
    assert.NoError(t, err)
    assert.Equal(t, "success", output["status"])
}
```

### Integration Testing

```bash
# Build plugin
go build -buildmode=plugin -o test.so

# Test loading
unagnt plugin scan --dirs .
```

## Distribution

### Plugin Registry (Planned)

Future versions will support a plugin registry:

```bash
# Publish plugin
unagnt plugin publish ./my-tool --registry https://plugins.Unagnt.dev

# Install plugin
unagnt plugin install my-tool@1.0.0
```

### Versioning

Follow semantic versioning:
- **Major**: Breaking API changes
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes

## Examples

### File Processor Plugin

```go
type FileProcessorTool struct{}

func (t *FileProcessorTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
    var params struct {
        FilePath string `json:"file_path"`
        Action   string `json:"action"`
    }
    
    if err := json.Unmarshal(input, &params); err != nil {
        return nil, err
    }
    
    // Check permissions via context
    // policy.CheckPermission(ctx, "fs:read", params.FilePath)
    
    data, err := os.ReadFile(params.FilePath)
    if err != nil {
        return nil, err
    }
    
    return map[string]any{
        "content": string(data),
        "size":    len(data),
    }, nil
}
```

### API Client Plugin

```go
type APIClientTool struct {
    httpClient *http.Client
}

func (t *APIClientTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
    var params struct {
        Endpoint string                 `json:"endpoint"`
        Method   string                 `json:"method"`
        Body     map[string]interface{} `json:"body"`
    }
    
    json.Unmarshal(input, &params)
    
    // Make HTTP request with timeout
    req, _ := http.NewRequestWithContext(ctx, params.Method, params.Endpoint, nil)
    resp, err := t.httpClient.Do(req)
    
    // ... handle response ...
    
    return result, nil
}
```

## Troubleshooting

### Plugin Not Loading

```bash
# Check plugin validity
file my_tool.so

# Expected output for Go plugins:
# my_tool.so: ELF 64-bit LSB shared object
```

### Symbol Not Found

Ensure `NewTool` function is exported (uppercase):

```go
func NewTool() tool.Tool {  // ✓ Correct
func newTool() tool.Tool {  // ✗ Wrong (lowercase)
```

### WASM Compilation Errors

```bash
# Install WASI target
rustup target add wasm32-wasi

# Build with correct target
cargo build --target wasm32-wasi --release
```

## Future Enhancements

- Plugin marketplace
- Automatic dependency resolution
- Plugin sandboxing for Go plugins
- Remote plugin loading
- Plugin analytics and monitoring
