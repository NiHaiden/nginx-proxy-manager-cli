# Nginx Proxy Manager workflows

[← Back to README](../../README.md) · [UniFi workflows](unifi-dns.md)

## Create a Cloudflare DNS challenge certificate

```bash
npmctl add-cert-cloudflare \
  --domain example.com \
  --domain '*.example.com'
```

Cloudflare token sources (first wins):

1. `--cloudflare-api-token`
2. `CLOUDFLARE_API_TOKEN`
3. keyring token from `npmctl cf-token-set`

## Create a proxy host with an existing certificate

```bash
npmctl add-proxy-host \
  --domain app.example.com \
  --forward-host 192.168.1.10 \
  --forward-port 8080 \
  --certificate-id 12 \
  --ssl-forced \
  --http2-support
```

## Create certificate + proxy host in one command

```bash
npmctl --debug add-proxy-with-cert \
  --domain app.example.com \
  --forward-host 192.168.1.10 \
  --forward-port 8080
```

This does:

1. Create Cloudflare DNS challenge cert
2. Create proxy host using the new certificate

## Authentication

NPM auth sources (first wins):

1. `--base-url` and `--token`
2. `NPM_BASE_URL` and `NPM_TOKEN`
3. keyring login from `npmctl login`
