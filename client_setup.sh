#!/bin/bash
# Stunnel Client Setup
# Usage: bash <(curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/client_setup.sh) <secret> <server_ip>

set -e

SECRET=${1:-""}
SERVER=${2:-""}

if [ -z "$SECRET" ] || [ -z "$SERVER" ]; then
    echo "Usage: bash client_setup.sh <secret> <server_ip>"
    echo "Example: bash client_setup.sh eylfLD9GStnstB+e 93.177.100.9"
    exit 1
fi

echo ""
echo "╔══════════════════════════════════════╗"
echo "║       STUNNEL CLIENT SETUP           ║"
echo "╚══════════════════════════════════════╝"
echo ""

# Download stunnel
echo "Downloading stunnel..."
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash

# Connect
echo "Connecting to $SERVER..."
echo ""
stunnel connect -s $SECRET -r $SERVER:443
