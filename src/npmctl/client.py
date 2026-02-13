from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Optional

import httpx

from .debug import is_debug, log, preview, redact_value
from .errors import NPMError


@dataclass
class NPMClient:
    base_url: str
    token: Optional[str] = None
    timeout_seconds: int = 180

    def request_token(self, identity: str, secret: str, scope: str = "user") -> dict[str, Any]:
        body = {"identity": identity, "secret": secret, "scope": scope}
        return self._request("POST", "/tokens", json_body=body, auth_required=False)

    def create_cloudflare_certificate(
        self,
        *,
        domain_names: list[str],
        cloudflare_api_token: str,
        nice_name: Optional[str] = None,
        propagation_seconds: int = 30,
        key_type: str = "rsa",
    ) -> dict[str, Any]:
        credentials = f"dns_cloudflare_api_token = {cloudflare_api_token}"
        payload = {
            "provider": "letsencrypt",
            "nice_name": nice_name or domain_names[0],
            "domain_names": domain_names,
            "meta": {
                "dns_challenge": True,
                "dns_provider": "cloudflare",
                "dns_provider_credentials": credentials,
                "propagation_seconds": propagation_seconds,
                "key_type": key_type,
            },
        }
        return self._request("POST", "/nginx/certificates", json_body=payload)

    def create_proxy_host(
        self,
        *,
        domain_names: list[str],
        forward_host: str,
        forward_port: int,
        forward_scheme: str = "http",
        certificate_id: Optional[int] = None,
        ssl_forced: bool = False,
        http2_support: bool = False,
        hsts_enabled: bool = False,
        hsts_subdomains: bool = False,
        block_exploits: bool = True,
        allow_websocket_upgrade: bool = True,
        caching_enabled: bool = False,
        enabled: bool = True,
    ) -> dict[str, Any]:
        payload = {
            "domain_names": domain_names,
            "forward_scheme": forward_scheme,
            "forward_host": forward_host,
            "forward_port": forward_port,
            "ssl_forced": ssl_forced,
            "http2_support": http2_support,
            "hsts_enabled": hsts_enabled,
            "hsts_subdomains": hsts_subdomains,
            "block_exploits": block_exploits,
            "allow_websocket_upgrade": allow_websocket_upgrade,
            "caching_enabled": caching_enabled,
            "enabled": enabled,
            "access_list_id": 0,
            "advanced_config": "",
            "locations": [],
        }
        if certificate_id is not None:
            payload["certificate_id"] = certificate_id

        return self._request("POST", "/nginx/proxy-hosts", json_body=payload)

    def _request(
        self,
        method: str,
        path: str,
        *,
        json_body: Optional[dict[str, Any]] = None,
        auth_required: bool = True,
    ) -> Any:
        headers = self._headers(auth_required)
        url = f"{self.base_url.rstrip('/')}{path}"

        self._log_request(method, url, json_body)
        response = self._send_request(method, url, headers, json_body)
        payload = self._parse_payload(response)
        self._log_response(response, payload)

        if response.status_code >= 400:
            self._raise_api_error(method, path, response, payload, json_body)

        return payload

    def _headers(self, auth_required: bool) -> dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        elif auth_required:
            raise NPMError("No API token configured. Run `npmctl login` first.")
        return headers

    def _send_request(
        self,
        method: str,
        url: str,
        headers: dict[str, str],
        json_body: Optional[dict[str, Any]],
    ) -> httpx.Response:
        try:
            return httpx.request(
                method,
                url,
                headers=headers,
                json=json_body,
                timeout=self.timeout_seconds,
            )
        except httpx.HTTPError as exc:
            raise NPMError(f"HTTP request failed: {exc}") from exc

    @staticmethod
    def _parse_payload(response: httpx.Response) -> Any:
        try:
            return response.json()
        except ValueError:
            return response.text

    @staticmethod
    def _error_message(payload: Any) -> str:
        if isinstance(payload, dict):
            error_obj = payload.get("error")
            if isinstance(error_obj, dict):
                return error_obj.get("message") or str(error_obj)
            return payload.get("message") or str(payload)
        return str(payload)

    def _log_request(self, method: str, url: str, json_body: Optional[dict[str, Any]]) -> None:
        log(f"Request: {method} {url}")
        if json_body is not None:
            body = preview(redact_value(json_body, "json_body"))
            log(f"Request payload:\n{body}")

    def _log_response(self, response: httpx.Response, payload: Any) -> None:
        request_id = response.headers.get("x-request-id")
        log(f"Response status: {response.status_code}")
        if request_id:
            log(f"Response request id: {request_id}")
        body = preview(redact_value(payload, "payload"))
        log(f"Response payload:\n{body}")

    def _raise_api_error(
        self,
        method: str,
        path: str,
        response: httpx.Response,
        payload: Any,
        json_body: Optional[dict[str, Any]],
    ) -> None:
        message = self._error_message(payload)
        request_id = response.headers.get("x-request-id")

        lines = [f"API {response.status_code} {method} {path}: {message}"]
        if request_id:
            lines.append(f"Request ID: {request_id}")

        if is_debug():
            lines.extend(["Response payload:", preview(redact_value(payload, "payload"))])
            if json_body is not None:
                lines.extend(["Request payload:", preview(redact_value(json_body, "json_body"))])
        else:
            lines.append("Tip: rerun with --debug (or set NPM_CLI_DEBUG=1) for full logs.")

        raise NPMError("\n".join(lines))
