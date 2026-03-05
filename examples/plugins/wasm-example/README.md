# Example WASM Plugin

This demonstrates a WebAssembly-based plugin for Unagnt with sandboxed execution.

## Building

### Prerequisites

- Rust toolchain with `wasm32-wasi` target

```bash
rustup target add wasm32-wasi
```

### Build

```bash
cargo build --target wasm32-wasi --release
cp target/wasm32-wasi/release/example_tool.wasm .
```

## Plugin Structure

WASM plugins must export an `execute` function that:
1. Receives a pointer to JSON input
2. Processes the data
3. Returns a pointer to JSON output

## Testing

```bash
# Scan for WASM plugins
unagnt plugin scan --dirs ./examples/plugins --verbose

# The runtime will load and execute via Wazero
```

## Sandboxing

WASM plugins run in a secure sandbox with:
- ✅ Memory isolation
- ✅ No direct file system access (unless granted via WASI)
- ✅ No network access (unless granted)
- ✅ Controlled execution time
- ✅ Resource limits

## Input Schema

```json
{
  "text": "Hello from WASM"
}
```

## Output

```json
{
  "result": "Processed: Hello from WASM",
  "status": "success",
  "sandbox": "wasm"
}
```

## Permissions

This plugin requires:
- `compute:cpu` - For processing

## Platform Support

- ✅ Linux (all architectures)
- ✅ macOS (Intel & Apple Silicon)
- ✅ Windows (all architectures)
- ✅ Any platform with Wazero support

WASM plugins are **cross-platform** and work everywhere!

## Security Benefits

Unlike Go plugins, WASM plugins:
- Cannot crash the runtime
- Cannot access runtime memory
- Cannot make arbitrary system calls
- Are safe to load from untrusted sources (with proper sandboxing)

## Performance

- Overhead: ~10-30% compared to native Go
- Cold start: ~1-5ms
- Warm execution: Near-native speed
