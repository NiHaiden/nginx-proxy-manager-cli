from __future__ import annotations

from typing import Optional

import typer

from ..cli_helpers import exit_with_error
from ..secrets_store import (
    CF_TOKEN_KEY,
    UNIFI_API_KEY_KEY,
    delete_secret,
    get_secret,
    set_secret,
)


def register(app: typer.Typer) -> None:
    app.command("cf-token-set")(cf_token_set)
    app.command("cf-token-delete")(cf_token_delete)
    app.command("cf-token-status")(cf_token_status)
    app.command("unifi-api-key-set")(unifi_api_key_set)
    app.command("unifi-api-key-delete")(unifi_api_key_delete)
    app.command("unifi-api-key-status")(unifi_api_key_status)


def cf_token_set(
    token: Optional[str] = typer.Option(
        None,
        "--token",
        help="Cloudflare API token",
        prompt=True,
        hide_input=True,
        confirmation_prompt=True,
    ),
) -> None:
    """Store Cloudflare API token securely in OS keyring."""
    try:
        set_secret(CF_TOKEN_KEY, token or "")
        typer.secho("Cloudflare token saved in OS keyring.", fg=typer.colors.GREEN)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def cf_token_delete() -> None:
    """Delete Cloudflare API token from OS keyring."""
    try:
        delete_secret(CF_TOKEN_KEY)
        typer.secho("Cloudflare token deleted from OS keyring.", fg=typer.colors.GREEN)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def cf_token_status() -> None:
    """Show whether a Cloudflare token is stored in keyring."""
    try:
        exists = bool(get_secret(CF_TOKEN_KEY))
        if exists:
            typer.secho("Cloudflare token is stored in keyring.", fg=typer.colors.GREEN)
        else:
            typer.secho("No Cloudflare token stored in keyring.", fg=typer.colors.YELLOW)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def unifi_api_key_set(
    api_key: Optional[str] = typer.Option(
        None,
        "--api-key",
        help="UniFi integration API key",
        prompt=True,
        hide_input=True,
        confirmation_prompt=True,
    ),
) -> None:
    """Store UniFi integration API key securely in OS keyring."""
    try:
        set_secret(UNIFI_API_KEY_KEY, api_key or "")
        typer.secho("UniFi API key saved in OS keyring.", fg=typer.colors.GREEN)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def unifi_api_key_delete() -> None:
    """Delete UniFi integration API key from OS keyring."""
    try:
        delete_secret(UNIFI_API_KEY_KEY)
        typer.secho("UniFi API key deleted from OS keyring.", fg=typer.colors.GREEN)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def unifi_api_key_status() -> None:
    """Show whether a UniFi API key is stored in keyring."""
    try:
        exists = bool(get_secret(UNIFI_API_KEY_KEY))
        if exists:
            typer.secho("UniFi API key is stored in keyring.", fg=typer.colors.GREEN)
        else:
            typer.secho("No UniFi API key stored in keyring.", fg=typer.colors.YELLOW)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)
