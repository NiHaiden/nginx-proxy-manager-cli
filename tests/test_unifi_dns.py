from __future__ import annotations

import unittest
from unittest.mock import Mock, patch

from npmctl.commands.unifi_dns import (
    _build_unifi_client,
    _create_dns_and_proxy,
    _resolve_site_id_from_sites,
)
from npmctl.errors import NPMError


class ResolveSiteIdFromSitesTests(unittest.TestCase):
    def test_build_unifi_client_sets_verify_false_when_insecure(self) -> None:
        with patch("npmctl.commands.unifi_dns.typer.secho") as secho_mock:
            client = _build_unifi_client("https://192.168.1.1", insecure=True)

        self.assertFalse(client.verify_tls)
        secho_mock.assert_called_once()

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


class CreateDnsAndProxyTests(unittest.TestCase):
    def test_create_dns_and_proxy_uses_shared_values(self) -> None:
        unifi_client = Mock()
        npm_client = Mock()
        unifi_client.create_unifi_dns_record.return_value = {"id": "dns-123"}
        npm_client.create_proxy_host.return_value = {"id": 99, "domain_names": ["app.local"]}

        with patch("npmctl.commands.unifi_dns._resolve_site_id", return_value="site-42") as resolve_mock:
            selected_site_id, dns_payload, proxy_payload = _create_dns_and_proxy(
                unifi_client=unifi_client,
                npm_client=npm_client,
                unifi_api_key="unifi-secret",
                explicit_site_id=None,
                domain="app.local",
                app_ip="192.168.1.50",
                app_port=3000,
                ttl_seconds=3600,
                enabled=True,
                record_type="A_RECORD",
                forward_scheme="http",
                certificate_id=None,
                ssl_forced=False,
                http2_support=False,
                hsts_enabled=False,
                hsts_subdomains=False,
            )

        self.assertEqual(selected_site_id, "site-42")
        self.assertEqual(dns_payload, {"id": "dns-123"})
        self.assertEqual(proxy_payload, {"id": 99, "domain_names": ["app.local"]})

        resolve_mock.assert_called_once_with(unifi_client, "unifi-secret", None)
        unifi_client.create_unifi_dns_record.assert_called_once_with(
            unifi_api_key="unifi-secret",
            site_id="site-42",
            domain="app.local",
            ipv4_address="192.168.1.50",
            ttl_seconds=3600,
            enabled=True,
            record_type="A_RECORD",
        )
        npm_client.create_proxy_host.assert_called_once_with(
            domain_names=["app.local"],
            forward_host="192.168.1.50",
            forward_port=3000,
            forward_scheme="http",
            certificate_id=None,
            ssl_forced=False,
            http2_support=False,
            hsts_enabled=False,
            hsts_subdomains=False,
        )


if __name__ == "__main__":
    unittest.main()
