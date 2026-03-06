from __future__ import annotations

import unittest

from npmctl.commands.unifi_dns import _resolve_site_id_from_sites
from npmctl.errors import NPMError


class ResolveSiteIdFromSitesTests(unittest.TestCase):
    def test_selects_single_site(self) -> None:
        site_id, site = _resolve_site_id_from_sites(
            [{"id": "site-123", "name": "default"}]
        )
        self.assertEqual(site_id, "site-123")
        self.assertEqual(site["name"], "default")

    def test_raises_for_zero_sites(self) -> None:
        with self.assertRaises(NPMError) as exc_info:
            _resolve_site_id_from_sites([])

        self.assertIn("No UniFi sites returned", str(exc_info.exception))

    def test_raises_for_multiple_sites(self) -> None:
        with self.assertRaises(NPMError) as exc_info:
            _resolve_site_id_from_sites(
                [
                    {"id": "site-1", "name": "Home"},
                    {"id": "site-2", "name": "Lab"},
                ]
            )

        message = str(exc_info.exception)
        self.assertIn("Multiple UniFi sites found", message)
        self.assertIn("site-1", message)
        self.assertIn("site-2", message)

    def test_raises_when_single_site_has_no_id(self) -> None:
        with self.assertRaises(NPMError) as exc_info:
            _resolve_site_id_from_sites([{"name": "Home"}])

        self.assertIn("without a usable site ID", str(exc_info.exception))


if __name__ == "__main__":
    unittest.main()
