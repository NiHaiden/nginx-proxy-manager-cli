package npmctl

import (
	"fmt"
	"strings"
)

func (c *CLI) printProxyHost(host map[string]any) {
	fmt.Fprintln(c.out, "Proxy host created successfully.")
	fmt.Fprintf(c.out, "Host ID: %v\n", host["id"])
	fmt.Fprintf(c.out, "Domains: %s\n", strings.Join(stringList(host["domain_names"]), ", "))
}

func (c *CLI) printCombinedSuccess(certID int, host map[string]any) {
	fmt.Fprintln(c.out, "Certificate and proxy host created successfully.")
	fmt.Fprintf(c.out, "Certificate ID: %d\n", certID)
	fmt.Fprintf(c.out, "Proxy Host ID: %v\n", host["id"])
	fmt.Fprintf(c.out, "Domains: %s\n", strings.Join(stringList(host["domain_names"]), ", "))
}

func (c *CLI) printCertificate(cert map[string]any) {
	fmt.Fprintln(c.out, "Certificate created successfully.")
	fmt.Fprintf(c.out, "Certificate ID: %v\n", cert["id"])
	fmt.Fprintf(c.out, "Name: %v\n", cert["nice_name"])
	fmt.Fprintf(c.out, "Domains: %s\n", strings.Join(stringList(cert["domain_names"]), ", "))
}

func (c *CLI) printSites(payload any) {
	sites := extractSites(payload)
	if len(sites) == 0 {
		fmt.Fprintln(c.out, "No UniFi sites returned.")
		return
	}
	fmt.Fprintln(c.out, "UniFi sites:")
	for _, site := range sites {
		if id := siteID(site); id != "" {
			fmt.Fprintf(c.out, "- %s: %s\n", siteLabel(site), id)
		} else {
			fmt.Fprintf(c.out, "- %s\n", siteLabel(site))
		}
	}
}

func (c *CLI) printDNSRecord(payload any, domain, selectedSiteID string) {
	fmt.Fprintln(c.out, "UniFi DNS record created successfully.")
	fmt.Fprintf(c.out, "Domain: %s\n", domain)
	fmt.Fprintf(c.out, "Site ID: %s\n", selectedSiteID)
	if values, ok := payload.(map[string]any); ok {
		if recordID := firstNonEmpty(stringFromAny(values["id"]), stringFromAny(values["_id"])); recordID != "" {
			fmt.Fprintf(c.out, "Record ID: %s\n", recordID)
		}
	}
}

func (c *CLI) printAddNewAppSuccess(domain, appIP string, appPort int, siteID string, dnsPayload any, proxyPayload map[string]any) {
	fmt.Fprintln(c.out, "New app DNS + proxy host created successfully.")
	fmt.Fprintf(c.out, "Domain: %s\n", domain)
	fmt.Fprintf(c.out, "App target: %s:%d\n", appIP, appPort)
	fmt.Fprintf(c.out, "UniFi Site ID: %s\n", siteID)
	if values, ok := dnsPayload.(map[string]any); ok {
		if recordID := firstNonEmpty(stringFromAny(values["id"]), stringFromAny(values["_id"])); recordID != "" {
			fmt.Fprintf(c.out, "UniFi DNS Record ID: %s\n", recordID)
		}
	}
	if proxyHostID := stringFromAny(proxyPayload["id"]); proxyHostID != "" {
		fmt.Fprintf(c.out, "Proxy Host ID: %s\n", proxyHostID)
	}
}

func stringList(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			values = append(values, fmt.Sprint(item))
		}
		return values
	default:
		return nil
	}
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case jsonNumber:
		parsed, _ := typed.Int64()
		return int(parsed)
	default:
		parsed, _ := strconvAtoi(fmt.Sprint(value))
		return parsed
	}
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

type jsonNumber interface {
	Int64() (int64, error)
}

func strconvAtoi(value string) (int, error) {
	if value == "" || value == "<nil>" {
		return 0, fmt.Errorf("empty")
	}
	var result int
	_, err := fmt.Sscan(value, &result)
	return result, err
}
