# Getting started

[← Back to README](../README.md) · [Command reference](commands.md)

## Requirements

- Python 3.9+
- [uv](https://docs.astral.sh/uv/)
- Nginx Proxy Manager API endpoint (for example `http://10.0.2.1:81/api`)
- Optional for UniFi features: UniFi gateway URL (for example `https://192.168.1.1`)

## Install

### Quick install from GitHub

```bash
curl -fsSL https://raw.githubusercontent.com/NiHaiden/nginx-proxy-manager-cli/main/install.sh | bash
```

### Install from local checkout

```bash
./install.sh
```

The installer puts:

- virtual environment in `~/.npmctl/venv`
- launcher at `~/.local/bin/npmctl`

It also tries to add `~/.local/bin` to your shell PATH config.

## Development setup

```bash
uv sync
```

## First-run checklist

### 1) Run diagnostics

```bash
npmctl doctor
```

### 2) Login and store NPM token in keyring

```bash
npmctl login --base-url http://192.168.1.246/api --identity admin@example.com
```

Check status:

```bash
npmctl login-status
```

Remove stored login:

```bash
npmctl logout
```

### 3) Store UniFi API key (recommended)

```bash
npmctl unifi-api-key-set
npmctl unifi-api-key-status
```

Remove stored key:

```bash
npmctl unifi-api-key-delete
```

### 4) Store Cloudflare token (if using cert commands)

```bash
npmctl cf-token-set
npmctl cf-token-status
```

Remove stored token:

```bash
npmctl cf-token-delete
```

## Next steps

- For UniFi DNS and one-command app setup: **[UniFi DNS workflows](workflows/unifi-dns.md)**
- For certs and proxy hosts: **[Nginx Proxy Manager workflows](workflows/nginx-proxy-manager.md)**
- For env vars and secret behavior: **[Configuration & secrets](configuration.md)**
