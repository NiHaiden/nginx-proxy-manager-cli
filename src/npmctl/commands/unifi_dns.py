from __future__ import annotations

from typing import Any, Optional

import typer

from ..cli_helpers import exit_with_error
from ..client import NPMClient
from ..errors import NPMError
from ..secrets_store import resolve_unifi_api_key


def register(app: typer.Typer) -> None:
    app.command("list-unifi-sites")(list_unifi_sites)
    app.command("add-unifi-dns-record")(add_unifi_dns_record)


def list_unifi_sites(
    gateway_url: str = typer.Option(
        ...,
        "--gateway-url",
        envvar="UNIFI_BASE_URL",
        help="UniFi gateway URL (e.g. https://192.168.1.1)",
    ),
    unifi_api_key: Optional[str] = typer.Option(
        None,
        "--unifi-api-key",
        envvar="UNIFI_API_KEY",
        help="UniFi integration API key",
    ),
) -> None:
    """List UniFi Network site IDs for DNS policy operations."""
    try:
        client = NPMClient(base_url=gateway_url, token=None)
        resolved_api_key = resolve_unifi_api_key(unifi_api_key)
        payload = client.list_unifi_sites(unifi_api_key=resolved_api_key)
        _print_sites(payload)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def add_unifi_dns_record(
    site_id: Optional[str] = typer.Option(
        None,
        "--site-id",
        help="UniFi site ID (optional when exactly one site exists)",
    ),
    domain: str = typer.Option(..., "--domain", "-d", help="DNS name to create"),
    ipv4_address: str = typer.Option(..., "--ipv4-address", help="IPv4 address for the A record"),
    ttl_seconds: int = typer.Option(14400, "--ttl-seconds", min=1, help="TTL in seconds"),
    enabled: bool = typer.Option(True, "--enabled/--disabled", help="Enable or disable record"),
    record_type: str = typer.Option("A_RECORD", "--record-type", help="Policy type, e.g. A_RECORD"),
    gateway_url: str = typer.Option(
        ...,
        "--gateway-url",
        envvar="UNIFI_BASE_URL",
        help="UniFi gateway URL (e.g. https://192.168.1.1)",
    ),
    unifi_api_key: Optional[str] = typer.Option(
        None,
        "--unifi-api-key",
        envvar="UNIFI_API_KEY",
        help="UniFi integration API key",
    ),
) -> None:
    """Create a UniFi DNS policy record using the integration API."""
    try:
        client = NPMClient(base_url=gateway_url, token=None)
        resolved_api_key = resolve_unifi_api_key(unifi_api_key)
        selected_site_id = _resolve_site_id(client, resolved_api_key, site_id)

        typer.echo("Creating UniFi DNS record...")
        payload = client.create_unifi_dns_record(
            unifi_api_key=resolved_api_key,
            site_id=selected_site_id,
            domain=domain,
            ipv4_address=ipv4_address,
            ttl_seconds=ttl_seconds,
            enabled=enabled,
            record_type=record_type,
        )

        _print_dns_record(payload, domain, selected_site_id)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def _extract_sites(payload: Any) -> list[dict[str, Any]]:
    if isinstance(payload, list):
        return [item for item in payload if isinstance(item, dict)]

    if isinstance(payload, dict):
        for key in ("sites", "data", "items", "results"):
            value = payload.get(key)
            if isinstance(value, list):
                return [item for item in value if isinstance(item, dict)]

    return []


def _resolve_site_id(client: NPMClient, unifi_api_key: str, explicit_site_id: Optional[str]) -> str:
    if explicit_site_id:
        return explicit_site_id

    payload = client.list_unifi_sites(unifi_api_key=unifi_api_key)
    sites = _extract_sites(payload)
    selected_site_id, selected_site = _resolve_site_id_from_sites(sites)
    typer.echo(f"Auto-selected site ID: {selected_site_id} ({_site_label(selected_site)})")
    return selected_site_id


def _resolve_site_id_from_sites(sites: list[dict[str, Any]]) -> tuple[str, dict[str, Any]]:
    if not sites:
        raise NPMError(
            "No UniFi sites returned. Pass --site-id explicitly or verify your API key and gateway URL."
        )

    if len(sites) > 1:
        choices = "\n".join(_format_site_choice(site) for site in sites)
        raise NPMError(
            "Multiple UniFi sites found. Pass --site-id explicitly.\n"
            f"Available sites:\n{choices}"
        )

    site = sites[0]
    selected_site_id = _site_id(site)
    if not selected_site_id:
        raise NPMError(
            "UniFi returned one site but without a usable site ID. Pass --site-id explicitly."
        )

    return selected_site_id, site


def _format_site_choice(site: dict[str, Any]) -> str:
    label = _site_label(site)
    site_id = _site_id(site) or "(missing id)"
    return f"- {label}: {site_id}"


def _site_label(site: dict[str, Any]) -> str:
    name = site.get("name") or site.get("desc") or site.get("description")
    if isinstance(name, str) and name:
        return name
    return "(unnamed)"


def _site_id(site: dict[str, Any]) -> str:
    candidates = ("id", "_id", "siteId")
    for key in candidates:
        value = site.get(key)
        if isinstance(value, str) and value:
            return value
    return ""


def _print_sites(payload: Any) -> None:
    sites = _extract_sites(payload)
    if not sites:
        typer.secho("No UniFi sites returned.", fg=typer.colors.YELLOW)
        return

    typer.secho("UniFi sites:", fg=typer.colors.GREEN)
    for site in sites:
        site_id = _site_id(site)
        label = _site_label(site)
        if site_id:
            typer.echo(f"- {label}: {site_id}")
        else:
            typer.echo(f"- {label}")


def _print_dns_record(payload: Any, domain: str, site_id: str) -> None:
    typer.secho("UniFi DNS record created successfully.", fg=typer.colors.GREEN)
    typer.echo(f"Domain: {domain}")
    typer.echo(f"Site ID: {site_id}")

    if isinstance(payload, dict):
        record_id = payload.get("id") or payload.get("_id")
        if record_id:
            typer.echo(f"Record ID: {record_id}")
