# Stunnel

A TCP tunnel tool for exposing local services through a relay server.

## How It Works

```
[Local Machine] ──→ [Relay Server] ←── [Remote Machine]
   (serve)            (bridge)           (connect)
```

Both machines connect to the relay. The relay matches them by secret and bridges the connection.

## Quick Start

### 1. Run relay server (on VPS)

```bash
stunnel relay --addr :7000
```

### 2. Expose local service

```bash
stunnel serve --relay VPS_IP:7000 --local :3000
```

Output:
```
  ╔══════════════════════════════════════╗
  ║       STUNNEL SERVE STARTED          ║
  ╚══════════════════════════════════════╝

  Secret: xK9mP2qR

  On another machine, run:
  stunnel connect --relay VPS_IP:7000 --secret xK9mP2qR
```

### 3. Connect from anywhere

```bash
stunnel connect --relay VPS_IP:7000 --secret xK9mP2qR
```

This pipes your stdin/stdout through the tunnel to the local service.

## Commands

### `stunnel relay`
Run the relay server (typically on a VPS).

- `--addr` - listen address (default `:7000`)

### `stunnel serve`
Expose a local service through the relay.

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
