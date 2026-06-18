# Stunnel

Connect like there is no firewall. Install once, connect anytime.

## Quick Deploy (One Command)

### Install on VPS
```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/deploy.sh)"
```

### Connect from anywhere
```bash
S="your-secret" bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/deploy.sh)"
```

### Uninstall
```bash
UNDO=1 bash -c "$(curl -fsSL https://raw.githubusercontent.com/Lanexus/stunnel/master/deploy.sh)"
```

## How It Works

```
[VPS] ← stunnel server (persistent, systemd)
  ↓
[You] ← stunnel connect -s secret
```

## Manual Usage

### 1. Install on VPS

```bash
# Download stunnel
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash

# Install as service
stunnel server --install
```

### 2. Connect from anywhere

```bash
# Install client
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash

# Connect
stunnel connect -s <secret>
```

## Commands

```bash
# Generate secret
stunnel server -g

# Run server (foreground)
stunnel server -s <secret> -p 3000

# Install as service (persistent)
stunnel server --install -s <secret>

# Connect
stunnel connect -s <secret> -a <server-ip>:3000

# Connect with shell
stunnel connect -s <secret> -a <server-ip>:3000 --shell

# Uninstall service
stunnel server --uninstall
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `S` | Secret key for connection |
| `X` | Predefined secret for installation |
| `PORT` | Port to expose (default: 3000) |
| `UNDO` | Set to 1 to uninstall |
| `SERVER` | Server address for connection |

## Manage Service

```bash
# Check status
systemctl status stunnel

# Restart
systemctl restart stunnel

# Stop
systemctl stop stunnel

# View logs
journalctl -u stunnel -f
```

## Build

```bash
make build
```

## Test

```bash
make test
```
