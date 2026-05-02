#!/bin/bash
set -euo pipefail

BINARY_NAME=vmrsync
INSTALL_DIR="${HOME}/.local/bin"
COMPLETION_DIR="${HOME}/.local/share/bash-completion/completions"

echo "Removing ${BINARY_NAME} from ${INSTALL_DIR}..."
rm -f "${INSTALL_DIR}/${BINARY_NAME}"
rm -f "${COMPLETION_DIR}/${BINARY_NAME}" 2>/dev/null || true
echo "Uninstallation complete."
