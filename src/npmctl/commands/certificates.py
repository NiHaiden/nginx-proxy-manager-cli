from __future__ import annotations

from typing import Optional

import typer

from ..cli_helpers import build_client, exit_with_error
from ..secrets_store import resolve_cloudflare_token


def register(app: typer.Typer) -> None:
    app.command("add-cert-cloudflare")(add_cert_cloudflare)


def add_cert_cloudflare(
    domain: list[str] = typer.Option(..., "--domain", "-d", help="Domain (repeatable)"),
    cloudflare_api_token: Optional[str] = typer.Option(
        None,
        help="Cloudflare API token",
        envvar="CLOUDFLARE_API_TOKEN",
    ),
    nice_name: Optional[str] = typer.Option(None, help="Human-friendly certificate name"),
    propagation_seconds: int = typer.Option(30, help="DNS propagation wait in seconds"),
    key_type: str = typer.Option("rsa", help="Certificate key type: rsa or ecdsa"),
    base_url: Optional[str] = typer.Option(None, help="API base URL"),
    token: Optional[str] = typer.Option(None, help="JWT token", envvar="NPM_TOKEN"),
) -> None:
    """Create a Let's Encrypt certificate using Cloudflare DNS challenge."""
    try:
        client = build_client(base_url, token)
        resolved_cf_token = resolve_cloudflare_token(cloudflare_api_token)
        typer.echo("Creating certificate via Cloudflare DNS challenge...")
        cert = client.create_cloudflare_certificate(
            domain_names=domain,
            cloudflare_api_token=resolved_cf_token,
            nice_name=nice_name,
            propagation_seconds=propagation_seconds,
            key_type=key_type,
        )
        _print_certificate(cert)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def _print_certificate(cert: dict) -> None:
    typer.secho("Certificate created successfully.", fg=typer.colors.GREEN)
    typer.echo(f"Certificate ID: {cert.get('id')}")
    typer.echo(f"Name: {cert.get('nice_name')}")
    typer.echo(f"Domains: {', '.join(cert.get('domain_names', []))}")
