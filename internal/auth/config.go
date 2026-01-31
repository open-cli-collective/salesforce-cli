// Package auth provides OAuth 2.0 authentication for Salesforce.
package auth

import (
	"github.com/open-cli-collective/salesforce-cli/internal/config"
)

// GetCredentialsPath returns the full path to the config file.
// Re-exported from config package for convenience.
func GetCredentialsPath() (string, error) {
	return config.GetConfigPath()
}

// GetTokenPath returns the full path to the token file (fallback storage).
// Re-exported from config package for convenience.
func GetTokenPath() (string, error) {
	return config.GetTokenPath()
}

// ShortenPath replaces the home directory prefix with ~ for display.
// Re-exported from config package for convenience.
func ShortenPath(path string) string {
	return config.ShortenPath(path)
}
