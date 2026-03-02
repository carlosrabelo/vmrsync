#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT_DIR"
go clean
rm -rf "${ROOT_DIR:?}/bin"
mkdir -p "${ROOT_DIR:?}/bin"
touch "${ROOT_DIR:?}/bin/.gitkeep"
echo "Cleaned build artifacts (kept bin/.gitkeep)"
