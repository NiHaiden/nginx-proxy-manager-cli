from __future__ import annotations

import unittest

from typer.testing import CliRunner

from npmctl.cli import create_app


class CliSurfaceTests(unittest.TestCase):
    def setUp(self) -> None:
        self.runner = CliRunner()
        self.app = create_app()

    def test_root_help_shows_grouped_commands_and_hides_legacy_flat_names(self) -> None:
        result = self.runner.invoke(self.app, ["--help"])
        self.assertEqual(result.exit_code, 0)
        self.assertIn("auth", result.output)
        self.assertIn("secret", result.output)
        self.assertIn("proxy", result.output)
        self.assertIn("add", result.output)
        self.assertNotIn("add-proxy-host", result.output)
        self.assertNotIn("login-status", result.output)

    def test_add_help_lists_creation_shortcuts(self) -> None:
        result = self.runner.invoke(self.app, ["add", "--help"])
        self.assertEqual(result.exit_code, 0)
        self.assertIn("proxy", result.output)
        self.assertIn("cert", result.output)
        self.assertIn("dns", result.output)
        self.assertIn("app", result.output)

    def test_proxy_add_help_includes_optional_certificate_workflow(self) -> None:
        result = self.runner.invoke(self.app, ["proxy", "add", "--help"])
        self.assertEqual(result.exit_code, 0)
        self.assertIn("--create-cert", result.output)
        self.assertIn("Existing", result.output)
        self.assertIn("certificate ID", result.output)

    def test_secret_set_help_uses_generic_secret_names(self) -> None:
        result = self.runner.invoke(self.app, ["secret", "set", "--help"])
        self.assertEqual(result.exit_code, 0)
        self.assertIn("cloudflare-token", result.output)
        self.assertIn("unifi-api-key", result.output)

    def test_hidden_legacy_aliases_still_exist(self) -> None:
        result = self.runner.invoke(self.app, ["add-proxy-host", "--help"])
        self.assertEqual(result.exit_code, 0)
        self.assertIn("--forward-host", result.output)


if __name__ == "__main__":
    unittest.main()
