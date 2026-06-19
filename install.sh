#!/usr/bin/env sh
# FormaTeX CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/formatexio/cli/main/install.sh | sh

set -e

REPO="formatexio/cli"
BIN_NAME="formatex"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  os="linux" ;;
  Darwin) os="darwin" ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

ASSET="${BIN_NAME}-${os}-${arch}"
RELEASE_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading FormaTeX CLI (${os}/${arch})..."
curl -fsSL "$RELEASE_URL" -o "/tmp/${BIN_NAME}"
chmod +x "/tmp/${BIN_NAME}"

# Try to install to /usr/local/bin, fall back to ~/.local/bin
if [ -w "$INSTALL_DIR" ]; then
  mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
elif command -v sudo >/dev/null 2>&1; then
  sudo mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
  echo "Installed to ${INSTALL_DIR}/${BIN_NAME}"
  echo "Make sure ${INSTALL_DIR} is in your PATH."
fi

echo "FormaTeX CLI installed successfully."
echo "Run 'formatex login' to get started."
