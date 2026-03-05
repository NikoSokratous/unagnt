# Air-Gapped Deployment (v2.0)

Deploy Unagnt with no outbound internet. Use local LLMs (e.g. Ollama) only.

## Quick Start

1. Build offline bundle (from machine with network): `./scripts/offline-install.sh bundle`
2. Transfer the generated `.tar.gz` to the air-gapped environment.
3. Unpack and run: `tar -xzf Unagnt-air-gapped-*.tar.gz && cd Unagnt-air-gapped && ./install.sh`
4. Configure `config/Unagnt.yaml` with `llm.default_provider: ollama` and your in-network Ollama URL.

## Scripts

- `scripts/offline-install.sh bundle` – creates tarball with binaries and configs (including `configs/compliance/`).
- Run `install.sh` from inside the unpacked directory to complete install.

## Kubernetes

Load the image into your private registry; set Ollama base URL to an in-cluster or on-prem service. No outbound APIs required for core runtime.
