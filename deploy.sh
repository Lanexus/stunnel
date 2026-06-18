#!/bin/bash
# Stunnel Deploy Script
# Usage:
#   Install: bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/deploy.sh)"
#   Access:  S="secret" bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/deploy.sh)"
#   Uninstall: UNDO=1 bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/deploy.sh)"

set -e

REPO="Lanexus/stunnel"
VERSION="v0.6.0"
BINARY_NAME="stunnel"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
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

# Generate random secret
generate_secret() {
    openssl rand -base64 16 2>/dev/null || head -c 24 /dev/urandom | base64 | tr -d '\n' | head -c 16
}

# Get public IP
get_public_ip() {
    curl -s --connect-timeout 5 ifconfig.me 2>/dev/null || \
    curl -s --connect-timeout 5 ipinfo.io/ip 2>/dev/null || \
    echo "0.0.0.0"
}

# Download binary
download_binary() {
    local os=$1
    local arch=$2
    local dest=$3
    
    local filename="${BINARY_NAME}-${os}-${arch}"
    local url="https://github.com/${REPO}/releases/download/${VERSION}/${filename}"
    
    echo -e "${YELLOW}Downloading ${filename}...${NC}"
    
    if command -v curl &> /dev/null; then
        curl -sL "$url" -o "$dest"
    elif command -v wget &> /dev/null; then
        wget -q "$url" -O "$dest"
    else
        echo -e "${RED}Error: curl or wget required${NC}"
        exit 1
    fi
    
    chmod +x "$dest"
}

# Install systemd service
install_service() {
    local secret=$1
    local port=$2
    local binary_path=$3
    
    cat > /etc/systemd/system/stunnel.service << EOF
[Unit]
Description=Stunnel Server
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=5
ExecStart=${binary_path} server -s ${secret} -p ${port}
Environment=SHELL=/bin/bash

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable stunnel
    systemctl start stunnel
}

# Uninstall
do_uninstall() {
    echo -e "${YELLOW}Uninstalling stunnel...${NC}"
    
    systemctl stop stunnel 2>/dev/null || true
    systemctl disable stunnel 2>/dev/null || true
    rm -f /etc/systemd/system/stunnel.service
    rm -f /usr/local/bin/${BINARY_NAME}
    systemctl daemon-reload
    
    echo -e "${GREEN}✓ Uninstalled${NC}"
    exit 0
}

# Main install
do_install() {
    local secret="${S:-${X:-$(generate_secret)}}"
    local port="${PORT:-3000}"
    
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       STUNNEL DEPLOY                 ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    
    # Detect system
    local os=$(detect_os)
    local arch=$(detect_arch)
    
    if [ "$os" = "unknown" ]; then
        echo -e "${RED}Error: Unsupported OS${NC}"
        exit 1
    fi
    
    echo -e "${CYAN}OS: $os/$arch${NC}"
    echo -e "${CYAN}Port: $port${NC}"
    echo ""
    
    # Find install location
    local install_dir="/usr/local/bin"
    if [ ! -w "$install_dir" ]; then
        install_dir="$HOME/.local/bin"
        mkdir -p "$install_dir"
    fi
    
    local binary_path="${install_dir}/${BINARY_NAME}"
    
    # Download binary
    download_binary "$os" "$arch" "$binary_path"
    
    # Get public IP
    local public_ip=$(get_public_ip)
    
    # Install service
    if [ -w "/etc/systemd/system" ]; then
        install_service "$secret" "$port" "$binary_path"
        
        echo ""
        echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║       ✓ INSTALLED SUCCESSFULLY       ║${NC}"
        echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
        echo ""
        echo -e "  ${YELLOW}Server:${NC} ${public_ip}:${port}"
        echo -e "  ${YELLOW}Secret:${NC} ${GREEN}${secret}${NC}"
        echo ""
        echo -e "  ${YELLOW}Connect from anywhere:${NC}"
        echo -e "    ${GREEN}S=\"${secret}\" bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/deploy.sh)\"${NC}"
        echo ""
        echo -e "  ${YELLOW}Or install stunnel client:${NC}"
        echo -e "    ${GREEN}curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash${NC}"
        echo -e "    ${GREEN}stunnel connect -s ${secret}${NC}"
        echo ""
        echo -e "  ${YELLOW}Manage service:${NC}"
        echo -e "    systemctl status stunnel"
        echo -e "    systemctl restart stunnel"
        echo -e "    journalctl -u stunnel -f"
        echo ""
    else
        echo ""
        echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║       ✓ INSTALLED (no service)       ║${NC}"
        echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
        echo ""
        echo -e "  ${YELLOW}Binary:${NC} ${binary_path}"
        echo -e "  ${YELLOW}Secret:${NC} ${GREEN}${secret}${NC}"
        echo ""
        echo -e "  ${YELLOW}Run manually:${NC}"
        echo -e "    ${GREEN}${binary_path} server -s ${secret} -p ${port}${NC}"
        echo ""
    fi
}

# Connect mode
do_connect() {
    local secret="${S}"
    local server="${SERVER}"
    
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       STUNNEL CONNECT                ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${YELLOW}Secret:${NC} ${secret}"
    echo ""
    
    # Check if stunnel is installed
    if ! command -v ${BINARY_NAME} &> /dev/null; then
        echo -e "${YELLOW}Installing stunnel client...${NC}"
        
        local os=$(detect_os)
        local arch=$(detect_arch)
        local install_dir="/usr/local/bin"
        
        if [ ! -w "$install_dir" ]; then
            install_dir="$HOME/.local/bin"
            mkdir -p "$install_dir"
        fi
        
        download_binary "$os" "$arch" "${install_dir}/${BINARY_NAME}"
        export PATH="$PATH:${install_dir}"
    fi
    
    # Ask for server address if not provided
    if [ -z "$server" ]; then
        echo -e "${YELLOW}Enter server address (e.g., 1.2.3.4:3000):${NC}"
        read -p "  > " server
    fi
    
    echo ""
    echo -e "  ${YELLOW}Connecting to ${server}...${NC}"
    
    # Connect
    ${BINARY_NAME} connect -s "${secret}" -a "${server}"
}

# Main
main() {
    # Uninstall mode
    if [ "${UNDO}" = "1" ]; then
        do_uninstall
    fi
    
    # Connect mode (if S is set)
    if [ -n "${S}" ]; then
        do_connect
    else
        do_install
    fi
}

main "$@"
