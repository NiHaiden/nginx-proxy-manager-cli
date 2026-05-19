package npmctl

import (
	"strings"
	"testing"
)

func TestResolveSiteIDFromSitesSelectsSingleSite(t *testing.T) {
	siteID, site, err := resolveSiteIDFromSites([]map[string]any{{"id": "site-123", "name": "default"}})
	if err != nil {
		t.Fatal(err)
	}
	if siteID != "site-123" || site["name"] != "default" {
		t.Fatalf("siteID/site = %q/%#v", siteID, site)
	}
}

func TestResolveSiteIDFromSitesRaisesForZeroSites(t *testing.T) {
	_, _, err := resolveSiteIDFromSites(nil)
	if err == nil || !strings.Contains(err.Error(), "No UniFi sites returned") {
		t.Fatalf("expected zero sites error, got %v", err)
	}
}

func TestResolveSiteIDFromSitesRaisesForMultipleSites(t *testing.T) {
	_, _, err := resolveSiteIDFromSites([]map[string]any{
		{"id": "site-1", "name": "Home"},
		{"id": "site-2", "name": "Lab"},
	})
	if err == nil {
		t.Fatal("expected multiple sites error")
	}
	message := err.Error()
	for _, expected := range []string{"Multiple UniFi sites found", "site-1", "site-2"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("error missing %q:\n%s", expected, message)
		}
	}
}

func TestResolveSiteIDFromSitesRaisesWhenSingleSiteHasNoID(t *testing.T) {
	_, _, err := resolveSiteIDFromSites([]map[string]any{{"name": "Home"}})
	if err == nil || !strings.Contains(err.Error(), "without a usable site ID") {
		t.Fatalf("expected missing ID error, got %v", err)
	}
}
