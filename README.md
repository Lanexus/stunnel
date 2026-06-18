# Stunnel

Connect like there is no firewall. Install once, connect anytime.

## Quick Install

```bash
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash
```

## How It Works

```
[VPS] ← stunnel server (persistent)
  ↓
[You] ← stunnel connect -s secret
```

## Usage

### 1. Install on VPS (once)

```bash
# Install stunnel
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash

# Install as system service (runs on boot)
stunnel server --install
```

Output:
```
  ╔══════════════════════════════════════╗
  ║       INSTALLING STUNNEL             ║
  ╚══════════════════════════════════════╝

  Secret: T-JBUwDOXjw
  Port: 3000

  ✓ Installed as systemd service
  ✓ Enabled on boot
  ✓ Started

  Your secret (save this!):
    T-JBUwDOXjw

  Connect from anywhere:
    stunnel connect -s T-JBUwDOXjw
```

### 2. Connect from anywhere

```bash
stunnel connect -s T-JBUwDOXjw
```

### 3. Get shell access

```bash
# On VPS (install with shell)
stunnel server --install --shell

# From anywhere
stunnel connect -s T-JBUwDOXjw --shell
```

## Commands

```bash
# Generate secret
stunnel server -g

# Run server (foreground)
stunnel server -s <secret> -p 3000

# Install as service (persistent)
stunnel server --install -s <secret>

# Install with shell
stunnel server --install -s <secret> --shell

# Connect
stunnel connect -s <secret>

# Connect with shell
stunnel connect -s <secret> --shell

# Uninstall service
stunnel server --uninstall
```

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
