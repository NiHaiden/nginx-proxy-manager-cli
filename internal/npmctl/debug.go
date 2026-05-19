package npmctl

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var sensitiveKeys = []string{
	"authorization",
	"token",
	"secret",
	"password",
	"dns_provider_credentials",
	"cloudflare_api_token",
	"x-api-key",
	"api_key",
	"unifi_api_key",
}

type debugLogger struct {
	enabled bool
	err     io.Writer
}

func (l debugLogger) Log(format string, args ...any) {
	if !l.enabled {
		return
	}
	fmt.Fprintf(l.err, "[debug] "+format+"\n", args...)
}

func maskSecret(value string) string {
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func isSensitiveKey(key string) bool {
	hint := strings.ToLower(key)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(hint, sensitive) {
			return true
		}
	}
	return false
}

func redactValue(value any, keyHint string) any {
	if isSensitiveKey(keyHint) {
		if text, ok := value.(string); ok {
			return maskSecret(text)
		}
		return "***"
	}

	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			redacted[key] = redactValue(item, key)
		}
		return redacted
	case []any:
		redacted := make([]any, len(typed))
		for i, item := range typed {
			redacted[i] = redactValue(item, keyHint)
		}
		return redacted
	default:
		return value
	}
}

func preview(value any, limit int) string {
	text := toText(value)
	if len(text) <= limit {
		return text
	}
	return fmt.Sprintf("%s... [truncated %d chars]", text[:limit], len(text)-limit)
}

func toText(value any) string {
	switch value.(type) {
	case map[string]any, []any:
		raw, err := json.MarshalIndent(value, "", "  ")
		if err == nil {
			return string(raw)
		}
	}
	return fmt.Sprint(value)
}
