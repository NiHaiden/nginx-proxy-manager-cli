# Troubleshooting

[← Back to README](../README.md) · [Configuration](configuration.md)

## Keyring backend is missing

Error example:

```text
Failed to read secret from keyring: No recommended backend was available
```

### Fix steps

1. Run diagnostics:

```bash
npmctl doctor
```

2. Check detected keyring backends:

```bash
~/.npmctl/venv/bin/python -m keyring --list-backends
```

3. Install/run a recommended OS keyring service:

- Secret Service (GNOME Keyring)
- KWallet

4. Re-check status:

```bash
npmctl auth status
```

### Fallback (less secure)

```bash
~/.npmctl/venv/bin/pip install keyrings.alt
```

## UniFi self-signed certificate errors

Error example:

```text
HTTP request failed: [SSL: CERTIFICATE_VERIFY_FAILED]
```

Use `--insecure` for UniFi commands:

```bash
npmctl site list --gateway-url https://192.168.1.1 --insecure
npmctl dns add --gateway-url https://192.168.1.1 --domain app.local --ipv4-address 192.168.1.50 --insecure
npmctl app add --gateway-url https://192.168.1.1 --domain app.local --app-ip 192.168.1.50 --app-port 3000 --insecure
```

> `--insecure` disables TLS certificate verification for UniFi requests.

## API errors with little detail

Run with debug enabled to include request/response details:

```bash
npmctl --debug <your-command>
```

If needed:

```bash
export NPM_CLI_DEBUG=1
```
