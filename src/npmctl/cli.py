from __future__ import annotations

import typer

from .commands import auth, certificates, doctor, proxy_hosts, secrets, unifi_dns
from .debug import set_debug


CONTEXT_SETTINGS = {"help_option_names": ["-h", "--help"]}


def _new_subapp(help_text: str) -> typer.Typer:
    return typer.Typer(
        help=help_text,
        no_args_is_help=True,
        context_settings=CONTEXT_SETTINGS,
    )


def create_app() -> typer.Typer:
    app = typer.Typer(
        help=(
            "CLI for Nginx Proxy Manager and UniFi DNS automation.\n\n"
            "Preferred command styles:\n"
            "  npmctl proxy add ...\n"
            "  npmctl add proxy ...\n"
            "  npmctl secret set cloudflare-token\n"
        ),
        no_args_is_help=True,
        context_settings=CONTEXT_SETTINGS,
    )

    @app.callback()
    def main_callback(
        debug: bool = typer.Option(
            False,
            "--debug",
            help="Enable verbose request/response debug logging",
            envvar="NPM_CLI_DEBUG",
        ),
    ) -> None:
        """Global CLI options."""
        set_debug(debug)

    auth_app = _new_subapp("Manage authentication and saved NPM login state.")
    auth_app.command("login")(auth.login)
    auth_app.command("status")(auth.login_status)
    auth_app.command("logout")(auth.logout)
    app.add_typer(auth_app, name="auth")

    secret_app = _new_subapp("Store and inspect local secrets in your OS keyring.")
    secret_app.command("set")(secrets.set_secret_value)
    secret_app.command("status")(secrets.secret_status)
    secret_app.command("delete")(secrets.delete_secret_value)
    app.add_typer(secret_app, name="secret")

    site_app = _new_subapp("Inspect UniFi sites.")
    site_app.command("list")(unifi_dns.list_unifi_sites)
    app.add_typer(site_app, name="site")

    dns_app = _new_subapp("Manage UniFi DNS records.")
    dns_app.command("add")(unifi_dns.add_unifi_dns_record)
    app.add_typer(dns_app, name="dns")

    proxy_app = _new_subapp("Manage Nginx Proxy Manager proxy hosts.")
    proxy_app.command("add")(proxy_hosts.proxy_add)
    app.add_typer(proxy_app, name="proxy")

    cert_app = _new_subapp("Manage Nginx Proxy Manager certificates.")
    cert_app.command("add")(certificates.add_cert_cloudflare)
    app.add_typer(cert_app, name="cert")

    application_app = _new_subapp("Provision an application workflow across DNS and proxy layers.")
    application_app.command("add")(unifi_dns.add_new_app)
    app.add_typer(application_app, name="app")

    add_app = _new_subapp("Workflow shortcuts for creating resources.")
    add_app.command("proxy")(proxy_hosts.proxy_add)
    add_app.command("cert")(certificates.add_cert_cloudflare)
    add_app.command("dns")(unifi_dns.add_unifi_dns_record)
    add_app.command("app")(unifi_dns.add_new_app)
    app.add_typer(add_app, name="add")

    list_app = _new_subapp("Workflow shortcuts for listing resources.")
    list_app.command("sites")(unifi_dns.list_unifi_sites)
    app.add_typer(list_app, name="list")

    status_app = _new_subapp("Workflow shortcuts for inspecting saved state.")
    status_app.command("login")(auth.login_status)
    status_app.command("cloudflare-token")(secrets.cf_token_status)
    status_app.command("unifi-api-key")(secrets.unifi_api_key_status)
    app.add_typer(status_app, name="status")

    delete_app = _new_subapp("Workflow shortcuts for deleting saved secrets.")
    delete_app.command("cloudflare-token")(secrets.cf_token_delete)
    delete_app.command("unifi-api-key")(secrets.unifi_api_key_delete)
    app.add_typer(delete_app, name="delete")

    app.command("doctor")(doctor.doctor)

    app.command("login", hidden=True)(auth.login)
    app.command("login-status", hidden=True)(auth.login_status)
    app.command("logout", hidden=True)(auth.logout)
    app.command("cf-token-set", hidden=True)(secrets.cf_token_set)
    app.command("cf-token-status", hidden=True)(secrets.cf_token_status)
    app.command("cf-token-delete", hidden=True)(secrets.cf_token_delete)
    app.command("unifi-api-key-set", hidden=True)(secrets.unifi_api_key_set)
    app.command("unifi-api-key-status", hidden=True)(secrets.unifi_api_key_status)
    app.command("unifi-api-key-delete", hidden=True)(secrets.unifi_api_key_delete)
    app.command("list-unifi-sites", hidden=True)(unifi_dns.list_unifi_sites)
    app.command("add-unifi-dns-record", hidden=True)(unifi_dns.add_unifi_dns_record)
    app.command("add-new-app", hidden=True)(unifi_dns.add_new_app)
    app.command("add-cert-cloudflare", hidden=True)(certificates.add_cert_cloudflare)
    app.command("add-proxy-host", hidden=True)(proxy_hosts.add_proxy_host)
    app.command("add-proxy-with-cert", hidden=True)(proxy_hosts.add_proxy_with_cert)
    return app


app = create_app()


def main() -> None:
    app()
