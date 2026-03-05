#!/usr/bin/env bash
# Air-gapped deployment: create an offline bundle or run install from it.
# Usage: ./scripts/offline-install.sh bundle | install

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VERSION="${VERSION:-0.0.0}"
BUNDLE_NAME="agentruntime-air-gapped-${VERSION}"

bundle() {
  echo "Building binaries..."
  cd "$REPO_ROOT"
  make build 2>/dev/null || true
  if [[ ! -f "$REPO_ROOT/bin/agentd" ]]; then
    echo "Run 'make build' first to produce bin/unagntd and bin/unagnt"
    exit 1
  fi

  STAGING="$REPO_ROOT/_airgap/$BUNDLE_NAME"
  rm -rf "$STAGING"
  mkdir -p "$STAGING"/{bin,config,configs/compliance}

  cp "$REPO_ROOT/bin/unagntd" "$REPO_ROOT/bin/unagnt" "$STAGING/bin/" 2>/dev/null || true
  cp -r "$REPO_ROOT/configs/compliance/"* "$STAGING/configs/compliance/" 2>/dev/null || true
  [[ -d "$REPO_ROOT/configs/examples" ]] && cp -r "$REPO_ROOT/configs/examples" "$STAGING/configs/" || true

  cat > "$STAGING/install.sh" << 'INSTALL'
#!/usr/bin/env bash
set -e
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "$DIR/config"
echo "Installing AgentRuntime into $DIR"
echo "Binaries: $DIR/bin/unagntd, $DIR/bin/unagnt"
echo "Configs:  $DIR/configs/ and $DIR/config/"
echo "Run: $DIR/bin/unagntd or add $DIR/bin to PATH and run unagnt"
INSTALL
  chmod +x "$STAGING/install.sh"

  TARBALL="$REPO_ROOT/${BUNDLE_NAME}.tar.gz"
  (cd "$REPO_ROOT/_airgap" && tar -czf "$TARBALL" "$BUNDLE_NAME")
  rm -rf "$STAGING"
  echo "Created $TARBALL"
}

install_from_bundle() {
  BUNDLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  if [[ ! -f "$BUNDLE_DIR/bin/agentd" ]]; then
    echo "Run this script from inside the unpacked agentruntime-air-gapped-* directory."
    exit 1
  fi
  "$BUNDLE_DIR/install.sh"
}

case "${1:-}" in
  bundle)  bundle ;;
  install) install_from_bundle ;;
  *)      echo "Usage: $0 bundle | install"; exit 1 ;;
esac
