# Stunnel

A TCP tunnel tool for exposing local services through a relay server.

## Usage

### 1. Start server (on VPS)

```bash
stunnel serve
```

Output:
```
  ╔══════════════════════════════════════╗
  ║       STUNNEL SERVER STARTED         ║
  ╚══════════════════════════════════════╝

  Key: MTcyLjI1LjY3LjIwMjo4MDgwOnJJVHJQT2lmMU4zQ0tRenNPdkE1d3c

  On your local machine, run:
  stunnel connect MTcyLjI1LjY3LjIwMjo4MDgwOnJJVHJQT2lmMU4zQ0tRenNPdkE1d3c --local :PORT
```

### 2. Connect from local machine

Copy the key from step 1 and run:

```bash
stunnel connect <PASTE_KEY_HERE> --local :3000
```

This exposes your local port 3000 through the server's public port.

### 3. Access from anywhere

```bash
curl http://vps-ip:8080
```

## Options

### `stunnel serve`
- `--addr` - address for client connections (default `:7000`)
- `--public-addr` - public-facing address (default `:8080`)

### `stunnel connect <key>`
- `--local` - local service to expose (default `localhost:3000`)

## Build

```bash
make build
```

## Test

```bash
make test
```
