# npmctl

CLI for managing **Nginx Proxy Manager** and **UniFi DNS policies**.

## Documentation map

Start here and jump to the section you need:

- **[Getting started](docs/getting-started.md)**
- **[UniFi DNS workflows](docs/workflows/unifi-dns.md)**
- **[Nginx Proxy Manager workflows](docs/workflows/nginx-proxy-manager.md)**
- **[Configuration & secrets](docs/configuration.md)**
- **[Troubleshooting](docs/troubleshooting.md)**
- **[Command quick reference](docs/commands.md)**

## Quick start (5 minutes)

### 1) Install

```bash
curl -fsSL https://raw.githubusercontent.com/NiHaiden/nginx-proxy-manager-cli/main/install.sh | bash
```

### 2) Check your environment

```bash
npmctl doctor
```

### 3) Login to Nginx Proxy Manager

```bash
npmctl login --base-url http://192.168.1.246/api --identity admin@example.com
```

### 4) Store your UniFi API key

```bash
npmctl unifi-api-key-set
```

### 5) Create DNS + proxy host in one command

```bash
npmctl add-new-app \
  --gateway-url https://192.168.1.1 \
  --domain app.nhaiden.io \
  --app-ip 192.168.1.246 \
  --app-port 3000 \
  --insecure
```

> `--insecure` is only needed when your UniFi gateway uses a self-signed TLS cert.

## Common tasks

- Add only a UniFi DNS record: [`add-unifi-dns-record`](docs/workflows/unifi-dns.md#add-a-unifi-dns-record)
- List UniFi sites: [`list-unifi-sites`](docs/workflows/unifi-dns.md#list-unifi-sites)
- Create a Cloudflare certificate: [`add-cert-cloudflare`](docs/workflows/nginx-proxy-manager.md#create-a-cloudflare-dns-challenge-certificate)
- Create a proxy host: [`add-proxy-host`](docs/workflows/nginx-proxy-manager.md#create-a-proxy-host-with-an-existing-certificate)

## Project requirements

- Python 3.9+
- [uv](https://docs.astral.sh/uv/)
- Running Nginx Proxy Manager API endpoint (for example `http://10.0.2.1:81/api`)
- (Optional for UniFi commands) UniFi gateway URL (for example `https://192.168.1.1`)

---

If something fails, check **[Troubleshooting](docs/troubleshooting.md)** first.
