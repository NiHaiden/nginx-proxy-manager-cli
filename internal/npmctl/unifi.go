package npmctl

import (
	"fmt"
	"strings"
)

func (c *CLI) resolveSiteID(client *NPMClient, unifiAPIKey, explicitSiteID string) (string, error) {
	if explicitSiteID != "" {
		return explicitSiteID, nil
	}
	payload, err := client.ListUnifiSites(unifiAPIKey)
	if err != nil {
		return "", err
	}
	sites := extractSites(payload)
	selectedSiteID, selectedSite, err := resolveSiteIDFromSites(sites)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(c.out, "Auto-selected site ID: %s (%s)\n", selectedSiteID, siteLabel(selectedSite))
	return selectedSiteID, nil
}

func resolveSiteIDFromSites(sites []map[string]any) (string, map[string]any, error) {
	if len(sites) == 0 {
		return "", nil, npmError("No UniFi sites returned. Pass --site-id explicitly or verify your API key and gateway URL.")
	}
	if len(sites) > 1 {
		lines := make([]string, 0, len(sites))
		for _, site := range sites {
			lines = append(lines, formatSiteChoice(site))
		}
		return "", nil, npmError("Multiple UniFi sites found. Pass --site-id explicitly.\nAvailable sites:\n" + strings.Join(lines, "\n"))
	}
	site := sites[0]
	siteID := siteID(site)
	if siteID == "" {
		return "", nil, npmError("UniFi returned one site but without a usable site ID. Pass --site-id explicitly.")
	}
	return siteID, site, nil
}

func extractSites(payload any) []map[string]any {
	switch typed := payload.(type) {
	case []any:
		return mapsFromList(typed)
	case []map[string]any:
		return typed
	case map[string]any:
		for _, key := range []string{"sites", "data", "items", "results"} {
			if list, ok := typed[key].([]any); ok {
				return mapsFromList(list)
			}
		}
	}
	return nil
}

func mapsFromList(items []any) []map[string]any {
	sites := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if site, ok := item.(map[string]any); ok {
			sites = append(sites, site)
		}
	}
	return sites
}

func formatSiteChoice(site map[string]any) string {
	id := siteID(site)
	if id == "" {
		id = "(missing id)"
	}
	return fmt.Sprintf("- %s: %s", siteLabel(site), id)
}

func siteLabel(site map[string]any) string {
	for _, key := range []string{"name", "desc", "description"} {
		if value, ok := site[key].(string); ok && value != "" {
			return value
		}
	}
	return "(unnamed)"
}

func siteID(site map[string]any) string {
	for _, key := range []string{"id", "_id", "siteId"} {
		if value, ok := site[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}
