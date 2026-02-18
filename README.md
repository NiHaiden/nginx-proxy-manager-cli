# npmctl

A small CLI for **Nginx Proxy Manager** that can:

- create Let's Encrypt certificates using **Cloudflare DNS challenge**
- create proxy hosts
- create a certificate + proxy host in one command

## Requirements

- Python 3.9+
- [uv](https://docs.astral.sh/uv/)
- A running Nginx Proxy Manager API endpoint (for example `http://10.0.2.1:81/api`)

## Install

### Quick install from GitHub (no clone required)

```bash
curl -fsSL https://raw.githubusercontent.com/NiHaiden/nginx-proxy-manager-cli/main/install.sh | bash
```

### Install from a local checkout

```bash
./install.sh
```

Both install methods place the app into `~/.npmctl/venv` and create a launcher at
`~/.local/bin/npmctl`. The installer also adds `~/.local/bin` to your shell PATH config.

Optional: install from a different branch/tag when using the GitHub installer:

```bash
NPMCTL_GITHUB_REF=main curl -fsSL https://raw.githubusercontent.com/NiHaiden/nginx-proxy-manager-cli/main/install.sh | bash
```

## Development setup

```bash
uv sync
```

## CLI usage

### 0) Run environment diagnostics (recommended on new systems)

```bash
npmctl doctor
```

This checks whether a recommended OS keyring backend is available and prints distro-specific
install hints for Fedora/Arch/Ubuntu when it is missing.

### 1) Login and store token in OS keyring (base URL required)

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

## Troubleshooting

### `Failed to read secret from keyring: No recommended backend was available`

`npmctl` stores login and Cloudflare tokens in your OS keyring. On some Linux setups,
no keyring backend is available by default.

Start with:

```bash
npmctl doctor
```

Check available backends:

```bash
~/.npmctl/venv/bin/python -m keyring --list-backends
```

If this shows no usable backend, you have two options:

1. **Recommended (secure):** install and run an OS keyring service (for example
   GNOME Keyring / Secret Service, or KWallet).
2. **Fallback (less secure):** install `keyrings.alt` in the same environment as `npmctl`.

If you installed `npmctl` with the installer, run:

```bash
~/.npmctl/venv/bin/pip install keyrings.alt
```

Then retry:

```bash
npmctl login-status
```

> Note: `keyrings.alt` uses non-recommended backends. Prefer a real OS keyring when possible.

## Environment variables

You can use env vars instead of passing values every time:

- `NPM_BASE_URL` (e.g. `http://192.168.1.246/api`)
- `NPM_TOKEN`
- `CLOUDFLARE_API_TOKEN`
- `NPM_CLI_CONFIG` (optional legacy fallback config file location)
- `NPM_CLI_DEBUG=1` (enable verbose request/response logs)
