#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_PATH="${REPO_DIR}/bin/nexa-rast"

mkdir -p "${REPO_DIR}/bin"
(cd "${REPO_DIR}" && go build -trimpath -ldflags="-s -w" -o "${BIN_PATH}" ./cmd/nexa-rast)

"${BIN_PATH}" install "$@"
