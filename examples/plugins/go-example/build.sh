#!/bin/bash

# Build the example Go plugin
# Note: Go plugins only work on Linux and macOS

set -e

echo "Building Go plugin example..."

# Check OS
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    echo "Error: Go plugins are not supported on Windows"
    echo "Please use WSL or build on Linux/macOS"
    exit 1
fi

# Build plugin
go build -buildmode=plugin -o example_plugin.so example_tool.go

echo "✓ Plugin built successfully: example_plugin.so"
echo ""
echo "Test with:"
echo "  unagnt plugin scan --dirs ."
