from __future__ import annotations

import unittest
from typing import Optional
from unittest.mock import patch

from typer.testing import CliRunner

from npmctl import secrets_store
from npmctl.cli import create_app


class _FakeBackend:
    pass


class _FakeKeyring:
    def get_keyring(self) -> _FakeBackend:
        return _FakeBackend()


class _FakeKeyringBackend:
    def get_all_keyring(self) -> list[_FakeBackend]:
        return [_FakeBackend()]


class _FakeKeyringCore:
    def recommended(self, backend: _FakeBackend) -> bool:
        return True


class DoctorSecretStatusTests(unittest.TestCase):
    def setUp(self) -> None:
        self.runner = CliRunner()
        self.app = create_app()

    def test_doctor_reports_missing_secrets_without_failing(self) -> None:
        with _patched_keyring_modules():
            with patch("npmctl.commands.doctor.get_login_info", return_value={}):
                with patch("npmctl.commands.doctor.get_secret", return_value=None):
                    result = self.runner.invoke(self.app, ["doctor"])

        self.assertEqual(result.exit_code, 0)
        self.assertIn("NPM login token: not stored", result.output)
        self.assertIn("Add it with `npmctl auth login`.", result.output)
        self.assertIn("Cloudflare token: not stored", result.output)
        self.assertIn("Add it with `npmctl secret set cloudflare-token`.", result.output)
        self.assertIn("UniFi API key: not stored", result.output)
        self.assertIn("Add it with `npmctl secret set unifi-api-key`.", result.output)

    def test_doctor_reports_stored_secrets(self) -> None:
        def get_secret(secret_name: str) -> Optional[str]:
            stored = {
                secrets_store.CF_TOKEN_KEY: "cf-token",
                secrets_store.UNIFI_API_KEY_KEY: "unifi-key",
            }
            return stored.get(secret_name)

        with _patched_keyring_modules():
            with patch(
                "npmctl.commands.doctor.get_login_info",
                return_value={"token": "npm-token"},
            ):
                with patch("npmctl.commands.doctor.get_secret", side_effect=get_secret):
                    result = self.runner.invoke(self.app, ["doctor"])

        self.assertEqual(result.exit_code, 0)
        self.assertIn("NPM login token: stored", result.output)
        self.assertIn("Cloudflare token: stored", result.output)
        self.assertIn("UniFi API key: stored", result.output)


def _patched_keyring_modules():
    return patch(
        "npmctl.commands.doctor._load_keyring_modules",
        return_value=(
            _FakeKeyring(),
            _FakeKeyringBackend(),
            _FakeKeyringCore(),
            None,
        ),
    )


if __name__ == "__main__":
    unittest.main()
