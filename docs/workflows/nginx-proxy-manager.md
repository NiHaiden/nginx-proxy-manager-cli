# Nginx Proxy Manager workflows

[← Back to README](../../README.md) · [UniFi workflows](unifi-dns.md)

## Create a Cloudflare DNS challenge certificate

```bash
npmctl cert add \
  --domain example.com \
  --domain '*.example.com'
```

Shortcut:

```bash
npmctl add cert \
  --domain example.com \
  --domain '*.example.com'
```

Cloudflare token sources (first wins):

1. `--cloudflare-api-token`
2. `CLOUDFLARE_API_TOKEN`
3. keyring token from `npmctl secret set cloudflare-token`

## Create a proxy host

```bash
npmctl proxy add \
  --domain app.example.com \
  --forward-host 192.168.1.10 \
  --forward-port 8080 \
  --certificate-id 12 \
  --ssl-forced \
  --http2-support
```

Shortcut:

```bash
npmctl add proxy \
  --domain app.example.com \
  --forward-host 192.168.1.10 \
  --forward-port 8080 \
  --certificate-id 12 \
  --ssl-forced \
  --http2-support
```

## Create certificate + proxy host in one command

```bash
npmctl --debug proxy add \
  --domain app.example.com \
  --forward-host 192.168.1.10 \
  --forward-port 8080 \
  --create-cert \
  --ssl-forced \
  --http2-support \
  --hsts-enabled \
  --hsts-subdomains
```

This does:

1. Create Cloudflare DNS challenge cert
2. Create proxy host using the new certificate

## Authentication

NPM auth sources (first wins):

1. `--base-url` and `--token`
2. `NPM_BASE_URL` and `NPM_TOKEN`
3. keyring login from `npmctl auth login`
