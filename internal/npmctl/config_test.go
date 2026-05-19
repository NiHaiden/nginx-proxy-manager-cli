package npmctl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBaseURLAndTokenPrefersExplicitCLIArguments(t *testing.T) {
	t.Setenv("NPM_BASE_URL", "")
	t.Setenv("NPM_TOKEN", "")
	store := newMemorySecretStore()
	if err := setLoginInfo(store, "https://kr", "kr", "", "admin@example.com"); err != nil {
		t.Fatal(err)
	}

	baseURL, token, err := resolveBaseURLAndToken(store, "https://cli.example/api", "cli-token", true, debugLogger{})
	if err != nil {
		t.Fatal(err)
	}
	if baseURL != "https://cli.example/api" || token != "cli-token" {
		t.Fatalf("baseURL/token = %q/%q", baseURL, token)
	}
}

func TestResolveBaseURLAndTokenPrefersEnvironmentOverStoredValues(t *testing.T) {
	t.Setenv("NPM_BASE_URL", "https://env.example/api")
	t.Setenv("NPM_TOKEN", "env-token")
	store := newMemorySecretStore()
	if err := setLoginInfo(store, "https://kr", "kr", "", "admin@example.com"); err != nil {
		t.Fatal(err)
	}

	baseURL, token, err := resolveBaseURLAndToken(store, "", "", true, debugLogger{})
	if err != nil {
		t.Fatal(err)
	}
	if baseURL != "https://env.example/api" || token != "env-token" {
		t.Fatalf("baseURL/token = %q/%q", baseURL, token)
	}
}

func TestResolveBaseURLAndTokenRaisesWhenMissing(t *testing.T) {
	t.Setenv("NPM_BASE_URL", "")
	t.Setenv("NPM_TOKEN", "")
	store := newMemorySecretStore()

	_, _, err := resolveBaseURLAndToken(store, "", "token", true, debugLogger{})
	if err == nil || err.Error() == "" {
		t.Fatal("expected missing base URL error")
	}
}

func TestLoadJSONFile(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "config.json")
	if err := os.WriteFile(valid, []byte(`{"base_url":"https://x","token":"t"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	values := loadJSONFile(valid)
	if values["base_url"] != "https://x" || values["token"] != "t" {
		t.Fatalf("values = %#v", values)
	}

	invalid := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalid, []byte(`{invalid`), 0o600); err != nil {
		t.Fatal(err)
	}
	if values := loadJSONFile(invalid); len(values) != 0 {
		t.Fatalf("invalid json returned %#v", values)
	}
}
