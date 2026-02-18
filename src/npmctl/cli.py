from __future__ import annotations

import typer

from .commands import auth, certificates, doctor, proxy_hosts, secrets
from .debug import set_debug


def create_app() -> typer.Typer:
    app = typer.Typer(
        help="CLI for managing Nginx Proxy Manager proxy hosts and certificates",
        context_settings={"help_option_names": ["-h", "--help"]},
    )

    @app.callback()
    def main_callback(
        debug: bool = typer.Option(
            False,
            "--debug",
            help="Enable verbose request/response debug logging",
            envvar="NPM_CLI_DEBUG",
        ),
    ) -> None:
        """Global CLI options."""
        set_debug(debug)

    auth.register(app)
    certificates.register(app)
    proxy_hosts.register(app)
    secrets.register(app)
    doctor.register(app)
    return app


app = create_app()


def main() -> None:
    app()
