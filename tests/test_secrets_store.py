from __future__ import annotations

import unittest
from unittest.mock import patch

from npmctl.errors import NPMError
from npmctl import secrets_store


class ResolveUnifiApiKeyTests(unittest.TestCase):
    def test_prefers_explicit_cli_value(self) -> None:
        with patch.object(secrets_store, "get_secret", return_value="stored-key"):
            value = secrets_store.resolve_unifi_api_key("cli-key")
        self.assertEqual(value, "cli-key")

    def test_uses_stored_key_when_cli_and_env_missing(self) -> None:
        with patch.object(secrets_store, "get_secret", return_value="stored-key"):
            value = secrets_store.resolve_unifi_api_key(None)
        self.assertEqual(value, "stored-key")

    def test_raises_when_key_missing_everywhere(self) -> None:
        with patch.object(secrets_store, "get_secret", return_value=None):
            with self.assertRaises(NPMError) as exc_info:
                secrets_store.resolve_unifi_api_key(None)
        self.assertIn("Missing UniFi API key", str(exc_info.exception))


if __name__ == "__main__":
    unittest.main()
