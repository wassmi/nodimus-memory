#!/bin/sh
#
# Nodimus Memory Installer
#
# This script downloads and installs the latest version of nodimus-memory.
# It tries to determine the OS and architecture and downloads the appropriate binary.
#
# Usage:
#   curl -sSf https://raw.githubusercontent.com/wassmi/nodimus-memory/main/install.sh | sh
#

set -e

# Define the GitHub repository
REPO="wassmi/nodimus-memory"

# Get the latest version from GitHub API
get_latest_version() {
  curl --silent "https://api.github.com/repos/${REPO}/releases/latest" |
  grep '"tag_name":' |
  sed -E 's/.*"([^"]+)".*/\1/'
}

# Determine the operating system and architecture
get_os_arch() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  case "$OS" in
    linux)
      OS="Linux"
      ;;
    darwin)
      OS="Darwin"
      ;;
    *)
      echo "Unsupported OS: $OS"
      exit 1
      ;;
  esac

  case "$ARCH" in
    x86_64)
      ARCH="x86_64"
      ;;
    arm64 | aarch64)
      ARCH="arm64"
      ;;
    *)
      echo "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac
}

main() {
  get_os_arch

  VERSION=$(get_latest_version)
  if [ -z "$VERSION" ]; then
    echo "Could not determine the latest version. Aborting."
    exit 1
  fi

  echo "Downloading Nodimus Memory ${VERSION} for ${OS} ${ARCH}..."

  # Construct the download URL
  FILENAME="nodimus-memory_${OS}_${ARCH}.tar.gz"
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
