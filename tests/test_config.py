from __future__ import annotations

import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from npmctl import config
from npmctl.errors import NPMError


class ResolveBaseUrlAndTokenTests(unittest.TestCase):
    def setUp(self) -> None:
        self.env_patch = patch.dict(os.environ, {}, clear=True)
        self.env_patch.start()
        self.addCleanup(self.env_patch.stop)

    def test_prefers_explicit_cli_arguments(self) -> None:
        with patch.object(config, "_load_keyring_login_info", return_value={"base_url": "https://kr", "token": "kr"}):
            with patch.object(
                config,
                "_load_legacy_login_info",
                return_value={"base_url": "https://legacy", "token": "legacy"},
            ):
                base_url, token = config.resolve_base_url_and_token(
                    "https://cli.example/api",
                    "cli-token",
                )

        self.assertEqual(base_url, "https://cli.example/api")
        self.assertEqual(token, "cli-token")

    def test_prefers_environment_over_stored_values(self) -> None:
        os.environ["NPM_BASE_URL"] = "https://env.example/api"
        os.environ["NPM_TOKEN"] = "env-token"

        with patch.object(config, "_load_keyring_login_info", return_value={"base_url": "https://kr", "token": "kr"}):
            with patch.object(
                config,
                "_load_legacy_login_info",
                return_value={"base_url": "https://legacy", "token": "legacy"},
            ):
                base_url, token = config.resolve_base_url_and_token(None, None)

        self.assertEqual(base_url, "https://env.example/api")
        self.assertEqual(token, "env-token")

    def test_raises_when_base_url_missing(self) -> None:
        with patch.object(config, "_load_keyring_login_info", return_value={}):
            with patch.object(config, "_load_legacy_login_info", return_value={}):
                with self.assertRaises(NPMError) as exc_info:
                    config.resolve_base_url_and_token(None, "token")

        self.assertIn("Missing NPM base URL", str(exc_info.exception))

    def test_raises_when_token_required_but_missing(self) -> None:
        with patch.object(config, "_load_keyring_login_info", return_value={"base_url": "https://stored.example/api"}):
            with patch.object(config, "_load_legacy_login_info", return_value={}):
                with self.assertRaises(NPMError) as exc_info:
                    config.resolve_base_url_and_token(None, None)

        self.assertIn("Missing API token", str(exc_info.exception))

    def test_allows_missing_token_when_not_required(self) -> None:
        with patch.object(config, "_load_keyring_login_info", return_value={"base_url": "https://stored.example/api"}):
            with patch.object(config, "_load_legacy_login_info", return_value={}):
                base_url, token = config.resolve_base_url_and_token(None, None, require_token=False)

        self.assertEqual(base_url, "https://stored.example/api")
        self.assertIsNone(token)


class LoadJsonFileTests(unittest.TestCase):
    def test_returns_empty_dict_when_file_missing(self) -> None:
        result = config._load_json_file(Path("/tmp/definitely-missing-npmctl-config.json"))
        self.assertEqual(result, {})

    def test_returns_empty_dict_for_invalid_json(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "config.json"
            path.write_text("{invalid", encoding="utf-8")
            result = config._load_json_file(path)
        self.assertEqual(result, {})

    def test_returns_empty_dict_for_non_object_json(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "config.json"
            path.write_text('["a", "b"]', encoding="utf-8")
            result = config._load_json_file(path)
        self.assertEqual(result, {})

    def test_returns_mapping_for_valid_json(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "config.json"
            path.write_text('{"base_url":"https://x","token":"t"}', encoding="utf-8")
            result = config._load_json_file(path)
        self.assertEqual(result, {"base_url": "https://x", "token": "t"})


if __name__ == "__main__":
    unittest.main()
