#!/bin/bash

# Build the WASM plugin example

set -e

echo "Building WASM plugin example..."

# Check if Rust is installed
if ! command -v cargo &> /dev/null; then
    echo "Error: Rust is not installed"
    echo "Install from: https://rustup.rs/"
    exit 1
fi

# Check if wasm32-wasi target is installed
if ! rustup target list | grep -q "wasm32-wasi (installed)"; then
    echo "Installing wasm32-wasi target..."
    rustup target add wasm32-wasi
fi

# Build
cargo build --target wasm32-wasi --release

# Copy to current directory
cp target/wasm32-wasi/release/example_wasm_tool.wasm example.wasm

echo "✓ WASM plugin built successfully: example.wasm"
echo ""
echo "Test with:"
echo "  unagnt plugin scan --dirs ."
