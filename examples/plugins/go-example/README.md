# Example Go Plugin

This is a complete working example of a Go plugin for Unagnt.

## Building

**Note**: Go plugins only work on Linux and macOS. Windows is not supported by Go's plugin system.

### On Linux/macOS:

```bash
go build -buildmode=plugin -o example_plugin.so example_tool.go
```

### Testing the Plugin

```bash
# Scan for the plugin
unagnt plugin scan --dirs ./examples/plugins

# List loaded plugins
unagnt plugin list --dirs ./examples/plugins
```

## Plugin Structure

The plugin implements the `tool.Tool` interface:

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

## Entry Point

The plugin **must** export a `NewTool() tool.Tool` function:

```go
func NewTool() tool.Tool {
    return &ExampleTool{}
}
```

## Usage in Agent

Add to agent configuration:

```yaml
tools:
  - name: example_plugin
    version: 1.0.0
    source: plugin
```

## Input Schema

```json
{
  "message": "Hello world",
  "uppercase": true
}
```

## Output

```json
{
  "result": "HELLO WORLD",
  "length": 11,
  "uppercase": true,
  "processed": true
}
```

## Permissions

This plugin requires:
- `compute:cpu` - For string processing

## Platform Support

- ✅ Linux (amd64, arm64)
- ✅ macOS (amd64, arm64)
- ❌ Windows (Go plugins not supported)

For Windows, use WASM plugins instead.
