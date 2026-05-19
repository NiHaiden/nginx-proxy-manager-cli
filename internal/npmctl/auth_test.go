package npmctl

import (
	"strings"
	"testing"
)

func TestExtractTokenData(t *testing.T) {
	token, expires, err := extractTokenData(map[string]any{"token": "abc123", "expires": "2026-12-31T00:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	if token != "abc123" || expires != "2026-12-31T00:00:00Z" {
		t.Fatalf("token/expires = %q/%q", token, expires)
	}
}

func TestExtractTokenDataRaisesWhen2FARequired(t *testing.T) {
	_, _, err := extractTokenData(map[string]any{"requires_2fa": true})
	if err == nil || !strings.Contains(err.Error(), "2FA required") {
		t.Fatalf("expected 2FA error, got %v", err)
	}
}

func TestExtractTokenDataRaisesWhenTokenMissing(t *testing.T) {
	_, _, err := extractTokenData(map[string]any{"expires": "2026-12-31T00:00:00Z"})
	if err == nil || !strings.Contains(err.Error(), "did not include a token") {
		t.Fatalf("expected token missing error, got %v", err)
	}
}
