# Stunnel

Connect like there is no firewall. Securely.

## Quick Install

```bash
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash
```

## Usage

### Generate Secret
```bash
stunnel -g
# Output: G1HfImJCnQB7fV0M
```

### Listen (Server)
```bash
stunnel -s G1HfImJCnQB7fV0M -l
```

### Connect (Client)
```bash
stunnel -s G1HfImJCnQB7fV0M
```

### Interactive Shell
```bash
# Server (listen with shell)
stunnel -s G1HfImJCnQB7fV0M -l --shell

# Client (connect with shell)
stunnel -s G1HfImJCnQB7fV0M --shell
```

### Port Forwarding
```bash
# Server (listen, forward port 22)
stunnel -s G1HfImJCnQB7fV0M -l -p 22

# Client (connect, forward port 22)
stunnel -s G1HfImJCnQB7fV0M -p 22
```

## How It Works

```
[Server] → Relay Server ← [Client]
   ↓           ↓           ↓
  -s secret  matches    -s secret
   ↓           ↓           ↓
   └───────────┴───────────┘
           Connected!
```

Both users use the same secret. The relay server matches them automatically.

## Run Your Own Relay

```bash
# Build relay
go build -o relay ./cmd/relay/

# Run relay on port 7000
./relay :7000
```

Then use custom relay:
```bash
stunnel -s secret -l -r your-server:7000
stunnel -s secret -r your-server:7000
```

## Features

- **Simple** - Just one command with a secret
- **Secure** - End-to-end encrypted
- **Bypass Firewall** - Works through NAT/Firewall
- **Multiple Modes** - Shell, port forwarding, pipe
- **Auto Match** - No IP/Port needed, just secret

## Build

```bash
make build
```

## Test

```bash
make test
```
