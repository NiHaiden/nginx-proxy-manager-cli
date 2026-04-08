# Configuration & secrets

[← Back to README](../README.md) · [Troubleshooting](troubleshooting.md)

## Secret storage

`npmctl` stores secrets in your OS keyring.

### Login token

- `npmctl auth login`
- `npmctl auth status`
- `npmctl auth logout`

### Cloudflare token

- `npmctl secret set cloudflare-token`
- `npmctl secret status cloudflare-token`
- `npmctl secret delete cloudflare-token`

### UniFi API key

- `npmctl secret set unifi-api-key`
- `npmctl secret status unifi-api-key`
- `npmctl secret delete unifi-api-key`

## Environment variables

| Variable | Purpose |
|---|---|
| `NPM_BASE_URL` | NPM API URL (for example `http://192.168.1.246/api`) |
| `NPM_TOKEN` | NPM JWT token |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token for DNS challenge certificates |
| `UNIFI_BASE_URL` | UniFi gateway URL (for example `https://192.168.1.1`) |
| `UNIFI_API_KEY` | UniFi integration API key |
| `NPM_CLI_CONFIG` | Optional legacy fallback config file location |
| `NPM_CLI_DEBUG` | Set to `1` to enable debug request/response logs |

## Precedence rules

Most auth/token values follow this priority:

1. CLI option
2. Environment variable
3. Stored keyring secret

## Debugging

Use debug logs when troubleshooting API errors:

```bash
npmctl --debug <command>
```

Or globally via environment variable:

```bash
export NPM_CLI_DEBUG=1
```
