#!/bin/bash
set -euo pipefail

BINARY_NAME=vmrsync
CMD_PATH="./vmrsync/cmd/${BINARY_NAME}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

COMMIT=$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION_PKG="github.com/carlosrabelo/vmrsync/vmrsync/internal/version"

mkdir -p "$ROOT_DIR/bin"
cd "$ROOT_DIR"
go build -ldflags "-X ${VERSION_PKG}.Version=${COMMIT}" -o "$ROOT_DIR/bin/${BINARY_NAME}" "$CMD_PATH"
echo "Binary ready at: $ROOT_DIR/bin/${BINARY_NAME} (version: ${COMMIT})"
