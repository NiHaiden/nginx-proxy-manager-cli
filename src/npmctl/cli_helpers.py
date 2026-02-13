from __future__ import annotations

import traceback
from typing import Optional

import typer

from .client import NPMClient
from .config import resolve_base_url_and_token
from .debug import is_debug, log


def build_client(
    base_url: Optional[str],
    token: Optional[str],
    *,
    require_token: bool = True,
) -> NPMClient:
    resolved_base_url, resolved_token = resolve_base_url_and_token(
        base_url,
        token,
        require_token=require_token,
    )
    return NPMClient(base_url=resolved_base_url, token=resolved_token)


def exit_with_error(exc: Exception) -> None:
    typer.secho(str(exc), fg=typer.colors.RED, err=True)
    if is_debug():
        log("Stack trace:")
        typer.secho(traceback.format_exc(), fg=typer.colors.BRIGHT_BLACK, err=True)
    raise typer.Exit(code=1)
