package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/open-cli-collective/salesforce-cli/internal/config"
	"github.com/open-cli-collective/salesforce-cli/internal/keychain"
)

const (
	// CallbackURL is the OAuth redirect URL that must match the Connected App configuration.
	// No server listens on this — the browser shows an error and the user copies the URL.
	CallbackURL = "http://localhost:8080/callback"

	// ProductionLoginURL is the Salesforce production login endpoint
	ProductionLoginURL = "https://login.salesforce.com"

	// SandboxLoginURL is the Salesforce sandbox login endpoint
	SandboxLoginURL = "https://test.salesforce.com"
)

// Scopes contains the OAuth scopes required by the CLI
var Scopes = []string{
	"api",
	"refresh_token",
	"offline_access",
}

// GetOAuthConfig creates an OAuth2 config for Salesforce.
// instanceURL should be the Salesforce login URL (e.g., login.salesforce.com or custom domain).
func GetOAuthConfig(instanceURL, clientID string) *oauth2.Config {
	// Normalize instance URL
	instanceURL = normalizeInstanceURL(instanceURL)

	return &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  instanceURL + "/services/oauth2/authorize",
			TokenURL: instanceURL + "/services/oauth2/token",
		},
		RedirectURL: CallbackURL,
		Scopes:      Scopes,
	}
}

// GetHTTPClient returns an HTTP client with OAuth2 authentication.
// It retrieves tokens from keychain (preferred) or falls back to file storage.
// Returns an error if no token is found - caller should direct user to run 'sfdc init'.
func GetHTTPClient(ctx context.Context) (*http.Client, error) {
	// Load config to get instance URL and client ID
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.InstanceURL == "" || cfg.ClientID == "" {
		return nil, fmt.Errorf("not configured - please run 'sfdc init' first")
	}

	// Get OAuth config
	oauthConfig := GetOAuthConfig(cfg.InstanceURL, cfg.ClientID)

	// Try to load token from keychain
	tok, err := keychain.GetToken()
	if err != nil {
		return nil, fmt.Errorf("no OAuth token found - please run 'sfdc init' first: %w", err)
	}

	// Create persistent token source that saves refreshed tokens
	tokenSource := keychain.NewPersistentTokenSource(oauthConfig, tok)
	return oauth2.NewClient(ctx, tokenSource), nil
}

// GetAuthURL returns the OAuth authorization URL for the given config.
func GetAuthURL(config *oauth2.Config) string {
	return config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

// ExchangeAuthCode exchanges an authorization code for a token.
func ExchangeAuthCode(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(ctx, code)
}

// normalizeInstanceURL ensures the instance URL has proper format.
func normalizeInstanceURL(url string) string {
	url = strings.TrimSpace(url)

	// Add https:// if no scheme provided
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")

	return url
}

// IsProductionURL returns true if the URL is the production login URL.
func IsProductionURL(url string) bool {
	normalized := normalizeInstanceURL(url)
	return normalized == ProductionLoginURL
}

// IsSandboxURL returns true if the URL is the sandbox login URL.
func IsSandboxURL(url string) bool {
	normalized := normalizeInstanceURL(url)
	return normalized == SandboxLoginURL
}
