# Stunnel

A TCP tunnel tool for exposing local services through a relay server.

## Usage

### Server (on VPS)

```bash
stunnel server --addr :7000 --public-addr :8080 --secret mysecret
```

- `--addr` - address to listen for client connections (default `:7000`)
- `--public-addr` - address for user-facing connections (default `:8080`)
- `--secret` - shared authentication secret

### Client (on local machine)

```bash
stunnel client --server vps-ip:7000 --secret mysecret --local :3000
```

- `--server` - address of the stunnel server
- `--secret` - shared authentication secret (must match server)
- `--local` - local address to tunnel traffic to

### User (any machine)

```bash
curl vps-ip:8080
```

## Build

```bash
make build
```

## Test

```bash
make test
```
