#!/bin/bash
set -e

REPO="lanexus/stunnel"
BINARY="stunnel"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $OS in
    linux)   OS="linux" ;;
    darwin)  OS="darwin" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *)       echo "Unsupported OS: $OS"; exit 1 ;;
esac

case $ARCH in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release
LATEST=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
    echo "Failed to get latest release"
    exit 1
fi

echo "Installing $BINARY $LATEST for $OS/$ARCH..."

# Download URL
FILENAME="${BINARY}-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    FILENAME="${FILENAME}.exe"
fi

URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"

# Download
echo "Downloading $URL..."
curl -sL "$URL" -o "/tmp/$FILENAME"
chmod +x "/tmp/$FILENAME"

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "/tmp/$FILENAME" "$INSTALL_DIR/$BINARY"
else
    echo "Need sudo to install to $INSTALL_DIR"
    sudo mv "/tmp/$FILENAME" "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "  ╔══════════════════════════════════════╗"
echo "  ║     STUNNEL INSTALLED SUCCESSFULLY   ║"
echo "  ╚══════════════════════════════════════╝"
echo ""
echo "  Version: $LATEST"
echo "  Binary:  $INSTALL_DIR/$BINARY"
echo ""
echo "  Usage:"
echo "    stunnel tunnel --local :3000    # Expose via Cloudflare (free)"
echo "    stunnel relay --addr :7000      # Run relay server"
echo "    stunnel --help                  # Show all commands"
echo ""
