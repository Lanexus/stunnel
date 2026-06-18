# Stunnel

A TCP tunnel tool for exposing local services through a relay server.

## Quick Start (No VPS Needed!)

### Cloudflare Tunnel Mode

```bash
# Expose local service to the internet
stunnel tunnel --local :3000

# Output:
#   URL: https://abc123.trycloudflare.com
#   Share this URL to access your service
```

That's it! No VPS, no domain, no config.

## Other Modes

### Relay Mode (Your Own VPS)

```bash
# On VPS
stunnel relay --addr :7000

# On local machine
stunnel serve --relay VPS_IP:7000 --local :3000
# Output: Secret: xK9mP2qR

# On another machine
stunnel connect --relay VPS_IP:7000 --secret xK9mP2qR
```

## Commands

### `stunnel tunnel`
Expose local service via Cloudflare Tunnel (free, no VPS needed).

- `--local` - local service to expose (default `localhost:3000`)

### `stunnel relay`
Run relay server (on VPS).

- `--addr` - listen address (default `:7000`)

### `stunnel serve`
Expose local service through relay.

- `--relay` - relay server address (default `localhost:7000`)
- `--local` - local service to expose (default `localhost:3000`)
- `--secret` - shared secret (auto-generated if empty)

### `stunnel connect`
Connect to a served tunnel.

- `--relay` - relay server address (default `localhost:7000`)
- `--secret` - shared secret (required)

## Build

```bash
make build
```

## Test

```bash
make test
```
