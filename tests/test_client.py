from __future__ import annotations

import unittest
from unittest.mock import patch

import httpx

from npmctl.client import NPMClient
from npmctl.debug import set_debug
from npmctl.errors import NPMError


class ClientBehaviorTests(unittest.TestCase):
    def tearDown(self) -> None:
        set_debug(False)

    def test_headers_include_bearer_token(self) -> None:
        client = NPMClient(base_url="https://npm.example/api", token="jwt-token")
        headers = client._headers(auth_required=True)
        self.assertEqual(headers["Authorization"], "Bearer jwt-token")
        self.assertEqual(headers["Content-Type"], "application/json")

    def test_headers_raise_when_auth_required_without_token(self) -> None:
        client = NPMClient(base_url="https://npm.example/api", token=None)
        with self.assertRaises(NPMError) as exc_info:
            client._headers(auth_required=True)
        self.assertIn("No API token configured", str(exc_info.exception))

    def test_cloudflare_certificate_payload_contains_dns_credentials(self) -> None:
        client = NPMClient(base_url="https://npm.example/api", token="jwt-token")

        with patch.object(client, "_request", return_value={"id": 1}) as request_mock:
            client.create_cloudflare_certificate(
                domain_names=["example.com", "*.example.com"],
                cloudflare_api_token="cf-secret",
                nice_name="Example",
                propagation_seconds=45,
                key_type="ecdsa",
            )

        _, _, kwargs = request_mock.mock_calls[0]
        payload = kwargs["json_body"]
        meta = payload["meta"]
        self.assertEqual(payload["provider"], "letsencrypt")
        self.assertEqual(payload["domain_names"], ["example.com", "*.example.com"])
        self.assertEqual(meta["dns_provider"], "cloudflare")
        self.assertEqual(meta["dns_provider_credentials"], "dns_cloudflare_api_token = cf-secret")
        self.assertEqual(meta["propagation_seconds"], 45)
        self.assertEqual(meta["key_type"], "ecdsa")

    def test_headers_include_extra_headers(self) -> None:
        client = NPMClient(base_url="https://npm.example/api", token=None)
        headers = client._headers(auth_required=False, extra_headers={"X-API-KEY": "abc", "Accept": "application/json"})
        self.assertEqual(headers["X-API-KEY"], "abc")
        self.assertEqual(headers["Accept"], "application/json")

    def test_list_unifi_sites_uses_expected_endpoint_and_headers(self) -> None:
        client = NPMClient(base_url="https://192.168.1.1", token=None)

        with patch.object(client, "_request", return_value=[{"id": "site-1"}]) as request_mock:
            client.list_unifi_sites(unifi_api_key="unifi-secret")

        _, args, kwargs = request_mock.mock_calls[0]
        self.assertEqual(args[0], "GET")
        self.assertEqual(args[1], "/proxy/network/integration/v1/sites")
        self.assertFalse(kwargs["auth_required"])
        self.assertEqual(kwargs["extra_headers"]["X-API-KEY"], "unifi-secret")
        self.assertEqual(kwargs["extra_headers"]["Accept"], "application/json")
        self.assertEqual(kwargs["extra_headers"]["User-Agent"], "npmctl")

    def test_create_unifi_dns_record_payload_matches_api_contract(self) -> None:
        client = NPMClient(base_url="https://192.168.1.1", token=None)

        with patch.object(client, "_request", return_value={"id": "dns-1"}) as request_mock:
            client.create_unifi_dns_record(
                unifi_api_key="unifi-secret",
                site_id="site-123",
                domain="test.nhaiden.io",
                ipv4_address="192.168.1.246",
                ttl_seconds=14400,
                enabled=True,
                record_type="A_RECORD",
            )

        _, args, kwargs = request_mock.mock_calls[0]
        self.assertEqual(args[0], "POST")
        self.assertEqual(args[1], "/proxy/network/integration/v1/sites/site-123/dns/policies")
        self.assertFalse(kwargs["auth_required"])
        self.assertEqual(kwargs["extra_headers"]["X-API-KEY"], "unifi-secret")
        self.assertEqual(kwargs["extra_headers"]["User-Agent"], "npmctl")
        self.assertEqual(
            kwargs["json_body"],
            {
                "type": "A_RECORD",
                "enabled": True,
                "domain": "test.nhaiden.io",
                "ipv4Address": "192.168.1.246",
                "ttlSeconds": 14400,
            },
        )

    def test_send_request_omits_json_when_body_is_none(self) -> None:
        client = NPMClient(base_url="https://npm.example/api", token="jwt-token")

        with patch("npmctl.client.httpx.request") as request_mock:
            request_mock.return_value = httpx.Response(
                status_code=200,
                request=httpx.Request("GET", "https://npm.example/api/proxy/network/integration/v1/sites"),
            )
            client._send_request(
                "GET",
                "https://npm.example/api/proxy/network/integration/v1/sites",
                {"Accept": "application/json"},
                None,
            )

        _, kwargs = request_mock.call_args
        self.assertNotIn("json", kwargs)

    def test_send_request_respects_verify_tls_setting(self) -> None:
        client = NPMClient(base_url="https://192.168.1.1", token=None, verify_tls=False)

        with patch("npmctl.client.httpx.request") as request_mock:
            request_mock.return_value = httpx.Response(
                status_code=200,
                request=httpx.Request("GET", "https://192.168.1.1/proxy/network/integration/v1/sites"),
            )
            client._send_request(
                "GET",
                "https://192.168.1.1/proxy/network/integration/v1/sites",
                {"Accept": "application/json"},
                None,
            )

        _, kwargs = request_mock.call_args
        self.assertIn("verify", kwargs)
        self.assertFalse(kwargs["verify"])

    def test_raise_api_error_includes_debug_tip_when_debug_off(self) -> None:
        set_debug(False)
        client = NPMClient(base_url="https://npm.example/api", token="jwt-token")
        response = httpx.Response(
            status_code=400,
            headers={"x-request-id": "req-123"},
            request=httpx.Request("POST", "https://npm.example/api/nginx/proxy-hosts"),
        )

        with self.assertRaises(NPMError) as exc_info:
            client._raise_api_error(
                "POST",
                "/nginx/proxy-hosts",
                response,
                {"message": "Bad request"},
                {"token": "super-secret"},
            )

        message = str(exc_info.exception)
        self.assertIn("API 400 POST /nginx/proxy-hosts: Bad request", message)
        self.assertIn("Request ID: req-123", message)
        self.assertIn("Tip: rerun with --debug", message)

    def test_raise_api_error_redacts_sensitive_request_data_when_debug_on(self) -> None:
        set_debug(True)
        client = NPMClient(base_url="https://npm.example/api", token="jwt-token")
        response = httpx.Response(
            status_code=401,
            request=httpx.Request("POST", "https://npm.example/api/tokens"),
        )

        with self.assertRaises(NPMError) as exc_info:
            client._raise_api_error(
                "POST",
                "/tokens",
                response,
                {"message": "Unauthorized"},
                {"secret": "plaintext-password", "identity": "admin@example.com"},
            )

        message = str(exc_info.exception)
        self.assertIn("Request payload:", message)
        self.assertIn("plai...word", message)
        self.assertNotIn("plaintext-password", message)

    def test_error_message_uses_nested_error_object(self) -> None:
        client = NPMClient(base_url="https://npm.example/api", token="jwt-token")
        message = client._error_message({"error": {"message": "Nested error"}})
        self.assertEqual(message, "Nested error")


if __name__ == "__main__":
    unittest.main()
