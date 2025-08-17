#!/bin/sh
#
# Nodimus Memory Installer
#
# This script downloads and installs nodimus-memory v1.5.2.
#
# Usage:
#   curl -sSf https://raw.githubusercontent.com/wassmi/nodimus-memory/main/install.sh | sh
#

set -e

# Define the GitHub repository and version
REPO="wassmi/nodimus-memory"
VERSION="v1.5.2"

main() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  # Normalize ARCH name to match GoReleaser's naming convention
  case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64 | aarch64) ARCH="arm64" ;;
  esac

  echo "Downloading Nodimus Memory ${VERSION} for ${OS} ${ARCH}..."

  # Construct the correct download URL
  FILENAME="nodimus-memory_1.5.2_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

  # Create a temporary directory for the download
  TMP_DIR=$(mktemp -d)
  trap 'rm -rf "$TMP_DIR"' EXIT

  # Download and extract the binary
  curl -sSfL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${FILENAME}"
  tar -xzf "${TMP_DIR}/${FILENAME}" -C "${TMP_DIR}"

  # Install the binary
  INSTALL_DIR="/usr/local/bin"
  echo "Installing nodimus-memory to ${INSTALL_DIR}..."
  if [ -w "${INSTALL_DIR}" ]; then
    mv "${TMP_DIR}/nodimus-memory" "${INSTALL_DIR}/nodimus-memory"
  else
    echo "Cannot write to ${INSTALL_DIR}. Trying with sudo."
    sudo mv "${TMP_DIR}/nodimus-memory" "${INSTALL_DIR}/nodimus-memory"
  fi

  chmod +x "${INSTALL_DIR}/nodimus-memory"

  echo "Nodimus Memory installed successfully!"
  echo "Run 'nodimus-memory --help' to get started."
}

main