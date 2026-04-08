# UniFi DNS workflows

[← Back to README](../../README.md) · [NPM workflows](nginx-proxy-manager.md)

## List UniFi sites

Use this to discover site IDs:

```bash
npmctl site list --gateway-url https://192.168.1.1
```

If UniFi uses a self-signed cert:

```bash
npmctl list sites --gateway-url https://192.168.1.1 --insecure
```

## Add a UniFi DNS record

If you have exactly one site, `--site-id` is optional.

```bash
npmctl dns add \
  --gateway-url https://192.168.1.1 \
  --domain test.nhaiden.io \
  --ipv4-address 192.168.1.246 \
  --ttl-seconds 14400
```

With self-signed UniFi cert:

```bash
npmctl add dns \
  --gateway-url https://192.168.1.1 \
  --domain test.nhaiden.io \
  --ipv4-address 192.168.1.246 \
  --ttl-seconds 14400 \
  --insecure
```

If you have multiple sites, pass `--site-id` explicitly:

```bash
npmctl dns add \
  --gateway-url https://192.168.1.1 \
  --site-id 88f7af54-98f8-306a-a1c7-c9349722b1f6 \
  --domain test.nhaiden.io \
  --ipv4-address 192.168.1.246
```

## Add a full app in one command (DNS + proxy)

`app add` creates:

1. UniFi DNS record (`domain -> app-ip`)
2. NPM proxy host (`domain -> app-ip:app-port`)

```bash
npmctl app add \
  --gateway-url https://192.168.1.1 \
  --domain app.nhaiden.io \
  --app-ip 192.168.1.246 \
  --app-port 3000 \
  --insecure
```

Optional NPM-related flags you can pass through:

- `--forward-scheme` (`http` or `https`)
- `--certificate-id`
- `--ssl-forced`
- `--http2-support`
- `--hsts-enabled`
- `--hsts-subdomains`

## Notes

- `--insecure` disables TLS verification for UniFi requests only.
- UniFi API key can come from:
  1. `--unifi-api-key`
  2. `UNIFI_API_KEY`
  3. stored keyring secret (`npmctl secret set unifi-api-key`)
