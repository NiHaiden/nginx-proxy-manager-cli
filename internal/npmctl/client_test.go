package npmctl

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHeadersIncludeBearerToken(t *testing.T) {
	client := NewNPMClient("https://npm.example/api", "jwt-token", true, debugLogger{})
	headers, err := client.Headers(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if headers["Authorization"] != "Bearer jwt-token" {
		t.Fatalf("Authorization = %q", headers["Authorization"])
	}
	if headers["Content-Type"] != "application/json" {
		t.Fatalf("Content-Type = %q", headers["Content-Type"])
	}
}

func TestHeadersRaiseWhenAuthRequiredWithoutToken(t *testing.T) {
	client := NewNPMClient("https://npm.example/api", "", true, debugLogger{})
	_, err := client.Headers(true, nil)
	if err == nil || !strings.Contains(err.Error(), "No API token configured") {
		t.Fatalf("expected auth error, got %v", err)
	}
}

func TestCloudflareCertificatePayloadContainsDNSCredentials(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nginx/certificates" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewNPMClient(server.URL, "jwt-token", true, debugLogger{})
	if _, err := client.CreateCloudflareCertificate([]string{"example.com", "*.example.com"}, "cf-secret", "Example", 45, "ecdsa"); err != nil {
		t.Fatal(err)
	}

	meta := payload["meta"].(map[string]any)
	if payload["provider"] != "letsencrypt" {
		t.Fatalf("provider = %v", payload["provider"])
	}
	if meta["dns_provider"] != "cloudflare" {
		t.Fatalf("dns_provider = %v", meta["dns_provider"])
	}
	if meta["dns_provider_credentials"] != "dns_cloudflare_api_token = cf-secret" {
		t.Fatalf("credentials = %v", meta["dns_provider_credentials"])
	}
	if meta["propagation_seconds"].(float64) != 45 {
		t.Fatalf("propagation_seconds = %v", meta["propagation_seconds"])
	}
	if meta["key_type"] != "ecdsa" {
		t.Fatalf("key_type = %v", meta["key_type"])
	}
}

func TestListUnifiSitesUsesExpectedEndpointAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proxy/network/integration/v1/sites" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-API-KEY") != "unifi-secret" {
			t.Fatalf("X-API-KEY = %q", r.Header.Get("X-API-KEY"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("Accept = %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("User-Agent") != "npmctl" {
			t.Fatalf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte(`[{"id":"site-1"}]`))
	}))
	defer server.Close()

	client := NewNPMClient(server.URL, "", true, debugLogger{})
	if _, err := client.ListUnifiSites("unifi-secret"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateUnifiDNSRecordPayloadMatchesAPIContract(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proxy/network/integration/v1/sites/site-123/dns/policies" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"id":"dns-1"}`))
	}))
	defer server.Close()

	client := NewNPMClient(server.URL, "", true, debugLogger{})
	if _, err := client.CreateUnifiDNSRecord("unifi-secret", "site-123", "test.nhaiden.io", "192.168.1.246", 14400, true, "A_RECORD"); err != nil {
		t.Fatal(err)
	}

	expected := map[string]any{
		"type":        "A_RECORD",
		"enabled":     true,
		"domain":      "test.nhaiden.io",
		"ipv4Address": "192.168.1.246",
		"ttlSeconds":  float64(14400),
	}
	for key, want := range expected {
		if got := payload[key]; got != want {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
	}
}

func TestAPIErrorIncludesDebugTipWhenDebugOff(t *testing.T) {
	client := NewNPMClient("https://npm.example/api", "jwt-token", true, debugLogger{})
	header := http.Header{}
	header.Set("x-request-id", "req-123")
	response := &http.Response{StatusCode: 400, Header: header}

	err := client.apiError("POST", "/nginx/proxy-hosts", response, map[string]any{"message": "Bad request"}, map[string]any{"token": "super-secret"})
	if err == nil {
		t.Fatal("expected error")
	}
	message := err.Error()
	for _, expected := range []string{"API 400 POST /nginx/proxy-hosts: Bad request", "Request ID: req-123", "Tip: rerun with --debug"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("error missing %q:\n%s", expected, message)
		}
	}
}

func TestAPIErrorRedactsSensitiveRequestDataWhenDebugOn(t *testing.T) {
	client := NewNPMClient("https://npm.example/api", "jwt-token", true, debugLogger{enabled: true})
	response := &http.Response{StatusCode: 401, Header: http.Header{}}

	err := client.apiError("POST", "/tokens", response, map[string]any{"message": "Unauthorized"}, map[string]any{"secret": "plaintext-password", "identity": "admin@example.com"})
	if err == nil {
		t.Fatal("expected error")
	}
	message := err.Error()
	if !strings.Contains(message, "Request payload:") || !strings.Contains(message, "plai...word") {
		t.Fatalf("redacted payload missing:\n%s", message)
	}
	if strings.Contains(message, "plaintext-password") {
		t.Fatalf("sensitive value leaked:\n%s", message)
	}
}
