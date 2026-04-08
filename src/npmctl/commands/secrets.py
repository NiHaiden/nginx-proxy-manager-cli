from __future__ import annotations

from enum import Enum
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


class SecretName(str, Enum):
    CLOUDFLARE_TOKEN = "cloudflare-token"
    UNIFI_API_KEY = "unifi-api-key"


def set_secret_value(
    secret_name: SecretName = typer.Argument(..., help="Secret to store"),
    value: Optional[str] = typer.Option(
        None,
        "--value",
        prompt=True,
        hide_input=True,
        confirmation_prompt=True,
        help="Secret value",
    ),
) -> None:
    """Store a secret securely in OS keyring."""
    try:
        spec = _secret_spec(secret_name)
        set_secret(spec["key"], value or "")
        typer.secho(spec["saved_message"], fg=typer.colors.GREEN)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def delete_secret_value(
    secret_name: SecretName = typer.Argument(..., help="Secret to delete"),
) -> None:
    """Delete a stored secret from OS keyring."""
    try:
        spec = _secret_spec(secret_name)
        delete_secret(spec["key"])
        typer.secho(spec["deleted_message"], fg=typer.colors.GREEN)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def secret_status(
    secret_name: SecretName = typer.Argument(..., help="Secret to inspect"),
) -> None:
    """Show whether a secret is stored in keyring."""
    try:
        spec = _secret_spec(secret_name)
        exists = bool(get_secret(spec["key"]))
        if exists:
            typer.secho(spec["present_message"], fg=typer.colors.GREEN)
        else:
            typer.secho(spec["missing_message"], fg=typer.colors.YELLOW)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


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


def _secret_spec(secret_name: SecretName) -> dict[str, str]:
    if secret_name == SecretName.CLOUDFLARE_TOKEN:
        return {
            "key": CF_TOKEN_KEY,
            "saved_message": "Cloudflare token saved in OS keyring.",
            "deleted_message": "Cloudflare token deleted from OS keyring.",
            "present_message": "Cloudflare token is stored in keyring.",
            "missing_message": "No Cloudflare token stored in keyring.",
        }

    return {
        "key": UNIFI_API_KEY_KEY,
        "saved_message": "UniFi API key saved in OS keyring.",
        "deleted_message": "UniFi API key deleted from OS keyring.",
        "present_message": "UniFi API key is stored in keyring.",
        "missing_message": "No UniFi API key stored in keyring.",
    }
