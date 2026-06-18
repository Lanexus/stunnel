#!/bin/bash
# Stunnel VPS Auto Setup
# Usage: bash <(curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/vps_setup.sh)

set -e

echo ""
echo "╔══════════════════════════════════════╗"
echo "║       STUNNEL VPS SETUP              ║"
echo "╚══════════════════════════════════════╝"
echo ""

# Generate secret
SECRET=$(openssl rand -base64 12)

# Stop old processes
pkill relay 2>/dev/null || true
pkill stunnel 2>/dev/null || true
systemctl stop stunnel 2>/dev/null || true

# Download relay
echo "Downloading relay..."
wget -q https://github.com/Lanexus/stunnel/releases/download/v0.7.1/relay-linux-amd64 -O /usr/local/bin/relay
chmod +x /usr/local/bin/relay

# Start relay on port 443 (always open)
echo "Starting relay on port 443..."
nohup relay 443 > /tmp/relay.log 2>&1 &

# Wait for relay to start
sleep 1

# Download stunnel
echo "Downloading stunnel..."
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash

# Install as service
echo "Installing stunnel service..."
stunnel server --install -s $SECRET -p 3000 -r localhost:443

echo ""
echo "╔══════════════════════════════════════╗"
echo "║       ✓ STUNNEL READY!               ║"
echo "╚══════════════════════════════════════╝"
echo ""
echo "  Secret: $SECRET"
echo ""
echo "  Connect from anywhere:"
echo "    stunnel connect -s $SECRET -r $(curl -s ifconfig.me):443"
echo ""
echo "  Manage:"
echo "    systemctl status stunnel"
echo "    systemctl restart stunnel"
echo ""
