#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "Running tests..."
cd "$ROOT_DIR"
go test -v ./...
