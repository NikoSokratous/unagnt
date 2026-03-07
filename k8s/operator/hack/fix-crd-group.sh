#!/usr/bin/env bash
# Fix controller-gen output: set group and CRD name for non-standard api path layout.
set -e
cd "$(dirname "$0")/.."
BASE=config/crd/bases
for f in "$BASE"/_agents.yaml "$BASE"/_workflows.yaml "$BASE"/_policies.yaml; do
  [ -f "$f" ] || continue
  # Fix spec.group
  sed -i.bak 's/group: ""/group: unagnt.io/' "$f"
  # Fix version name (4-space indent = version-level field)
  sed -i.bak 's/^    name: ""$/    name: v1/' "$f"
  # Fix metadata.name (agents., workflows., policies.)
  case "$f" in
    *_agents.yaml)   sed -i.bak 's/name: agents\./name: agents.unagnt.io/' "$f" ;;
    *_workflows.yaml) sed -i.bak 's/name: workflows\./name: workflows.unagnt.io/' "$f" ;;
    *_policies.yaml) sed -i.bak 's/name: policies\./name: policies.unagnt.io/' "$f" ;;
  esac
  rm -f "${f}.bak"
done
