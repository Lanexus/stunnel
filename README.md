# Stunnel

A TCP tunnel tool for exposing local services through a relay server.

## Quick Install

```bash
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash
```

## Quick Start (No VPS Needed!)

```bash
# Expose local service to the internet
stunnel tunnel --local :3000

# Output:
#   URL: https://abc123.trycloudflare.com
#   Share this URL to access your service
```

## Commands

### `stunnel tunnel`
Expose local service via Cloudflare Tunnel (free, no VPS needed).

```bash
stunnel tunnel --local :3000
```

### `stunnel relay`
Run relay server (on VPS).

```bash
stunnel relay --addr :7000
```

### `stunnel serve`
Expose local service through relay.

```bash
stunnel serve --relay VPS_IP:7000 --local :3000
```

### `stunnel connect`
Connect to a served tunnel.

```bash
stunnel connect --relay VPS_IP:7000 --secret KEY
```

### `stunnel nc`
Netcat mode - connect or listen.

```bash
# Listen for connections
stunnel nc -l :8080

# Connect to address
stunnel nc example.com:80

# Execute command on connection
stunnel nc -l :8080 -e "echo hello"
```

### `stunnel file`
File transfer mode.

```bash
# Send file
stunnel file receiver:9090 myfile.txt

# Receive file
stunnel file -l :9090
```

### `stunnel stop`
Stop running daemon.

```bash
stunnel stop
```

## Build from Source

```bash
git clone https://github.com/lanexus/stunnel.git
cd stunnel
make build
```

## Test

```bash
make test
```

## Architecture

```
Cloudflare Mode (Recommended):
[Local:3000] → cloudflared → [Cloudflare] → [User]

Relay Mode:
[Local:3000] → stunnel serve → [VPS Relay] → stunnel connect → [User]
```

## License

MIT
