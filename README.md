# npmctl

A small CLI for **Nginx Proxy Manager** that can:

- create Let's Encrypt certificates using **Cloudflare DNS challenge**
- create proxy hosts
- create a certificate + proxy host in one command

## Requirements

- Python 3.9+
- [uv](https://docs.astral.sh/uv/)
- A running Nginx Proxy Manager API endpoint (for example `http://10.0.2.1:81/api`)

## Install (to your home directory)

```bash
./install.sh
```

This installs the app into `~/.npmctl/venv` and places a launcher at `~/.local/bin/npmctl`.
The installer also adds `~/.local/bin` to your shell PATH config.

## Development setup

```bash
uv sync
```

## CLI usage

### 1) Login and store token in OS keyring

```bash
uv run npmctl login \
  --base-url http://192.168.1.246/api \
  --identity admin@example.com
```

This stores login info securely in your OS keyring (Keychain/Credential Manager/Secret Service).

Check current login status:

```bash
uv run npmctl login-status
```

To remove stored login later:

```bash
uv run npmctl logout
```

### 2) (Recommended) Store Cloudflare token in OS keyring

```bash
uv run npmctl cf-token-set
uv run npmctl cf-token-status
```

After that, certificate commands can use the stored token automatically.

To remove it later:

```bash
uv run npmctl cf-token-delete
```

### 3) Create a Cloudflare DNS challenge certificate

```bash
uv run npmctl add-cert-cloudflare \
  --domain example.com \
  --domain '*.example.com'
```

(You can still override with `--cloudflare-api-token` or `CLOUDFLARE_API_TOKEN`.)

### 4) Create a proxy host with an existing certificate

```bash
uv run npmctl add-proxy-host \
  --domain app.example.com \
  --forward-host 192.168.1.10 \
  --forward-port 8080 \
  --certificate-id 12 \
  --ssl-forced \
  --http2-support
```

### 5) Create certificate + proxy host in one command

```bash
uv run npmctl --debug add-proxy-with-cert \
  --domain app.example.com \
  --forward-host 192.168.1.10 \
  --forward-port 8080
```

## Environment variables

You can use env vars instead of passing values every time:

- `NPM_BASE_URL` (e.g. `http://192.168.1.246/api`)
- `NPM_TOKEN`
- `CLOUDFLARE_API_TOKEN`
- `NPM_CLI_CONFIG` (optional legacy fallback config file location)
- `NPM_CLI_DEBUG=1` (enable verbose request/response logs)
