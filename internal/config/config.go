// Package config provides configuration management for the Salesforce CLI.
// It has no external dependencies to avoid circular imports with other internal packages.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	// DirName is the name of the configuration directory
	DirName = "salesforce-cli"
	// ConfigFile is the name of the configuration file
	ConfigFile = "config.json"
	// TokenFile is the name of the OAuth token file (fallback storage)
	TokenFile = "token.json"
)

// File and directory permission constants for consistent security settings.
const (
	// DirPerm is the permission for config directories (owner read/write/execute only)
	DirPerm = 0700
	// FilePerm is the permission for config files (owner read/write only)
	FilePerm = 0600
)

// Config represents the CLI configuration.
type Config struct {
	// InstanceURL is the Salesforce instance URL (e.g., https://mycompany.my.salesforce.com)
	InstanceURL string `json:"instance_url,omitempty"`
	// ClientID is the OAuth Connected App Consumer Key
	ClientID string `json:"client_id,omitempty"`
}

// GetConfigDir returns the configuration directory path, creating it if needed.
// Uses XDG_CONFIG_HOME if set, otherwise ~/.config/salesforce-cli
func GetConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	configDir := filepath.Join(configHome, DirName)

	if err := os.MkdirAll(configDir, DirPerm); err != nil {
		return "", err
	}

	return configDir, nil
}

// GetConfigPath returns the full path to config.json
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

// GetTokenPath returns the full path to token.json (fallback storage)
func GetTokenPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, TokenFile), nil
}

// ShortenPath replaces the home directory prefix with ~ for display purposes.
// This prevents exposing full paths including usernames in error messages.
func ShortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) >= len(home) && path[:len(home)] == home {
		return "~" + path[len(home):]
	}
	return path
}

// Load loads the configuration from config.json with environment variable overrides.
// Environment variable precedence: SFDC_* → SALESFORCE_* → config file
func Load() (*Config, error) {
	cfg := &Config{}

	// Try to load from file
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// File doesn't exist, continue with empty config
	} else {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// Override with environment variables (SFDC_* takes precedence over SALESFORCE_*)
	if v := getEnvWithFallback("SFDC_INSTANCE_URL", "SALESFORCE_INSTANCE_URL"); v != "" {
		cfg.InstanceURL = v
	}
	if v := getEnvWithFallback("SFDC_CLIENT_ID", "SALESFORCE_CLIENT_ID"); v != "" {
		cfg.ClientID = v
	}

	return cfg, nil
}

// Save saves the configuration to config.json
func Save(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, FilePerm)
}

// Clear removes the configuration file
func Clear() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsConfigured returns true if the minimum required configuration is set.
func IsConfigured() bool {
	cfg, err := Load()
	if err != nil {
		return false
	}
	return cfg.InstanceURL != "" && cfg.ClientID != ""
}

// getEnvWithFallback returns the value of the primary environment variable,
// or the fallback if the primary is not set.
func getEnvWithFallback(primary, fallback string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(fallback)
}
