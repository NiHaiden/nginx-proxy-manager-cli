from __future__ import annotations

import json
from typing import Any

import typer

_DEBUG = False
SENSITIVE_KEYS = {
    "authorization",
    "token",
    "secret",
    "password",
    "dns_provider_credentials",
    "cloudflare_api_token",
}


def set_debug(enabled: bool) -> None:
    global _DEBUG
    _DEBUG = enabled


def is_debug() -> bool:
    return _DEBUG


def log(message: str) -> None:
    if _DEBUG:
        typer.secho(f"[debug] {message}", fg=typer.colors.BRIGHT_BLACK, err=True)


def mask_secret(value: str) -> str:
    if not value or len(value) <= 8:
        return "***"
    return f"{value[:4]}...{value[-4:]}"


def _is_sensitive_key(key_hint: str) -> bool:
    hint = key_hint.lower()
    return any(secret_key in hint for secret_key in SENSITIVE_KEYS)


def redact_value(value: Any, key_hint: str = "") -> Any:
    if _is_sensitive_key(key_hint):
        if isinstance(value, str):
            return mask_secret(value)
        return "***"

    if isinstance(value, dict):
        return {k: redact_value(v, k) for k, v in value.items()}

    if isinstance(value, list):
        return [redact_value(item, key_hint) for item in value]

    if isinstance(value, tuple):
        return tuple(redact_value(item, key_hint) for item in value)

    return value


def preview(value: Any, limit: int = 1600) -> str:
    text = _to_text(value)
    if len(text) <= limit:
        return text
    hidden = len(text) - limit
    return f"{text[:limit]}... [truncated {hidden} chars]"


def _to_text(value: Any) -> str:
    if isinstance(value, (dict, list, tuple)):
        return json.dumps(value, indent=2, ensure_ascii=False)
    return str(value)
