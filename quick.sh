#!/bin/bash
# Stunnel - One command install and run
# Usage: bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/quick.sh)"

set -e

REPO="Lanexus/stunnel"
BINARY="stunnel"
VERSION="v0.5.0"

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
        *)        echo "unknown" ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)             echo "amd64" ;;
    esac
}

# Main
main() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       STUNNEL QUICK START            ║${NC}"
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
    
    # Check if stunnel is installed
    if ! command -v stunnel &> /dev/null; then
        echo -e "${YELLOW}Installing stunnel...${NC}"
        
        # Download binary
        local filename="${BINARY}-${os}-${arch}"
        local url="https://github.com/$REPO/releases/download/$VERSION/$filename"
        
        curl -sL "$url" -o "/tmp/$filename"
        chmod +x "/tmp/$filename"
        
        # Install
        if [ -w "/usr/local/bin" ]; then
            mv "/tmp/$filename" "/usr/local/bin/$BINARY"
        else
            mkdir -p "$HOME/.local/bin"
            mv "/tmp/$filename" "$HOME/.local/bin/$BINARY"
            export PATH="$PATH:$HOME/.local/bin"
        fi
        
        echo -e "${GREEN}✓ Stunnel installed${NC}"
    else
        echo -e "${GREEN}✓ Stunnel already installed${NC}"
    fi
    
    # Generate secret
    local secret=$(stunnel -g)
    
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       YOUR SECRET KEY                ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${YELLOW}$secret${NC}"
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       HOW TO USE                     ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${YELLOW}Server (expose port 3000):${NC}"
    echo -e "    stunnel -l -p 3000 -s $secret"
    echo ""
    echo -e "  ${YELLOW}Client (connect):${NC}"
    echo -e "    stunnel -s $secret"
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       FEATURES                       ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  • ${GREEN}No VPS needed${NC} - works through NAT/Firewall"
    echo -e "  • ${GREEN}No root needed${NC} - runs in user space"
    echo -e "  • ${GREEN}Encrypted${NC} - end-to-end encryption"
    echo -e "  • ${GREEN}Simple${NC} - just one command"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to exit${NC}"
    echo ""
}

main "$@"
