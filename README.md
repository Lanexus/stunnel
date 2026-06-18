# Stunnel

Connect like there is no firewall. No VPS needed.

## Quick Install

```bash
curl -sL https://raw.githubusercontent.com/Lanexus/stunnel/master/install.sh | bash
```

## Usage

### Server (expose local service)
```bash
stunnel -l -p 3000
```

Output:
```
  ╔══════════════════════════════════════╗
  ║       STUNNEL SERVER                 ║
  ╚══════════════════════════════════════╝

  Secret: nGW_8dn8n24
  Port:   3000

  Starting tunnel...

  ╔══════════════════════════════════════╗
  ║       TUNNEL ACTIVE                  ║
  ╚══════════════════════════════════════╝

  URL: https://abc123.trycloudflare.com

  Share this URL to access your service
```

### Client (access the service)
Just open the URL in browser, or:
```bash
curl https://abc123.trycloudflare.com
```

## Commands

```bash
# Expose port 3000
stunnel -l -p 3000

# Expose with custom secret
stunnel -l -p 3000 -s mysecret

# Generate secret
stunnel -g

# Interactive shell
stunnel -l --shell
```

## How It Works

```
[Your Computer] → Cloudflare Tunnel → [Internet]
   (local:3000)    (*.trycloudflare.com)
```

No VPS needed. No port forwarding. Just works.

## Build

```bash
make build
```

## Test

```bash
make test
```
