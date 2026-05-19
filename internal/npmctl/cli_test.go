package npmctl

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpShowsGroupedCommandsAndHidesLegacyFlatNames(t *testing.T) {
	var out, err bytes.Buffer
	cli := NewCLI(strings.NewReader(""), &out, &err, newMemorySecretStore())

	if code := cli.Run([]string{"--help"}); code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, err.String())
	}

	output := out.String()
	for _, expected := range []string{"auth", "secret", "proxy", "add"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("help missing %q:\n%s", expected, output)
		}
	}
	for _, hidden := range []string{"add-proxy-host", "login-status"} {
		if strings.Contains(output, hidden) {
			t.Fatalf("help should hide %q:\n%s", hidden, output)
		}
	}
}

func TestProxyAddHelpIncludesOptionalCertificateWorkflow(t *testing.T) {
	var out, err bytes.Buffer
	cli := NewCLI(strings.NewReader(""), &out, &err, newMemorySecretStore())

	if code := cli.Run([]string{"proxy", "add", "--help"}); code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, err.String())
	}

	output := out.String()
	for _, expected := range []string{"--create-cert", "Existing", "certificate ID"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("help missing %q:\n%s", expected, output)
		}
	}
}

func TestHiddenLegacyAliasStillExists(t *testing.T) {
	var out, err bytes.Buffer
	cli := NewCLI(strings.NewReader(""), &out, &err, newMemorySecretStore())

	if code := cli.Run([]string{"add-proxy-host", "--help"}); code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, err.String())
	}
	if !strings.Contains(out.String(), "--forward-host") {
		t.Fatalf("legacy help missing --forward-host:\n%s", out.String())
	}
}

func TestSecretSetStatusDeleteUsesStore(t *testing.T) {
	store := newMemorySecretStore()
	var out, err bytes.Buffer
	cli := NewCLI(strings.NewReader(""), &out, &err, store)

	if code := cli.Run([]string{"secret", "set", "cloudflare-token", "--value", "cf-secret"}); code != 0 {
		t.Fatalf("set exit code = %d, stderr = %s", code, err.String())
	}
	if got := store.values[cfTokenKey]; got != "cf-secret" {
		t.Fatalf("stored token = %q", got)
	}

	out.Reset()
	if code := cli.Run([]string{"secret", "status", "cloudflare-token"}); code != 0 {
		t.Fatalf("status exit code = %d, stderr = %s", code, err.String())
	}
	if !strings.Contains(out.String(), "Cloudflare token is stored") {
		t.Fatalf("unexpected status output: %s", out.String())
	}

	if code := cli.Run([]string{"secret", "delete", "cloudflare-token"}); code != 0 {
		t.Fatalf("delete exit code = %d, stderr = %s", code, err.String())
	}
	if _, ok := store.values[cfTokenKey]; ok {
		t.Fatal("token was not deleted")
	}
}
