from __future__ import annotations

import unittest

from npmctl.commands.auth import _extract_token_data
from npmctl.errors import NPMError


class ExtractTokenDataTests(unittest.TestCase):
    def test_extracts_token_and_expiry(self) -> None:
        token, expires = _extract_token_data({"token": "abc123", "expires": "2026-12-31T00:00:00Z"})
        self.assertEqual(token, "abc123")
        self.assertEqual(expires, "2026-12-31T00:00:00Z")

    def test_raises_when_2fa_required(self) -> None:
        with self.assertRaises(NPMError) as exc_info:
            _extract_token_data({"requires_2fa": True})
        self.assertIn("2FA required", str(exc_info.exception))

    def test_raises_when_token_missing(self) -> None:
        with self.assertRaises(NPMError) as exc_info:
            _extract_token_data({"expires": "2026-12-31T00:00:00Z"})
        self.assertIn("did not include a token", str(exc_info.exception))


if __name__ == "__main__":
    unittest.main()
