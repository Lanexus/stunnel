# Stunnel - TCP Tunnel Tool

## Quick Install (One-Liner)

```bash
bash <(curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh)
```

Or with wget:

```bash
bash <(wget -qO- https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh)
```

## Manual Install

```bash
# Linux amd64
wget https://github.com/Lanexus/stunnel/releases/download/v0.2.0/stunnel-linux-amd64 -O /usr/local/bin/stunnel
chmod +x /usr/local/bin/stunnel

# Linux arm64
wget https://github.com/Lanexus/stunnel/releases/download/v0.2.0/stunnel-linux-arm64 -O /usr/local/bin/stunnel
chmod +x /usr/local/bin/stunnel

# macOS
curl -sL https://github.com/Lanexus/stunnel/releases/download/v0.2.0/stunnel-darwin-amd64 -o /usr/local/bin/stunnel
chmod +x /usr/local/bin/stunnel
```

## Usage

### Expose local service (Free, No VPS)
```bash
stunnel tunnel --local :3000
# Output: https://abc123.trycloudflare.com
```

### Relay Mode (Your VPS)
```bash
# On VPS
stunnel relay --addr :7000

# On local
stunnel serve --relay VPS_IP:7000 --local :3000

# On another machine
stunnel connect --relay VPS_IP:7000 --secret KEY
```

### Netcat Mode
```bash
stunnel nc -l :8080          # Listen
stunnel nc example.com:80    # Connect
```

### File Transfer
```bash
stunnel file -l :9090        # Receive
stunnel file host:9090 file  # Send
```
