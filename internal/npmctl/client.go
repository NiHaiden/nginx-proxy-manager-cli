package npmctl

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type NPMClient struct {
	BaseURL        string
	Token          string
	TimeoutSeconds int
	VerifyTLS      bool
	Logger         debugLogger
	HTTPClient     *http.Client
}

func NewNPMClient(baseURL, token string, verifyTLS bool, logger debugLogger) *NPMClient {
	return &NPMClient{
		BaseURL:        baseURL,
		Token:          token,
		TimeoutSeconds: 180,
		VerifyTLS:      verifyTLS,
		Logger:         logger,
	}
}

func (c *NPMClient) RequestToken(identity, secret, scope string) (map[string]any, error) {
	payload := map[string]any{"identity": identity, "secret": secret, "scope": scope}
	result, err := c.Request("POST", "/tokens", payload, false, nil)
	if err != nil {
		return nil, err
	}
	return asMap(result), nil
}

func (c *NPMClient) CreateCloudflareCertificate(domainNames []string, cloudflareAPIToken, niceName string, propagationSeconds int, keyType string) (map[string]any, error) {
	if niceName == "" && len(domainNames) > 0 {
		niceName = domainNames[0]
	}
	payload := map[string]any{
		"provider":     "letsencrypt",
		"nice_name":    niceName,
		"domain_names": domainNames,
		"meta": map[string]any{
			"dns_challenge":            true,
			"dns_provider":             "cloudflare",
			"dns_provider_credentials": "dns_cloudflare_api_token = " + cloudflareAPIToken,
			"propagation_seconds":      propagationSeconds,
			"key_type":                 keyType,
		},
	}
	result, err := c.Request("POST", "/nginx/certificates", payload, true, nil)
	if err != nil {
		return nil, err
	}
	return asMap(result), nil
}

type ProxyHostInput struct {
	DomainNames           []string
	ForwardHost           string
	ForwardPort           int
	ForwardScheme         string
	CertificateID         int
	HasCertificateID      bool
	SSLForced             bool
	HTTP2Support          bool
	HSTSEnabled           bool
	HSTSSubdomains        bool
	BlockExploits         bool
	AllowWebsocketUpgrade bool
	CachingEnabled        bool
	Enabled               bool
}

func (c *NPMClient) CreateProxyHost(input ProxyHostInput) (map[string]any, error) {
	payload := map[string]any{
		"domain_names":            input.DomainNames,
		"forward_scheme":          defaultString(input.ForwardScheme, "http"),
		"forward_host":            input.ForwardHost,
		"forward_port":            input.ForwardPort,
		"ssl_forced":              input.SSLForced,
		"http2_support":           input.HTTP2Support,
		"hsts_enabled":            input.HSTSEnabled,
		"hsts_subdomains":         input.HSTSSubdomains,
		"block_exploits":          input.BlockExploits,
		"allow_websocket_upgrade": input.AllowWebsocketUpgrade,
		"caching_enabled":         input.CachingEnabled,
		"enabled":                 input.Enabled,
		"access_list_id":          0,
		"advanced_config":         "",
		"locations":               []any{},
	}
	if input.HasCertificateID {
		payload["certificate_id"] = input.CertificateID
	}
	result, err := c.Request("POST", "/nginx/proxy-hosts", payload, true, nil)
	if err != nil {
		return nil, err
	}
	return asMap(result), nil
}

func (c *NPMClient) ListUnifiSites(unifiAPIKey string) (any, error) {
	headers := map[string]string{
		"User-Agent": "npmctl",
		"Accept":     "application/json",
		"X-API-KEY":  unifiAPIKey,
	}
	return c.Request("GET", "/proxy/network/integration/v1/sites", nil, false, headers)
}

func (c *NPMClient) CreateUnifiDNSRecord(unifiAPIKey, siteID, domain, ipv4Address string, ttlSeconds int, enabled bool, recordType string) (any, error) {
	payload := map[string]any{
		"type":        recordType,
		"enabled":     enabled,
		"domain":      domain,
		"ipv4Address": ipv4Address,
		"ttlSeconds":  ttlSeconds,
	}
	headers := map[string]string{
		"User-Agent": "npmctl",
		"Accept":     "application/json",
		"X-API-KEY":  unifiAPIKey,
	}
	return c.Request("POST", "/proxy/network/integration/v1/sites/"+siteID+"/dns/policies", payload, false, headers)
}

func (c *NPMClient) Headers(authRequired bool, extraHeaders map[string]string) (map[string]string, error) {
	headers := map[string]string{"Content-Type": "application/json"}
	if c.Token != "" {
		headers["Authorization"] = "Bearer " + c.Token
	} else if authRequired {
		return nil, npmError("No API token configured. Run `npmctl auth login` first.")
	}
	for key, value := range extraHeaders {
		headers[key] = value
	}
	return headers, nil
}

func (c *NPMClient) Request(method, path string, jsonBody map[string]any, authRequired bool, extraHeaders map[string]string) (any, error) {
	headers, err := c.Headers(authRequired, extraHeaders)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(c.BaseURL, "/") + path
	c.Logger.Log("Request: %s %s", method, url)
	if jsonBody != nil {
		c.Logger.Log("Request payload:\n%s", preview(redactValue(jsonBody, "json_body"), 1600))
	}

	response, payload, err := c.SendRequest(method, url, headers, jsonBody)
	if err != nil {
		return nil, err
	}
	c.logResponse(response, payload)

	if response.StatusCode >= 400 {
		return nil, c.apiError(method, path, response, payload, jsonBody)
	}
	return payload, nil
}

func (c *NPMClient) SendRequest(method, url string, headers map[string]string, jsonBody map[string]any) (*http.Response, any, error) {
	var body io.Reader
	if jsonBody != nil {
		raw, err := json.Marshal(jsonBody)
		if err != nil {
			return nil, nil, err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := c.HTTPClient
	if client == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		if !c.VerifyTLS {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
		}
		client = &http.Client{
			Timeout:   time.Duration(c.TimeoutSeconds) * time.Second,
			Transport: transport,
		}
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	payload := parsePayload(raw)
	return response, payload, nil
}

func parsePayload(raw []byte) any {
	if len(raw) == 0 {
		return ""
	}
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return string(raw)
	}
	return payload
}

func (c *NPMClient) logResponse(response *http.Response, payload any) {
	requestID := response.Header.Get("x-request-id")
	c.Logger.Log("Response status: %d", response.StatusCode)
	if requestID != "" {
		c.Logger.Log("Response request id: %s", requestID)
	}
	c.Logger.Log("Response payload:\n%s", preview(redactValue(payload, "payload"), 1600))
}

func (c *NPMClient) apiError(method, path string, response *http.Response, payload any, jsonBody map[string]any) error {
	lines := []string{fmt.Sprintf("API %d %s %s: %s", response.StatusCode, method, path, errorMessage(payload))}
	if requestID := response.Header.Get("x-request-id"); requestID != "" {
		lines = append(lines, "Request ID: "+requestID)
	}
	if c.Logger.enabled {
		lines = append(lines, "Response payload:", preview(redactValue(payload, "payload"), 1600))
		if jsonBody != nil {
			lines = append(lines, "Request payload:", preview(redactValue(jsonBody, "json_body"), 1600))
		}
	} else {
		lines = append(lines, "Tip: rerun with --debug (or set NPM_CLI_DEBUG=1) for full logs.")
	}
	return npmError(strings.Join(lines, "\n"))
}

func errorMessage(payload any) string {
	if values, ok := payload.(map[string]any); ok {
		if errorObj, ok := values["error"].(map[string]any); ok {
			if message, ok := errorObj["message"].(string); ok && message != "" {
				return message
			}
			return fmt.Sprint(errorObj)
		}
		if message, ok := values["message"].(string); ok && message != "" {
			return message
		}
		return fmt.Sprint(values)
	}
	return fmt.Sprint(payload)
}

func asMap(value any) map[string]any {
	if values, ok := value.(map[string]any); ok {
		return values
	}
	return map[string]any{}
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
