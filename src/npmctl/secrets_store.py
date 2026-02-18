from __future__ import annotations

import json
from typing import Any, Optional

from .errors import NPMError

SERVICE_NAME = "npmctl"
LEGACY_SERVICE_NAME = "npm-cli"

CF_TOKEN_KEY = "cloudflare_api_token"
LOGIN_INFO_KEY = "login_info"

try:
    import keyring
    from keyring.errors import KeyringError
except Exception:  # pragma: no cover
    keyring = None
    KeyringError = Exception


def _require_keyring() -> None:
    if keyring is None:
        raise NPMError(
            "Secure storage backend unavailable. Install `keyring` and an OS keyring backend. "
            "Fallback: install `keyrings.alt` in the npmctl environment "
            "(for example `~/.npmctl/venv/bin/pip install keyrings.alt`)."
        )


def _keyring_backend_hint(exc: Exception) -> str:
    message = str(exc).lower()
    if "no recommended backend was available" in message:
        return (
            " Hint: install and run an OS keyring backend (Secret Service/KWallet), "
            "or install `keyrings.alt` in the npmctl environment "
            "(for example `~/.npmctl/venv/bin/pip install keyrings.alt`)."
        )
    return ""


def get_secret(secret_name: str) -> Optional[str]:
    _require_keyring()
    try:
        value = keyring.get_password(SERVICE_NAME, secret_name)
        if value:
            return value
        return keyring.get_password(LEGACY_SERVICE_NAME, secret_name)
    except KeyringError as exc:
        raise NPMError(f"Failed to read secret from keyring: {exc}{_keyring_backend_hint(exc)}") from exc


def set_secret(secret_name: str, value: str) -> None:
    _require_keyring()
    try:
        keyring.set_password(SERVICE_NAME, secret_name, value)
    except KeyringError as exc:
        raise NPMError(f"Failed to store secret in keyring: {exc}{_keyring_backend_hint(exc)}") from exc


def delete_secret(secret_name: str) -> None:
    _require_keyring()
    _delete_if_exists(SERVICE_NAME, secret_name)
    _delete_if_exists(LEGACY_SERVICE_NAME, secret_name)


def _delete_if_exists(service_name: str, secret_name: str) -> None:
    try:
        keyring.delete_password(service_name, secret_name)
    except KeyringError as exc:
        message = str(exc).lower()
        if "not found" in message:
            return
        raise NPMError(f"Failed to delete secret from keyring: {exc}{_keyring_backend_hint(exc)}") from exc


def get_login_info() -> dict[str, Any]:
    raw = get_secret(LOGIN_INFO_KEY)
    if not raw:
        return {}

    try:
        parsed = json.loads(raw)
    except json.JSONDecodeError:
        return {}

    if isinstance(parsed, dict):
        return parsed
    return {}


def set_login_info(base_url: str, token: str, expires: Optional[str], identity: str) -> None:
    payload = {
        "base_url": base_url,
        "token": token,
        "expires": expires,
        "identity": identity,
    }
    set_secret(LOGIN_INFO_KEY, json.dumps(payload))


def clear_login_info() -> None:
    delete_secret(LOGIN_INFO_KEY)


def resolve_cloudflare_token(cli_or_env_token: Optional[str]) -> str:
    if cli_or_env_token:
        return cli_or_env_token

    token = get_secret(CF_TOKEN_KEY)
    if token:
        return token

    raise NPMError(
        "Missing Cloudflare API token. Use --cloudflare-api-token, set CLOUDFLARE_API_TOKEN, "
        "or run `npmctl cf-token-set`."
    )
