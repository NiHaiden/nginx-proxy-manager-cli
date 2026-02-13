from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Any, Optional

from .debug import log
from .errors import NPMError
from .secrets_store import get_login_info, set_login_info

LEGACY_CONFIG_PATH = Path("~/.config/npm-cli/config.json").expanduser()
LEGACY_NPMCTL_CONFIG_PATH = Path(
    os.environ.get("NPM_CLI_CONFIG", "~/.config/npmctl/config.json")
).expanduser()


def save_login(base_url: str, token: str, expires: Optional[str], identity: str) -> None:
    set_login_info(base_url=base_url, token=token, expires=expires, identity=identity)


def resolve_base_url_and_token(
    base_url: Optional[str],
    token: Optional[str],
    *,
    require_token: bool = True,
) -> tuple[str, Optional[str]]:
    keyring_login = _load_keyring_login_info()
    legacy_login = _load_legacy_login_info()

    resolved_base_url = (
        base_url
        or os.environ.get("NPM_BASE_URL")
        or keyring_login.get("base_url")
        or legacy_login.get("base_url")
    )
    if not resolved_base_url:
        raise NPMError("Missing NPM base URL. Use --base-url or set NPM_BASE_URL.")

    resolved_token = (
        token or os.environ.get("NPM_TOKEN") or keyring_login.get("token") or legacy_login.get("token")
    )
    if require_token and not resolved_token:
        raise NPMError("Missing API token. Run `npmctl login` or set NPM_TOKEN.")

    log(f"Using base URL: {resolved_base_url}")
    log(f"Using token: {'yes' if resolved_token else 'no'}")
    return resolved_base_url, resolved_token


def _load_keyring_login_info() -> dict[str, Any]:
    try:
        return get_login_info()
    except NPMError as exc:
        log(f"Keyring login lookup skipped: {exc}")
        return {}


def _load_legacy_login_info() -> dict[str, Any]:
    for path in (LEGACY_NPMCTL_CONFIG_PATH, LEGACY_CONFIG_PATH):
        loaded = _load_json_file(path)
        if loaded:
            return loaded
    return {}


def _load_json_file(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}

    try:
        parsed = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return {}

    if isinstance(parsed, dict):
        return parsed
    return {}
