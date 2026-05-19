package npmctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	serviceName       = "npmctl"
	legacyServiceName = "npm-cli"

	cfTokenKey     = "cloudflare_api_token"
	unifiAPIKeyKey = "unifi_api_key"
	loginInfoKey   = "login_info"
)

type SecretStore interface {
	Get(secretName string) (string, bool, error)
	Set(secretName, value string) error
	Delete(secretName string) error
}

type OSSecretStore struct{}

func NewOSSecretStore() OSSecretStore {
	return OSSecretStore{}
}

func (s OSSecretStore) Get(secretName string) (string, bool, error) {
	value, ok, err := osKeyringGet(serviceName, secretName)
	if err != nil {
		return "", false, fmt.Errorf("Failed to read secret from keyring: %w", err)
	}
	if ok && value != "" {
		return value, true, nil
	}

	value, ok, err = osKeyringGet(legacyServiceName, secretName)
	if err != nil {
		return "", false, fmt.Errorf("Failed to read secret from keyring: %w", err)
	}
	return value, ok && value != "", nil
}

func (s OSSecretStore) Set(secretName, value string) error {
	if err := osKeyringSet(serviceName, secretName, value); err != nil {
		return fmt.Errorf("Failed to store secret in keyring: %w", err)
	}
	return nil
}

func (s OSSecretStore) Delete(secretName string) error {
	var firstErr error
	for _, service := range []string{serviceName, legacyServiceName} {
		if err := osKeyringDelete(service, secretName); err != nil && !isNotFoundError(err) && firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return fmt.Errorf("Failed to delete secret from keyring: %w", firstErr)
	}
	return nil
}

func osKeyringGet(service, account string) (string, bool, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w").Output()
		if err != nil {
			if isNotFoundError(err) {
				return "", false, nil
			}
			return "", false, err
		}
		return strings.TrimRight(string(out), "\r\n"), true, nil
	case "linux":
		out, err := exec.Command("secret-tool", "lookup", "service", service, "username", account).Output()
		if err != nil && isNotFoundError(err) {
			out, err = exec.Command("secret-tool", "lookup", "service", service, "account", account).Output()
		}
		if err != nil {
			if isNotFoundError(err) {
				return "", false, nil
			}
			return "", false, err
		}
		return strings.TrimRight(string(out), "\r\n"), true, nil
	default:
		return "", false, npmError("OS keyring is only implemented for macOS Keychain and Linux Secret Service")
	}
}

func osKeyringSet(service, account, value string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("security", "add-generic-password", "-U", "-s", service, "-a", account, "-w", value).Run()
	case "linux":
		cmd := exec.Command("secret-tool", "store", "--label", service+" "+account, "service", service, "username", account)
		cmd.Stdin = strings.NewReader(value)
		return cmd.Run()
	default:
		return npmError("OS keyring is only implemented for macOS Keychain and Linux Secret Service")
	}
}

func osKeyringDelete(service, account string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("security", "delete-generic-password", "-s", service, "-a", account).Run()
	case "linux":
		if err := exec.Command("secret-tool", "clear", "service", service, "username", account).Run(); err != nil && !isNotFoundError(err) {
			return err
		}
		return exec.Command("secret-tool", "clear", "service", service, "account", account).Run()
	default:
		return npmError("OS keyring is only implemented for macOS Keychain and Linux Secret Service")
	}
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not found") || strings.Contains(message, "could not be found")
}

type loginInfo struct {
	BaseURL  string `json:"base_url"`
	Token    string `json:"token"`
	Expires  string `json:"expires"`
	Identity string `json:"identity"`
}

func getLoginInfo(store SecretStore) (loginInfo, error) {
	raw, ok, err := store.Get(loginInfoKey)
	if err != nil || !ok {
		return loginInfo{}, err
	}

	var info loginInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return loginInfo{}, nil
	}
	return info, nil
}

func setLoginInfo(store SecretStore, baseURL, token, expires, identity string) error {
	raw, err := json.Marshal(loginInfo{
		BaseURL:  baseURL,
		Token:    token,
		Expires:  expires,
		Identity: identity,
	})
	if err != nil {
		return err
	}
	return store.Set(loginInfoKey, string(raw))
}

func resolveCloudflareToken(store SecretStore, cliOrEnvToken string) (string, error) {
	if cliOrEnvToken != "" {
		return cliOrEnvToken, nil
	}
	if token, ok, err := store.Get(cfTokenKey); err != nil {
		return "", err
	} else if ok {
		return token, nil
	}
	return "", npmError("Missing Cloudflare API token. Use --cloudflare-api-token, set CLOUDFLARE_API_TOKEN, or run `npmctl secret set cloudflare-token`.")
}

func resolveUnifiAPIKey(store SecretStore, cliOrEnvAPIKey string) (string, error) {
	if cliOrEnvAPIKey != "" {
		return cliOrEnvAPIKey, nil
	}
	if apiKey, ok, err := store.Get(unifiAPIKeyKey); err != nil {
		return "", err
	} else if ok {
		return apiKey, nil
	}
	return "", npmError("Missing UniFi API key. Use --unifi-api-key, set UNIFI_API_KEY, or run `npmctl secret set unifi-api-key`.")
}
