#!/bin/bash
set -euo pipefail

BINARY_NAME=vmrsync
INSTALL_DIR="${HOME}/.local/bin"
COMPLETION_DIR="${HOME}/.local/share/bash-completion/completions"

echo "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
mkdir -p "${INSTALL_DIR}"
cp "bin/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
mkdir -p "${COMPLETION_DIR}"
cp "vmrsync.bash-completion" "${COMPLETION_DIR}/${BINARY_NAME}" 2>/dev/null || true
echo "Installation complete."
