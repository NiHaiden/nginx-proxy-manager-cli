package npmctl

import "fmt"

func (c *CLI) printRootHelp() {
	fmt.Fprintln(c.out, `CLI for Nginx Proxy Manager and UniFi DNS automation.

Preferred command styles:
  npmctl proxy add ...
  npmctl add proxy ...
  npmctl secret set cloudflare-token

Usage:
  npmctl [--debug] <command> [options]

Commands:
  auth      Manage authentication and saved NPM login state.
  secret    Store and inspect local secrets in your OS keyring.
  site      Inspect UniFi sites.
  dns       Manage UniFi DNS records.
  proxy     Manage Nginx Proxy Manager proxy hosts.
  cert      Manage Nginx Proxy Manager certificates.
  app       Provision an application workflow across DNS and proxy layers.
  add       Workflow shortcuts for creating resources.
  list      Workflow shortcuts for listing resources.
  status    Workflow shortcuts for inspecting saved state.
  delete    Workflow shortcuts for deleting saved secrets.
  doctor    Run environment diagnostics and keyring backend checks.

Options:
  -h, --help   Show help.
  --debug      Enable verbose request/response debug logging.`)
}

func (c *CLI) printGroupHelp(group string) {
	helps := map[string]string{
		"auth": `Manage authentication and saved NPM login state.

Commands:
  login
  status
  logout`,
		"secret": `Store and inspect local secrets in your OS keyring.

Commands:
  set     Store cloudflare-token or unifi-api-key.
  status  Show whether cloudflare-token or unifi-api-key is stored.
  delete  Delete cloudflare-token or unifi-api-key.`,
		"site": `Inspect UniFi sites.

Commands:
  list`,
		"dns": `Manage UniFi DNS records.

Commands:
  add`,
		"proxy": `Manage Nginx Proxy Manager proxy hosts.

Commands:
  add`,
		"cert": `Manage Nginx Proxy Manager certificates.

Commands:
  add`,
		"app": `Provision an application workflow across DNS and proxy layers.

Commands:
  add`,
		"add": `Workflow shortcuts for creating resources.

Commands:
  proxy
  cert
  dns
  app`,
		"list": `Workflow shortcuts for listing resources.

Commands:
  sites`,
		"status": `Workflow shortcuts for inspecting saved state.

Commands:
  login
  cloudflare-token
  unifi-api-key`,
		"delete": `Workflow shortcuts for deleting saved secrets.

Commands:
  cloudflare-token
  unifi-api-key`,
	}
	fmt.Fprintln(c.out, helps[group])
}

func (c *CLI) printCommandHelp(command string) {
	helps := map[string]string{
		"auth login": `Login and save your JWT token in OS keyring.

Options:
  --identity <email>   NPM user email.
  --secret <password>  NPM password. Prompts when omitted.
  --base-url <url>     API base URL, e.g. http://host/api.
  --scope <scope>      Token scope. Default: user.`,
		"secret set": `Store a secret securely in OS keyring.

Arguments:
  cloudflare-token
  unifi-api-key

Options:
  --value <value>      Secret value. Prompts when omitted.`,
		"proxy add": `Create a proxy host, optionally creating a Cloudflare certificate first.

Options:
  --domain, -d <name>              Domain (repeatable).
  --forward-host <host>            Upstream host/IP.
  --forward-port <port>            Upstream port.
  --forward-scheme <http|https>    Upstream scheme. Default: http.
  --certificate-id <id>            Existing certificate ID.
  --create-cert                    Create a new Cloudflare certificate before creating the proxy host.
  --cloudflare-api-token <token>   Cloudflare API token (used with --create-cert).
  --certificate-name <name>        Human-friendly certificate name (used with --create-cert).
  --propagation-seconds <seconds>  DNS propagation wait in seconds. Default: 30.
  --key-type <rsa|ecdsa>           Certificate key type. Default: rsa.
  --ssl-forced                     Force HTTPS redirect.
  --http2-support                  Enable HTTP/2.
  --hsts-enabled                   Enable HSTS.
  --hsts-subdomains                Apply HSTS to subdomains.
  --base-url <url>                 API base URL.
  --token <jwt>                    JWT token.`,
		"cert add": `Create a Let's Encrypt certificate using Cloudflare DNS challenge.

Options:
  --domain, -d <name>              Domain (repeatable).
  --cloudflare-api-token <token>   Cloudflare API token.
  --nice-name <name>               Human-friendly certificate name.
  --propagation-seconds <seconds>  DNS propagation wait in seconds. Default: 30.
  --key-type <rsa|ecdsa>           Certificate key type. Default: rsa.
  --base-url <url>                 API base URL.
  --token <jwt>                    JWT token.`,
		"site list": `List UniFi Network site IDs for DNS policy operations.

Options:
  --gateway-url <url>       UniFi gateway URL. Also reads UNIFI_BASE_URL.
  --unifi-api-key <key>     UniFi integration API key. Also reads UNIFI_API_KEY.
  --insecure                Disable TLS certificate verification (unsafe).`,
		"dns add": `Create a UniFi DNS policy record using the integration API.

Options:
  --site-id <id>            UniFi site ID (optional when exactly one site exists).
  --domain, -d <name>       DNS name to create.
  --ipv4-address <addr>     IPv4 address for the A record.
  --ttl-seconds <seconds>   TTL in seconds. Default: 14400.
  --enabled / --disabled    Enable or disable record. Default: enabled.
  --record-type <type>      Policy type. Default: A_RECORD.
  --gateway-url <url>       UniFi gateway URL. Also reads UNIFI_BASE_URL.
  --unifi-api-key <key>     UniFi integration API key. Also reads UNIFI_API_KEY.
  --insecure                Disable TLS certificate verification (unsafe).`,
		"app add": `Create UniFi DNS and NPM proxy host for a new app in one command.

Options:
  --domain, -d <name>       Domain for DNS and proxy host.
  --app-ip <addr>           App IPv4 address.
  --app-port <port>         App upstream port.
  --site-id <id>            UniFi site ID (optional when exactly one site exists).
  --ttl-seconds <seconds>   TTL in seconds. Default: 14400.
  --gateway-url <url>       UniFi gateway URL. Also reads UNIFI_BASE_URL.
  --unifi-api-key <key>     UniFi integration API key. Also reads UNIFI_API_KEY.
  --insecure                Disable TLS certificate verification for UniFi calls.
  --base-url <url>          NPM API base URL.
  --token <jwt>             NPM JWT token.
  --certificate-id <id>     Existing certificate ID.`,
		"doctor": `Run environment diagnostics and keyring backend checks.`,
	}
	if help, ok := helps[command]; ok {
		fmt.Fprintln(c.out, help)
	}
}
