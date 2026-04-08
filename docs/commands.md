# Command quick reference

[← Back to README](../README.md)

## Auth

- `npmctl auth login`
- `npmctl auth status`
- `npmctl auth logout`

## Secrets

- `npmctl secret set cloudflare-token`
- `npmctl secret status cloudflare-token`
- `npmctl secret delete cloudflare-token`
- `npmctl secret set unifi-api-key`
- `npmctl secret status unifi-api-key`
- `npmctl secret delete unifi-api-key`

## UniFi DNS

- `npmctl site list`
- `npmctl dns add`
- `npmctl app add`

## Nginx Proxy Manager

- `npmctl cert add`
- `npmctl proxy add`

## Diagnostics

- `npmctl doctor`

## Workflow shortcuts

- `npmctl add proxy`
- `npmctl add cert`
- `npmctl add dns`
- `npmctl add app`
- `npmctl list sites`
- `npmctl status login`
- `npmctl status cloudflare-token`
- `npmctl status unifi-api-key`
- `npmctl delete cloudflare-token`
- `npmctl delete unifi-api-key`

## Handy help commands

```bash
npmctl --help
npmctl <command> --help
```
