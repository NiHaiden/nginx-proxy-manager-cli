from __future__ import annotations

from typing import Optional

import typer

from ..cli_helpers import build_client, exit_with_error
from ..errors import NPMError
from ..secrets_store import resolve_cloudflare_token


def register(app: typer.Typer) -> None:
    app.command("add-proxy-host")(add_proxy_host)
    app.command("add-proxy-with-cert")(add_proxy_with_cert)


def add_proxy_host(
    domain: list[str] = typer.Option(..., "--domain", "-d", help="Domain (repeatable)"),
    forward_host: str = typer.Option(..., help="Upstream host/IP"),
    forward_port: int = typer.Option(..., min=1, max=65535, help="Upstream port"),
    forward_scheme: str = typer.Option("http", help="Upstream scheme: http or https"),
    certificate_id: Optional[int] = typer.Option(None, help="Existing certificate ID"),
    ssl_forced: bool = typer.Option(False, help="Force HTTPS redirect"),
    http2_support: bool = typer.Option(False, help="Enable HTTP/2"),
    hsts_enabled: bool = typer.Option(False, help="Enable HSTS"),
    hsts_subdomains: bool = typer.Option(False, help="Apply HSTS to subdomains"),
    base_url: Optional[str] = typer.Option(None, help="API base URL"),
    token: Optional[str] = typer.Option(None, help="JWT token", envvar="NPM_TOKEN"),
) -> None:
    """Create a proxy host."""
    try:
        client = build_client(base_url, token)
        typer.echo("Creating proxy host...")
        host = client.create_proxy_host(
            domain_names=domain,
            forward_host=forward_host,
            forward_port=forward_port,
            forward_scheme=forward_scheme,
            certificate_id=certificate_id,
            ssl_forced=ssl_forced,
            http2_support=http2_support,
            hsts_enabled=hsts_enabled,
            hsts_subdomains=hsts_subdomains,
        )
        _print_proxy_host(host)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def add_proxy_with_cert(
    domain: list[str] = typer.Option(..., "--domain", "-d", help="Domain (repeatable)"),
    forward_host: str = typer.Option(..., help="Upstream host/IP"),
    forward_port: int = typer.Option(..., min=1, max=65535, help="Upstream port"),
    cloudflare_api_token: Optional[str] = typer.Option(
        None,
        help="Cloudflare API token",
        envvar="CLOUDFLARE_API_TOKEN",
    ),
    forward_scheme: str = typer.Option("http", help="Upstream scheme: http or https"),
    propagation_seconds: int = typer.Option(30, help="DNS propagation wait in seconds"),
    key_type: str = typer.Option("rsa", help="Certificate key type: rsa or ecdsa"),
    ssl_forced: bool = typer.Option(True, help="Force HTTPS redirect"),
    http2_support: bool = typer.Option(True, help="Enable HTTP/2"),
    hsts_enabled: bool = typer.Option(True, help="Enable HSTS"),
    hsts_subdomains: bool = typer.Option(True, help="Apply HSTS to subdomains"),
    base_url: Optional[str] = typer.Option(None, help="API base URL"),
    token: Optional[str] = typer.Option(None, help="JWT token", envvar="NPM_TOKEN"),
) -> None:
    """Create Cloudflare DNS cert, then create a proxy host with it."""
    try:
        client = build_client(base_url, token)
        resolved_cf_token = resolve_cloudflare_token(cloudflare_api_token)
        cert_id = _create_cloudflare_cert(
            client,
            domain,
            resolved_cf_token,
            propagation_seconds,
            key_type,
        )
        host = _create_proxy_host_with_cert(
            client,
            domain,
            forward_host,
            forward_port,
            forward_scheme,
            cert_id,
            ssl_forced,
            http2_support,
            hsts_enabled,
            hsts_subdomains,
        )
        _print_combined_success(cert_id, host)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def _create_cloudflare_cert(
    client,
    domain: list[str],
    cloudflare_api_token: str,
    propagation_seconds: int,
    key_type: str,
) -> int:
    typer.echo("Step 1/2: Creating certificate via Cloudflare DNS challenge...")
    cert = client.create_cloudflare_certificate(
        domain_names=domain,
        cloudflare_api_token=cloudflare_api_token,
        propagation_seconds=propagation_seconds,
        key_type=key_type,
    )

    cert_id = cert.get("id")
    if not cert_id:
        raise NPMError("Certificate creation succeeded but no certificate id was returned.")

    typer.echo(f"Certificate created (id={cert_id}).")
    return int(cert_id)


def _create_proxy_host_with_cert(
    client,
    domain: list[str],
    forward_host: str,
    forward_port: int,
    forward_scheme: str,
    cert_id: int,
    ssl_forced: bool,
    http2_support: bool,
    hsts_enabled: bool,
    hsts_subdomains: bool,
) -> dict:
    typer.echo("Step 2/2: Creating proxy host with that certificate...")
    return client.create_proxy_host(
        domain_names=domain,
        forward_host=forward_host,
        forward_port=forward_port,
        forward_scheme=forward_scheme,
        certificate_id=cert_id,
        ssl_forced=ssl_forced,
        http2_support=http2_support,
        hsts_enabled=hsts_enabled,
        hsts_subdomains=hsts_subdomains,
    )


def _print_proxy_host(host: dict) -> None:
    typer.secho("Proxy host created successfully.", fg=typer.colors.GREEN)
    typer.echo(f"Host ID: {host.get('id')}")
    typer.echo(f"Domains: {', '.join(host.get('domain_names', []))}")


def _print_combined_success(cert_id: int, host: dict) -> None:
    typer.secho("Certificate and proxy host created successfully.", fg=typer.colors.GREEN)
    typer.echo(f"Certificate ID: {cert_id}")
    typer.echo(f"Proxy Host ID: {host.get('id')}")
    typer.echo(f"Domains: {', '.join(host.get('domain_names', []))}")
