# Stunnel

Connect like there is no firewall. Both sides connect OUTBOUND to relay. No ports needed.

## Quick Start

### 1. On VPS (run relay)
```bash
# Download relay
wget https://github.com/Lanexus/stunnel/releases/latest/download/relay-linux-amd64 -O /usr/local/bin/relay
chmod +x /usr/local/bin/relay

# Run relay
relay 7000
```

### 2. On Server (expose local service)
```bash
# Install
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash

# Install as service (connects to relay)
stunnel server --install -s <secret> -r VPS_IP:7000
```

### 3. On Client (connect)
```bash
# Install
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash

# Connect
stunnel connect -s <secret> -r VPS_IP:7000
```

## How It Works

```
[Server] → OUTBOUND → [Relay on VPS] ← OUTBOUND ← [Client]
   ↓                      ↓                      ↓
 local:3000          matches by secret         stdin/stdout
```

Both sides connect OUTBOUND to the relay. No inbound ports needed on server or client.

## Commands

```bash
# Generate secret
stunnel server -g

# Run relay (on VPS)
relay 7000

# Expose local service (connects to relay)
stunnel server -s <secret> -r VPS_IP:7000 -p 3000

# Install as service
stunnel server --install -s <secret> -r VPS_IP:7000

# Connect (connects to relay)
stunnel connect -s <secret> -r VPS_IP:7000

# Uninstall service
stunnel server --uninstall
```

## Manage Service

```bash
systemctl status stunnel
systemctl restart stunnel
systemctl stop stunnel
journalctl -u stunnel -f
```
