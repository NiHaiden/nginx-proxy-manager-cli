package npmctl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type CLI struct {
	in     io.Reader
	out    io.Writer
	err    io.Writer
	store  SecretStore
	debug  bool
	logger debugLogger
}

func NewCLI(in io.Reader, out io.Writer, err io.Writer, store SecretStore) *CLI {
	return &CLI{in: in, out: out, err: err, store: store}
}

func (c *CLI) Run(args []string) int {
	args = c.configureGlobalOptions(args)
	c.logger = debugLogger{enabled: c.debug, err: c.err}

	if len(args) == 0 || isHelp(args[0]) {
		c.printRootHelp()
		return 0
	}

	if err := c.dispatch(args); err != nil {
		fmt.Fprintln(c.err, err)
		return 1
	}
	return 0
}

func (c *CLI) configureGlobalOptions(args []string) []string {
	c.debug = truthy(os.Getenv("NPM_CLI_DEBUG"))
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--debug" {
			c.debug = true
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered
}

func (c *CLI) dispatch(args []string) error {
	switch args[0] {
	case "auth":
		return c.dispatchAuth(args[1:])
	case "secret":
		return c.dispatchSecret(args[1:])
	case "site":
		return c.dispatchSite(args[1:])
	case "dns":
		return c.dispatchDNS(args[1:])
	case "proxy":
		return c.dispatchProxy(args[1:])
	case "cert":
		return c.dispatchCert(args[1:])
	case "app":
		return c.dispatchApp(args[1:])
	case "add":
		return c.dispatchAdd(args[1:])
	case "list":
		return c.dispatchList(args[1:])
	case "status":
		return c.dispatchStatus(args[1:])
	case "delete":
		return c.dispatchDelete(args[1:])
	case "doctor":
		if len(args) > 1 && isHelp(args[1]) {
			c.printCommandHelp("doctor")
			return nil
		}
		return c.runDoctor()
	case "login":
		return c.runAuthLogin(args[1:])
	case "login-status":
		return c.runAuthStatus()
	case "logout":
		return c.runAuthLogout()
	case "cf-token-set":
		return c.runSecretSet("cloudflare-token", args[1:])
	case "cf-token-status":
		return c.runSecretStatus("cloudflare-token")
	case "cf-token-delete":
		return c.runSecretDelete("cloudflare-token")
	case "unifi-api-key-set":
		return c.runSecretSet("unifi-api-key", args[1:])
	case "unifi-api-key-status":
		return c.runSecretStatus("unifi-api-key")
	case "unifi-api-key-delete":
		return c.runSecretDelete("unifi-api-key")
	case "list-unifi-sites":
		return c.runListUnifiSites(args[1:])
	case "add-unifi-dns-record":
		return c.runAddUnifiDNSRecord(args[1:])
	case "add-new-app":
		return c.runAddNewApp(args[1:])
	case "add-cert-cloudflare":
		return c.runAddCertCloudflare(args[1:])
	case "add-proxy-host":
		return c.runAddProxyHost(args[1:], false)
	case "add-proxy-with-cert":
		return c.runAddProxyHost(args[1:], true)
	default:
		return fmt.Errorf("Unknown command: %s\nRun `npmctl --help` for usage.", args[0])
	}
}

func (c *CLI) dispatchAuth(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("auth")
		return nil
	}
	switch args[0] {
	case "login":
		return c.runAuthLogin(args[1:])
	case "status":
		return c.runAuthStatus()
	case "logout":
		return c.runAuthLogout()
	default:
		return unknownSubcommand("auth", args[0])
	}
}

func (c *CLI) dispatchSecret(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("secret")
		return nil
	}
	switch args[0] {
	case "set":
		secretName, rest, err := consumeSecretName(args[1:])
		if err != nil {
			return err
		}
		return c.runSecretSet(secretName, rest)
	case "status":
		secretName, _, err := consumeSecretName(args[1:])
		if err != nil {
			return err
		}
		return c.runSecretStatus(secretName)
	case "delete":
		secretName, _, err := consumeSecretName(args[1:])
		if err != nil {
			return err
		}
		return c.runSecretDelete(secretName)
	default:
		return unknownSubcommand("secret", args[0])
	}
}

func (c *CLI) dispatchSite(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("site")
		return nil
	}
	if args[0] != "list" {
		return unknownSubcommand("site", args[0])
	}
	return c.runListUnifiSites(args[1:])
}

func (c *CLI) dispatchDNS(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("dns")
		return nil
	}
	if args[0] != "add" {
		return unknownSubcommand("dns", args[0])
	}
	return c.runAddUnifiDNSRecord(args[1:])
}

func (c *CLI) dispatchProxy(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("proxy")
		return nil
	}
	if args[0] != "add" {
		return unknownSubcommand("proxy", args[0])
	}
	return c.runProxyAdd(args[1:])
}

func (c *CLI) dispatchCert(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("cert")
		return nil
	}
	if args[0] != "add" {
		return unknownSubcommand("cert", args[0])
	}
	return c.runAddCertCloudflare(args[1:])
}

func (c *CLI) dispatchApp(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("app")
		return nil
	}
	if args[0] != "add" {
		return unknownSubcommand("app", args[0])
	}
	return c.runAddNewApp(args[1:])
}

func (c *CLI) dispatchAdd(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("add")
		return nil
	}
	switch args[0] {
	case "proxy":
		return c.runProxyAdd(args[1:])
	case "cert":
		return c.runAddCertCloudflare(args[1:])
	case "dns":
		return c.runAddUnifiDNSRecord(args[1:])
	case "app":
		return c.runAddNewApp(args[1:])
	default:
		return unknownSubcommand("add", args[0])
	}
}

func (c *CLI) dispatchList(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("list")
		return nil
	}
	if args[0] != "sites" {
		return unknownSubcommand("list", args[0])
	}
	return c.runListUnifiSites(args[1:])
}

func (c *CLI) dispatchStatus(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("status")
		return nil
	}
	switch args[0] {
	case "login":
		return c.runAuthStatus()
	case "cloudflare-token":
		return c.runSecretStatus("cloudflare-token")
	case "unifi-api-key":
		return c.runSecretStatus("unifi-api-key")
	default:
		return unknownSubcommand("status", args[0])
	}
}

func (c *CLI) dispatchDelete(args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		c.printGroupHelp("delete")
		return nil
	}
	switch args[0] {
	case "cloudflare-token":
		return c.runSecretDelete("cloudflare-token")
	case "unifi-api-key":
		return c.runSecretDelete("unifi-api-key")
	default:
		return unknownSubcommand("delete", args[0])
	}
}

func (c *CLI) runAuthLogin(args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("auth login")
		return nil
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}

	identity := options.String("identity", "")
	baseURL := options.String("base-url", "")
	scope := options.String("scope", "user")
	secret := options.String("secret", "")
	if identity == "" {
		return npmError("Missing option --identity.")
	}
	if baseURL == "" {
		return npmError("Missing option --base-url.")
	}
	if secret == "" {
		secret = c.promptHidden("Secret: ")
	}

	client, err := c.buildClient(baseURL, "", false)
	if err != nil {
		return err
	}
	data, err := client.RequestToken(identity, secret, scope)
	if err != nil {
		return err
	}
	token, expires, err := extractTokenData(data)
	if err != nil {
		return err
	}
	if err := setLoginInfo(c.store, client.BaseURL, token, expires, identity); err != nil {
		return err
	}
	fmt.Fprintln(c.out, "Login successful. Token saved in OS keyring.")
	if expires != "" {
		fmt.Fprintf(c.out, "Token expires: %s\n", expires)
	}
	return nil
}

func (c *CLI) runAuthStatus() error {
	info, err := getLoginInfo(c.store)
	if err != nil {
		return err
	}
	if info.Token == "" && info.BaseURL == "" && info.Identity == "" {
		fmt.Fprintln(c.out, "No stored login found in OS keyring.")
		return nil
	}
	fmt.Fprintln(c.out, "Stored login found in OS keyring.")
	fmt.Fprintf(c.out, "Identity: %s\n", fallback(info.Identity, "(unknown)"))
	fmt.Fprintf(c.out, "Base URL: %s\n", fallback(info.BaseURL, "(unknown)"))
	fmt.Fprintf(c.out, "Token expires: %s\n", fallback(info.Expires, "(unknown)"))
	return nil
}

func (c *CLI) runAuthLogout() error {
	if err := c.store.Delete(loginInfoKey); err != nil {
		return err
	}
	fmt.Fprintln(c.out, "Stored login removed from OS keyring.")
	return nil
}

func (c *CLI) runSecretSet(secretName string, args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("secret set")
		return nil
	}
	key, err := secretKey(secretName)
	if err != nil {
		return err
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	value := firstNonEmpty(options.String("value", ""), options.String("token", ""), options.String("api-key", ""))
	if value == "" {
		value = c.promptHidden("Secret value: ")
	}
	if err := c.store.Set(key, value); err != nil {
		return err
	}
	fmt.Fprintln(c.out, secretSavedMessage(secretName))
	return nil
}

func (c *CLI) runSecretStatus(secretName string) error {
	key, err := secretKey(secretName)
	if err != nil {
		return err
	}
	_, ok, err := c.store.Get(key)
	if err != nil {
		return err
	}
	if ok {
		fmt.Fprintln(c.out, secretPresentMessage(secretName))
	} else {
		fmt.Fprintln(c.out, secretMissingMessage(secretName))
	}
	return nil
}

func (c *CLI) runSecretDelete(secretName string) error {
	key, err := secretKey(secretName)
	if err != nil {
		return err
	}
	if err := c.store.Delete(key); err != nil {
		return err
	}
	fmt.Fprintln(c.out, secretDeletedMessage(secretName))
	return nil
}

func (c *CLI) runProxyAdd(args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("proxy add")
		return nil
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	if options.Bool("create-cert") && options.Has("certificate-id") {
		return npmError("Choose either --certificate-id or --create-cert, not both.")
	}

	client, err := c.buildClient(options.String("base-url", ""), options.String("token", ""), true)
	if err != nil {
		return err
	}
	input, err := proxyInputFromOptions(options)
	if err != nil {
		return err
	}

	if options.Bool("create-cert") {
		cfToken, err := resolveCloudflareToken(c.store, firstNonEmpty(options.String("cloudflare-api-token", ""), os.Getenv("CLOUDFLARE_API_TOKEN")))
		if err != nil {
			return err
		}
		certID, err := c.createCloudflareCert(client, input.DomainNames, cfToken, options.Int("propagation-seconds", 30), options.String("key-type", "rsa"), options.String("certificate-name", ""))
		if err != nil {
			return err
		}
		input.CertificateID = certID
		input.HasCertificateID = true
		host, err := c.createProxyHostWithCert(client, input)
		if err != nil {
			return err
		}
		c.printCombinedSuccess(certID, host)
		return nil
	}

	fmt.Fprintln(c.out, "Creating proxy host...")
	host, err := client.CreateProxyHost(input)
	if err != nil {
		return err
	}
	c.printProxyHost(host)
	return nil
}

func (c *CLI) runAddProxyHost(args []string, withCert bool) error {
	if !withCert {
		return c.runProxyAdd(args)
	}
	return c.runAddProxyWithCert(args)
}

func (c *CLI) runAddProxyWithCert(args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("proxy add")
		return nil
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	client, err := c.buildClient(options.String("base-url", ""), options.String("token", ""), true)
	if err != nil {
		return err
	}
	input, err := proxyInputFromOptions(options)
	if err != nil {
		return err
	}
	input.SSLForced = true
	input.HTTP2Support = true
	input.HSTSEnabled = true
	input.HSTSSubdomains = true

	cfToken, err := resolveCloudflareToken(c.store, firstNonEmpty(options.String("cloudflare-api-token", ""), os.Getenv("CLOUDFLARE_API_TOKEN")))
	if err != nil {
		return err
	}
	certID, err := c.createCloudflareCert(client, input.DomainNames, cfToken, options.Int("propagation-seconds", 30), options.String("key-type", "rsa"), "")
	if err != nil {
		return err
	}
	input.CertificateID = certID
	input.HasCertificateID = true
	host, err := c.createProxyHostWithCert(client, input)
	if err != nil {
		return err
	}
	c.printCombinedSuccess(certID, host)
	return nil
}

func (c *CLI) runAddCertCloudflare(args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("cert add")
		return nil
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	client, err := c.buildClient(options.String("base-url", ""), options.String("token", ""), true)
	if err != nil {
		return err
	}
	domains := options.Values("domain")
	if len(domains) == 0 {
		return npmError("Missing option --domain.")
	}
	cfToken, err := resolveCloudflareToken(c.store, firstNonEmpty(options.String("cloudflare-api-token", ""), os.Getenv("CLOUDFLARE_API_TOKEN")))
	if err != nil {
		return err
	}
	fmt.Fprintln(c.out, "Creating certificate via Cloudflare DNS challenge...")
	cert, err := client.CreateCloudflareCertificate(domains, cfToken, options.String("nice-name", ""), options.Int("propagation-seconds", 30), options.String("key-type", "rsa"))
	if err != nil {
		return err
	}
	c.printCertificate(cert)
	return nil
}

func (c *CLI) runListUnifiSites(args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("site list")
		return nil
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	client, err := c.buildUnifiClient(options.String("gateway-url", ""), options.Bool("insecure"))
	if err != nil {
		return err
	}
	apiKey, err := resolveUnifiAPIKey(c.store, firstNonEmpty(options.String("unifi-api-key", ""), os.Getenv("UNIFI_API_KEY")))
	if err != nil {
		return err
	}
	payload, err := client.ListUnifiSites(apiKey)
	if err != nil {
		return err
	}
	c.printSites(payload)
	return nil
}

func (c *CLI) runAddUnifiDNSRecord(args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("dns add")
		return nil
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	client, err := c.buildUnifiClient(options.String("gateway-url", ""), options.Bool("insecure"))
	if err != nil {
		return err
	}
	apiKey, err := resolveUnifiAPIKey(c.store, firstNonEmpty(options.String("unifi-api-key", ""), os.Getenv("UNIFI_API_KEY")))
	if err != nil {
		return err
	}
	domain := options.String("domain", "")
	ipv4 := options.String("ipv4-address", "")
	if domain == "" {
		return npmError("Missing option --domain.")
	}
	if ipv4 == "" {
		return npmError("Missing option --ipv4-address.")
	}
	siteID, err := c.resolveSiteID(client, apiKey, options.String("site-id", ""))
	if err != nil {
		return err
	}
	fmt.Fprintln(c.out, "Creating UniFi DNS record...")
	payload, err := client.CreateUnifiDNSRecord(apiKey, siteID, domain, ipv4, options.Int("ttl-seconds", 14400), options.Enabled(), options.String("record-type", "A_RECORD"))
	if err != nil {
		return err
	}
	c.printDNSRecord(payload, domain, siteID)
	return nil
}

func (c *CLI) runAddNewApp(args []string) error {
	if hasHelp(args) {
		c.printCommandHelp("app add")
		return nil
	}
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	domain := options.String("domain", "")
	appIP := options.String("app-ip", "")
	appPort := options.Int("app-port", 0)
	if domain == "" {
		return npmError("Missing option --domain.")
	}
	if appIP == "" {
		return npmError("Missing option --app-ip.")
	}
	if appPort < 1 || appPort > 65535 {
		return npmError("Missing or invalid option --app-port.")
	}

	unifiClient, err := c.buildUnifiClient(options.String("gateway-url", ""), options.Bool("insecure"))
	if err != nil {
		return err
	}
	npmClient, err := c.buildClient(options.String("base-url", ""), options.String("token", ""), true)
	if err != nil {
		return err
	}
	apiKey, err := resolveUnifiAPIKey(c.store, firstNonEmpty(options.String("unifi-api-key", ""), os.Getenv("UNIFI_API_KEY")))
	if err != nil {
		return err
	}
	siteID, dnsPayload, proxyPayload, err := c.createDNSAndProxy(unifiClient, npmClient, apiKey, options, domain, appIP, appPort)
	if err != nil {
		return err
	}
	c.printAddNewAppSuccess(domain, appIP, appPort, siteID, dnsPayload, proxyPayload)
	return nil
}

func (c *CLI) buildClient(baseURL, token string, requireToken bool) (*NPMClient, error) {
	resolvedBaseURL, resolvedToken, err := resolveBaseURLAndToken(c.store, baseURL, token, requireToken, c.logger)
	if err != nil {
		return nil, err
	}
	return NewNPMClient(resolvedBaseURL, resolvedToken, true, c.logger), nil
}

func (c *CLI) buildUnifiClient(gatewayURL string, insecure bool) (*NPMClient, error) {
	gatewayURL = firstNonEmpty(gatewayURL, os.Getenv("UNIFI_BASE_URL"))
	if gatewayURL == "" {
		return nil, npmError("Missing option --gateway-url.")
	}
	if insecure {
		fmt.Fprintln(c.err, "Warning: TLS certificate verification is disabled for this request.")
	}
	return NewNPMClient(gatewayURL, "", !insecure, c.logger), nil
}

func (c *CLI) createCloudflareCert(client *NPMClient, domains []string, cfToken string, propagationSeconds int, keyType, niceName string) (int, error) {
	fmt.Fprintln(c.out, "Step 1/2: Creating certificate via Cloudflare DNS challenge...")
	cert, err := client.CreateCloudflareCertificate(domains, cfToken, niceName, propagationSeconds, keyType)
	if err != nil {
		return 0, err
	}
	certID := intFromAny(cert["id"])
	if certID == 0 {
		return 0, npmError("Certificate creation succeeded but no certificate id was returned.")
	}
	fmt.Fprintf(c.out, "Certificate created (id=%d).\n", certID)
	return certID, nil
}

func (c *CLI) createProxyHostWithCert(client *NPMClient, input ProxyHostInput) (map[string]any, error) {
	fmt.Fprintln(c.out, "Step 2/2: Creating proxy host with that certificate...")
	return client.CreateProxyHost(input)
}

func (c *CLI) createDNSAndProxy(unifiClient, npmClient *NPMClient, apiKey string, options parsedOptions, domain, appIP string, appPort int) (string, any, map[string]any, error) {
	siteID, err := c.resolveSiteID(unifiClient, apiKey, options.String("site-id", ""))
	if err != nil {
		return "", nil, nil, err
	}
	fmt.Fprintln(c.out, "Step 1/2: Creating UniFi DNS record...")
	dnsPayload, err := unifiClient.CreateUnifiDNSRecord(apiKey, siteID, domain, appIP, options.Int("ttl-seconds", 14400), options.Enabled(), options.String("record-type", "A_RECORD"))
	if err != nil {
		return "", nil, nil, err
	}
	fmt.Fprintln(c.out, "Step 2/2: Creating NPM proxy host...")
	proxyPayload, err := npmClient.CreateProxyHost(ProxyHostInput{
		DomainNames:           []string{domain},
		ForwardHost:           appIP,
		ForwardPort:           appPort,
		ForwardScheme:         options.String("forward-scheme", "http"),
		CertificateID:         options.Int("certificate-id", 0),
		HasCertificateID:      options.Has("certificate-id"),
		SSLForced:             options.Bool("ssl-forced"),
		HTTP2Support:          options.Bool("http2-support"),
		HSTSEnabled:           options.Bool("hsts-enabled"),
		HSTSSubdomains:        options.Bool("hsts-subdomains"),
		BlockExploits:         true,
		AllowWebsocketUpgrade: true,
		Enabled:               true,
	})
	if err != nil {
		return "", nil, nil, err
	}
	return siteID, dnsPayload, proxyPayload, nil
}

func proxyInputFromOptions(options parsedOptions) (ProxyHostInput, error) {
	domains := options.Values("domain")
	if len(domains) == 0 {
		return ProxyHostInput{}, npmError("Missing option --domain.")
	}
	forwardHost := options.String("forward-host", "")
	if forwardHost == "" {
		return ProxyHostInput{}, npmError("Missing option --forward-host.")
	}
	forwardPort := options.Int("forward-port", 0)
	if forwardPort < 1 || forwardPort > 65535 {
		return ProxyHostInput{}, npmError("Missing or invalid option --forward-port.")
	}
	return ProxyHostInput{
		DomainNames:           domains,
		ForwardHost:           forwardHost,
		ForwardPort:           forwardPort,
		ForwardScheme:         options.String("forward-scheme", "http"),
		CertificateID:         options.Int("certificate-id", 0),
		HasCertificateID:      options.Has("certificate-id"),
		SSLForced:             options.Bool("ssl-forced"),
		HTTP2Support:          options.Bool("http2-support"),
		HSTSEnabled:           options.Bool("hsts-enabled"),
		HSTSSubdomains:        options.Bool("hsts-subdomains"),
		BlockExploits:         true,
		AllowWebsocketUpgrade: true,
		Enabled:               true,
	}, nil
}

func extractTokenData(data map[string]any) (string, string, error) {
	if required, ok := data["requires_2fa"].(bool); ok && required {
		return "", "", npmError("2FA required. Use /tokens/2fa flow manually for now.")
	}
	token, _ := data["token"].(string)
	expires, _ := data["expires"].(string)
	if token == "" {
		return "", "", npmError("Login response did not include a token.")
	}
	return token, expires, nil
}

func (c *CLI) prompt(label string) string {
	fmt.Fprint(c.err, label)
	scanner := bufio.NewScanner(c.in)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func (c *CLI) promptHidden(label string) string {
	file, ok := c.in.(*os.File)
	if !ok {
		return c.prompt(label)
	}

	_ = runStty(file, "-echo")
	defer func() {
		_ = runStty(file, "echo")
		fmt.Fprintln(c.err)
	}()
	return c.prompt(label)
}

func runStty(stdin *os.File, mode string) error {
	cmd := exec.Command("stty", mode)
	cmd.Stdin = stdin
	return cmd.Run()
}

func isHelp(arg string) bool {
	return arg == "--help" || arg == "-h"
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if isHelp(arg) {
			return true
		}
	}
	return false
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func consumeSecretName(args []string) (string, []string, error) {
	if len(args) == 0 || isHelp(args[0]) {
		return "", nil, npmError("Missing secret name. Expected cloudflare-token or unifi-api-key.")
	}
	if _, err := secretKey(args[0]); err != nil {
		return "", nil, err
	}
	return args[0], args[1:], nil
}

func secretKey(secretName string) (string, error) {
	switch secretName {
	case "cloudflare-token":
		return cfTokenKey, nil
	case "unifi-api-key":
		return unifiAPIKeyKey, nil
	default:
		return "", npmError("Unknown secret. Expected cloudflare-token or unifi-api-key.")
	}
}

func secretSavedMessage(secretName string) string {
	if secretName == "cloudflare-token" {
		return "Cloudflare token saved in OS keyring."
	}
	return "UniFi API key saved in OS keyring."
}

func secretDeletedMessage(secretName string) string {
	if secretName == "cloudflare-token" {
		return "Cloudflare token deleted from OS keyring."
	}
	return "UniFi API key deleted from OS keyring."
}

func secretPresentMessage(secretName string) string {
	if secretName == "cloudflare-token" {
		return "Cloudflare token is stored in keyring."
	}
	return "UniFi API key is stored in keyring."
}

func secretMissingMessage(secretName string) string {
	if secretName == "cloudflare-token" {
		return "No Cloudflare token stored in keyring."
	}
	return "No UniFi API key stored in keyring."
}

func unknownSubcommand(group, subcommand string) error {
	return fmt.Errorf("Unknown %s subcommand: %s", group, subcommand)
}

type parsedOptions struct {
	values map[string][]string
	flags  map[string]bool
}

func parseOptions(args []string) (parsedOptions, error) {
	options := parsedOptions{values: map[string][]string{}, flags: map[string]bool{}}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			continue
		}
		if arg == "-d" {
			if i+1 >= len(args) {
				return options, npmError("Option -d requires a value.")
			}
			options.values["domain"] = append(options.values["domain"], args[i+1])
			i++
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			return options, fmt.Errorf("Unexpected argument: %s", arg)
		}

		nameValue := strings.TrimPrefix(arg, "--")
		if strings.Contains(nameValue, "/--") {
			nameValue = strings.Split(nameValue, "/--")[0]
		}
		if nameValue == "disabled" {
			options.flags["disabled"] = true
			continue
		}
		if nameValue == "enabled" || isBooleanFlag(nameValue) {
			options.flags[nameValue] = true
			continue
		}

		name, value, ok := strings.Cut(nameValue, "=")
		if !ok {
			if i+1 >= len(args) {
				return options, fmt.Errorf("Option --%s requires a value.", name)
			}
			value = args[i+1]
			i++
		}
		options.values[name] = append(options.values[name], value)
	}
	return options, nil
}

func isBooleanFlag(name string) bool {
	switch name {
	case "create-cert", "ssl-forced", "http2-support", "hsts-enabled", "hsts-subdomains", "insecure":
		return true
	default:
		return false
	}
}

func (o parsedOptions) Has(name string) bool {
	if _, ok := o.values[name]; ok {
		return true
	}
	return o.flags[name]
}

func (o parsedOptions) Values(name string) []string {
	values := append([]string{}, o.values[name]...)
	return values
}

func (o parsedOptions) String(name, defaultValue string) string {
	values := o.values[name]
	if len(values) == 0 {
		return defaultValue
	}
	return values[len(values)-1]
}

func (o parsedOptions) Bool(name string) bool {
	return o.flags[name]
}

func (o parsedOptions) Enabled() bool {
	return !o.flags["disabled"]
}

func (o parsedOptions) Int(name string, defaultValue int) int {
	raw := o.String(name, "")
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return value
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
