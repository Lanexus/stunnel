# Stunnel

Connect like there is no firewall. No VPS needed.

## Quick Install

```bash
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash
```

## How It Works

```
[Server] → Signaling Server ← [Client]
    ↓           ↓           ↓
  -s secret   matches    -s secret
    ↓           ↓           ↓
    └───────────┴───────────┘
        Direct Connection
```

## Usage

### 1. Start Signaling Server (on VPS)

```bash
# Download signaling server
wget https://github.com/Lanexus/stunnel/releases/download/v0.5.0/signaling-linux-amd64 -O /usr/local/bin/signaling
chmod +x /usr/local/bin/signaling

# Run signaling server
signaling 8080
```

### 2. Start Server (expose local service)

```bash
stunnel -l -p 3000 --signaling http://VPS_IP:8080
```

### 3. Connect from Client

```bash
stunnel -s <secret> --signaling http://VPS_IP:8080
```

## Commands

```bash
# Generate secret
stunnel -g

# Server (expose port 3000)
stunnel -l -p 3000 --signaling http://VPS_IP:8080

# Client (connect)
stunnel -s <secret> --signaling http://VPS_IP:8080

# Server with shell
stunnel -l --shell --signaling http://VPS_IP:8080

# Client with shell
stunnel -s <secret> --shell --signaling http://VPS_IP:8080
```

## Build

```bash
make build
```

## Test

```bash
make test
```
