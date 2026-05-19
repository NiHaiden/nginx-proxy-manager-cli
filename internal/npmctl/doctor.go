package npmctl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func (c *CLI) runDoctor() error {
	issues := []string{}

	fmt.Fprintln(c.out, "npmctl doctor")
	fmt.Fprintf(c.out, "Platform: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(c.out, "Go runtime: %s\n", runtime.Version())
	if distro := linuxDistroLabel(); distro != "" {
		fmt.Fprintf(c.out, "Linux distro: %s\n", distro)
	}

	if err := checkKeyringTool(); err != nil {
		issues = append(issues, err.Error())
	} else {
		fmt.Fprintln(c.out, "keyring backend command available")
	}

	fmt.Fprintln(c.out, "\nStored secrets:")
	for _, check := range []struct {
		label      string
		exists     func() (bool, error)
		fixCommand string
	}{
		{"NPM login token", c.hasLoginToken, "npmctl auth login"},
		{"Cloudflare token", c.hasStoredSecret(cfTokenKey), "npmctl secret set cloudflare-token"},
		{"UniFi API key", c.hasStoredSecret(unifiAPIKeyKey), "npmctl secret set unifi-api-key"},
	} {
		exists, err := check.exists()
		if err != nil {
			fmt.Fprintf(c.out, "%s: unable to read keyring\n", check.label)
			fmt.Fprintf(c.out, "  %v\n", err)
			issues = append(issues, "Unable to read "+check.label+" from keyring.")
			continue
		}
		if exists {
			fmt.Fprintf(c.out, "%s: stored\n", check.label)
		} else {
			fmt.Fprintf(c.out, "%s: not stored\n", check.label)
			fmt.Fprintf(c.out, "  Add it with `%s`.\n", check.fixCommand)
		}
	}

	if len(issues) == 0 {
		fmt.Fprintln(c.out, "\nNo blocking issues found.")
		return nil
	}

	fmt.Fprintln(c.out, "\nIssues found:")
	for _, issue := range issues {
		fmt.Fprintf(c.out, "  - %s\n", issue)
	}
	c.printDoctorFixHints()
	return npmError("doctor found blocking issues")
}

func (c *CLI) hasLoginToken() (bool, error) {
	info, err := getLoginInfo(c.store)
	return info.Token != "", err
}

func (c *CLI) hasStoredSecret(key string) func() (bool, error) {
	return func() (bool, error) {
		_, ok, err := c.store.Get(key)
		return ok, err
	}
}

func checkKeyringTool() error {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("security"); err != nil {
			return npmError("macOS `security` command is not available. Secure token storage may fail.")
		}
	case "linux":
		if _, err := exec.LookPath("secret-tool"); err != nil {
			return npmError("Linux `secret-tool` command is not available. Secure token storage may fail.")
		}
	default:
		return npmError("No supported OS keyring backend detected. macOS Keychain and Linux Secret Service are supported.")
	}
	return nil
}

func (c *CLI) printDoctorFixHints() {
	fmt.Fprintln(c.out, "\nHow to fix keyring backend issues:")
	switch runtime.GOOS {
	case "darwin":
		fmt.Fprintln(c.out, "macOS: make sure the `security` command can access your login keychain from this user session.")
	case "linux":
		distroInfo := loadLinuxDistroInfo()
		distroID := strings.ToLower(distroInfo["ID"])
		distroLike := strings.ToLower(distroInfo["ID_LIKE"])
		switch {
		case hasAnyToken(distroID, distroLike, "fedora", "rhel", "centos", "rocky", "almalinux"):
			fmt.Fprintln(c.out, "Fedora/RHEL family:")
			fmt.Fprintln(c.out, "  sudo dnf install -y libsecret")
		case hasAnyToken(distroID, distroLike, "arch", "manjaro", "endeavouros"):
			fmt.Fprintln(c.out, "Arch family:")
			fmt.Fprintln(c.out, "  sudo pacman -S --needed libsecret")
		case hasAnyToken(distroID, distroLike, "ubuntu", "debian", "linuxmint", "pop"):
			fmt.Fprintln(c.out, "Ubuntu/Debian family:")
			fmt.Fprintln(c.out, "  sudo apt update && sudo apt install -y libsecret-tools dbus-user-session")
		default:
			fmt.Fprintln(c.out, "Install and run a Secret Service compatible keyring backend.")
		}
		fmt.Fprintln(c.out, "Then run `npmctl auth status` again from your normal user session.")
	}
}

func linuxDistroLabel() string {
	if runtime.GOOS != "linux" {
		return ""
	}
	values := loadLinuxDistroInfo()
	if values["PRETTY_NAME"] != "" {
		return values["PRETTY_NAME"]
	}
	if values["ID"] != "" && values["VERSION_ID"] != "" {
		return values["ID"] + " " + values["VERSION_ID"]
	}
	return values["ID"]
}

func loadLinuxDistroInfo() map[string]string {
	for _, path := range []string{"/etc/os-release", "/usr/lib/os-release"} {
		raw, err := os.ReadFile(path)
		if err == nil {
			return parseOSRelease(string(raw))
		}
	}
	return map[string]string{}
}

func parseOSRelease(raw string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		key, value, _ := strings.Cut(line, "=")
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		values[key] = value
	}
	return values
}

func hasAnyToken(first, rest string, tokens ...string) bool {
	all := map[string]bool{first: true}
	for _, token := range strings.Fields(rest) {
		all[token] = true
	}
	for _, token := range tokens {
		if all[token] {
			return true
		}
	}
	return false
}
