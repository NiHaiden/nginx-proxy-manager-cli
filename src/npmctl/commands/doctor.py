from __future__ import annotations

import platform
import sys
from pathlib import Path
from typing import Any, Optional

import typer


def register(app: typer.Typer) -> None:
    app.command()(doctor)


def doctor() -> None:
    """Run environment diagnostics and keyring backend checks."""
    issues = _run_doctor()
    if issues:
        raise typer.Exit(code=1)


def _run_doctor() -> list[str]:
    issues: list[str] = []

    typer.secho("npmctl doctor", fg=typer.colors.CYAN, bold=True)
    typer.echo(f"Platform: {platform.system()} {platform.release()}")
    typer.echo(f"Python: {sys.version.split()[0]} ({sys.executable})")

    distro_info = _load_linux_distro_info()
    distro_label = _distro_label(distro_info)
    if distro_label:
        typer.echo(f"Linux distro: {distro_label}")

    keyring, keyring_backend, keyring_core, keyring_error = _load_keyring_modules()
    if keyring_error is not None:
        issues.append(
            "Python package `keyring` is not importable in this environment "
            f"({keyring_error})."
        )
    else:
        typer.secho("✔ keyring Python package importable", fg=typer.colors.GREEN)

        current_backend = keyring.get_keyring()
        current_backend_name = _backend_name(current_backend)
        typer.echo(f"Current keyring backend: {current_backend_name}")

        current_recommended = _is_recommended_backend(keyring_core, current_backend)
        available_recommended = [
            _backend_name(backend)
            for backend in keyring_backend.get_all_keyring()
            if _is_recommended_backend(keyring_core, backend)
        ]

        if current_recommended:
            typer.secho("✔ current backend is recommended", fg=typer.colors.GREEN)
        else:
            issues.append(
                "Current keyring backend is not recommended. Secure token storage may fail."
            )

        if available_recommended:
            typer.secho("✔ at least one recommended keyring backend is available", fg=typer.colors.GREEN)
        else:
            issues.append("No recommended keyring backend is available.")

        if available_recommended:
            typer.echo("Recommended backends detected:")
            for backend_name in sorted(set(available_recommended)):
                typer.echo(f"  - {backend_name}")

    if issues:
        typer.secho("\nIssues found:", fg=typer.colors.RED, bold=True)
        for issue in issues:
            typer.echo(f"  - {issue}")

        _print_fix_hints(distro_info)
    else:
        typer.secho("\nAll checks passed.", fg=typer.colors.GREEN, bold=True)

    return issues


def _load_keyring_modules() -> tuple[Any, Any, Any, Optional[Exception]]:
    try:
        import keyring  # type: ignore
        import keyring.backend as keyring_backend  # type: ignore
        import keyring.core as keyring_core  # type: ignore

        return keyring, keyring_backend, keyring_core, None
    except Exception as exc:  # noqa: BLE001
        return None, None, None, exc


def _is_recommended_backend(keyring_core: Any, backend: Any) -> bool:
    try:
        return bool(keyring_core.recommended(backend))
    except Exception:  # noqa: BLE001
        return False


def _backend_name(backend: Any) -> str:
    module = backend.__class__.__module__
    name = backend.__class__.__name__
    return f"{module}.{name}"


def _load_linux_distro_info() -> dict[str, str]:
    if platform.system().lower() != "linux":
        return {}

    for path in (Path("/etc/os-release"), Path("/usr/lib/os-release")):
        if path.exists():
            return _parse_os_release(path)
    return {}


def _parse_os_release(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue

        key, value = line.split("=", 1)
        value = value.strip().strip('"').strip("'")
        values[key] = value

    return values


def _distro_label(distro_info: dict[str, str]) -> str:
    if not distro_info:
        return ""

    pretty = distro_info.get("PRETTY_NAME")
    if pretty:
        return pretty

    distro_id = distro_info.get("ID", "")
    version_id = distro_info.get("VERSION_ID", "")
    if distro_id and version_id:
        return f"{distro_id} {version_id}"
    return distro_id


def _print_fix_hints(distro_info: dict[str, str]) -> None:
    typer.secho("\nHow to fix keyring backend issues:", fg=typer.colors.YELLOW, bold=True)

    distro_id = distro_info.get("ID", "").lower()
    distro_like = distro_info.get("ID_LIKE", "").lower()

    if _is_fedora_like(distro_id, distro_like):
        typer.echo("Fedora/RHEL family (recommended backend):")
        typer.echo("  sudo dnf install -y gnome-keyring libsecret")
    elif _is_arch_like(distro_id, distro_like):
        typer.echo("Arch family (recommended backend):")
        typer.echo("  sudo pacman -S --needed gnome-keyring libsecret")
    elif _is_debian_like(distro_id, distro_like):
        typer.echo("Ubuntu/Debian family (recommended backend):")
        typer.echo("  sudo apt update && sudo apt install -y gnome-keyring libsecret-1-0 dbus-user-session")
    else:
        typer.echo("Install and run an OS keyring backend (Secret Service or KWallet).")

    typer.echo(
        "Then run `npmctl login-status` again from your normal user session "
        "(not a minimal headless shell without DBus)."
    )

    typer.echo("\nFallback (less secure):")
    typer.echo(f"  {sys.executable} -m pip install keyrings.alt")


def _is_fedora_like(distro_id: str, distro_like: str) -> bool:
    tokens = {distro_id, *distro_like.split()}
    return bool(tokens & {"fedora", "rhel", "centos", "rocky", "almalinux"})


def _is_arch_like(distro_id: str, distro_like: str) -> bool:
    tokens = {distro_id, *distro_like.split()}
    return bool(tokens & {"arch", "manjaro", "endeavouros"})


def _is_debian_like(distro_id: str, distro_like: str) -> bool:
    tokens = {distro_id, *distro_like.split()}
    return bool(tokens & {"ubuntu", "debian", "linuxmint", "pop"})
