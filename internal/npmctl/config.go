package npmctl

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func resolveBaseURLAndToken(store SecretStore, baseURL, token string, requireToken bool, logger debugLogger) (string, string, error) {
	keyringLogin, err := getLoginInfo(store)
	if err != nil {
		logger.Log("Keyring login lookup skipped: %v", err)
		keyringLogin = loginInfo{}
	}
	legacyLogin := loadLegacyLoginInfo()

	resolvedBaseURL := firstNonEmpty(
		baseURL,
		os.Getenv("NPM_BASE_URL"),
		keyringLogin.BaseURL,
		stringFromMap(legacyLogin, "base_url"),
	)
	if resolvedBaseURL == "" {
		return "", "", npmError("Missing NPM base URL. Use --base-url or set NPM_BASE_URL.")
	}

	resolvedToken := firstNonEmpty(
		token,
		os.Getenv("NPM_TOKEN"),
		keyringLogin.Token,
		stringFromMap(legacyLogin, "token"),
	)
	if requireToken && resolvedToken == "" {
		return "", "", npmError("Missing API token. Run `npmctl auth login` or set NPM_TOKEN.")
	}

	logger.Log("Using base URL: %s", resolvedBaseURL)
	logger.Log("Using token: %s", yesNo(resolvedToken != ""))
	return resolvedBaseURL, resolvedToken, nil
}

func loadLegacyLoginInfo() map[string]any {
	for _, path := range []string{legacyNPMCTLConfigPath(), expandHome("~/.config/npm-cli/config.json")} {
		if loaded := loadJSONFile(path); len(loaded) > 0 {
			return loaded
		}
	}
	return map[string]any{}
}

func legacyNPMCTLConfigPath() string {
	if path := os.Getenv("NPM_CLI_CONFIG"); path != "" {
		return expandHome(path)
	}
	return expandHome("~/.config/npmctl/config.json")
}

func loadJSONFile(path string) map[string]any {
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}

	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return map[string]any{}
	}
	if parsed == nil {
		return map[string]any{}
	}
	return parsed
}

func expandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if len(path) > 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func stringFromMap(values map[string]any, key string) string {
	if value, ok := values[key].(string); ok {
		return value
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
