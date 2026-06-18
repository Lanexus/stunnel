#!/bin/bash
# Stunnel installer - works like gsocket
# Usage: bash <(curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh)

set -e

REPO="Lanexus/stunnel"
BINARY="stunnel"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Detect OS
detect_os() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $os in
        linux*)   echo "linux" ;;
        darwin*)  echo "darwin" ;;
        mingw*|msys*|cygwin*) echo "windows" ;;
        *)        echo "unknown" ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv7l|armhf)  echo "arm" ;;
        *)             echo "amd64" ;;
    esac
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download file
download() {
    local url="$1"
    local output="$2"
    
    if command_exists curl; then
        curl -sL --connect-timeout 10 --max-time 60 "$url" -o "$output"
    elif command_exists wget; then
        wget --timeout=10 -q "$url" -O "$output"
    elif command_exists fetch; then
        fetch -o "$output" "$url"
    else
        echo -e "${RED}Error: No download tool found (curl, wget, fetch)${NC}"
        exit 1
    fi
}

# Get latest release version
get_latest_version() {
    local version
    # Try GitHub API first
    version=$(curl -s --connect-timeout 5 "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        # Fallback to hardcoded version
        version="v0.2.0"
    fi
    
    echo "$version"
}

# Main installation
main() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       STUNNEL INSTALLER              ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    
    # Detect system
    local os=$(detect_os)
    local arch=$(detect_arch)
    
    if [ "$os" = "unknown" ]; then
        echo -e "${RED}Error: Unsupported OS${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}Detected: $os/$arch${NC}"
    
    # Get version
    local version=$(get_latest_version)
    echo -e "${YELLOW}Version: $version${NC}"
    
    # Build download URL
    local filename="${BINARY}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        filename="${filename}.exe"
    fi
    
    local url="https://github.com/$REPO/releases/download/$version/$filename"
    echo -e "${YELLOW}Downloading: $url${NC}"
    
    # Download to temp
    local tmpfile="/tmp/$filename"
    download "$url" "$tmpfile"
    
    if [ ! -f "$tmpfile" ] || [ ! -s "$tmpfile" ]; then
        echo -e "${RED}Error: Download failed${NC}"
        echo -e "${YELLOW}Try manually: wget $url -O /usr/local/bin/$BINARY${NC}"
        exit 1
    fi
    
    chmod +x "$tmpfile"
    
    # Install location
    local installdir="/usr/local/bin"
    if [ ! -w "$installdir" ]; then
        installdir="$HOME/.local/bin"
        mkdir -p "$installdir"
    fi
    
    # Move binary
    mv "$tmpfile" "$installdir/$BINARY"
    
    # Add to PATH if needed
    if [[ ":$PATH:" != *":$installdir:"* ]]; then
        export PATH="$PATH:$installdir"
        # Add to shell profile
        for profile in ~/.bashrc ~/.zshrc ~/.profile; do
            if [ -f "$profile" ]; then
                echo "export PATH=\"\$PATH:$installdir\"" >> "$profile"
            fi
        done
    fi
    
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     ✓ STUNNEL INSTALLED!             ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  Binary:  ${GREEN}$installdir/$BINARY${NC}"
    echo ""
    echo -e "  ${YELLOW}Quick Start:${NC}"
    echo -e "    stunnel tunnel --local :3000     # Expose via Cloudflare (free)"
    echo -e "    stunnel relay --addr :7000       # Run relay server"
    echo -e "    stunnel --help                   # Show all commands"
    echo ""
}

main "$@"
