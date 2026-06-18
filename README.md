# Stunnel

Connect like there is no firewall. No VPS needed.

## Quick Start (One Command)

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/quick.sh)"
```

## How It Works

```
[Server] → Relay Server ← [Client]
    ↓           ↓           ↓
  -s secret   matches    -s secret
    ↓           ↓           ↓
    └───────────┴───────────┘
        Direct Connection
```

## Usage

### 1. Start Relay Server (on VPS)

```bash
# Download relay server
wget https://github.com/Lanexus/stunnel/releases/download/v0.5.0/relay-linux-amd64 -O /usr/local/bin/relay
chmod +x /usr/local/bin/relay

# Run relay
relay 8080
```

### 2. Generate Secret

```bash
stunnel -g
# Output: T-JBUwDOXjw
```

### 3. Server (expose local service)

```bash
stunnel -l -p 3000 -s T-JBUwDOXjw
```

### 4. Client (connect)

```bash
stunnel -s T-JBUwDOXjw
```

## Commands

```bash
# Generate secret
stunnel -g

# Server (expose port 3000)
stunnel -l -p 3000 -s <secret>

# Client (connect)
stunnel -s <secret>

# Server with shell
stunnel -l --shell -s <secret>

# Client with shell
stunnel -s <secret> --shell

# Custom signaling server
stunnel -l -p 3000 -s <secret> --signaling http://your-server:8080
```

## Build

```bash
make build
```

## Test

```bash
make test
```
