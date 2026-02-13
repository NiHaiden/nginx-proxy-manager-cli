from __future__ import annotations

from typing import Optional

import typer

from ..cli_helpers import build_client, exit_with_error
from ..config import save_login
from ..errors import NPMError
from ..secrets_store import clear_login_info, get_login_info


def register(app: typer.Typer) -> None:
    app.command()(login)
    app.command()(login_status)
    app.command()(logout)


def login(
    identity: str = typer.Option(..., help="NPM user email"),
    secret: str = typer.Option(..., prompt=True, hide_input=True, help="NPM password"),
    base_url: Optional[str] = typer.Option(None, help="API base URL (e.g. http://host/api)"),
    scope: str = typer.Option("user", help="Token scope"),
) -> None:
    """Login and save your JWT token in OS keyring."""
    try:
        client = build_client(base_url, None, require_token=False)
        data = client.request_token(identity=identity, secret=secret, scope=scope)
        token, expires = _extract_token_data(data)
        save_login(client.base_url, token, expires, identity)
        _print_login_success(expires)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def login_status() -> None:
    """Show whether login info is stored and display non-sensitive details."""
    try:
        info = get_login_info()
        if not info:
            typer.secho("No stored login found in OS keyring.", fg=typer.colors.YELLOW)
            return

        identity = info.get("identity") or "(unknown)"
        base_url = info.get("base_url") or "(unknown)"
        expires = info.get("expires") or "(unknown)"

        typer.secho("Stored login found in OS keyring.", fg=typer.colors.GREEN)
        typer.echo(f"Identity: {identity}")
        typer.echo(f"Base URL: {base_url}")
        typer.echo(f"Token expires: {expires}")
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def logout() -> None:
    """Remove stored login token from OS keyring."""
    try:
        clear_login_info()
        typer.secho("Stored login removed from OS keyring.", fg=typer.colors.GREEN)
    except Exception as exc:  # noqa: BLE001
        exit_with_error(exc)


def _extract_token_data(data: dict) -> tuple[str, Optional[str]]:
    if data.get("requires_2fa"):
        raise NPMError("2FA required. Use /tokens/2fa flow manually for now.")

    token = data.get("token")
    expires = data.get("expires")
    if not token:
        raise NPMError("Login response did not include a token.")

    return token, expires


def _print_login_success(expires: Optional[str]) -> None:
    typer.secho("Login successful. Token saved in OS keyring.", fg=typer.colors.GREEN)
    if expires:
        typer.echo(f"Token expires: {expires}")
