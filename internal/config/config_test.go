package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDir(t *testing.T) {
	// Save and restore XDG_CONFIG_HOME
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		dir, err := GetConfigDir()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(tmpDir, DirName), dir)

		// Directory should be created
		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")

		dir, err := GetConfigDir()
		require.NoError(t, err)

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", DirName)
		assert.Equal(t, expected, dir)
	})
}

func TestShortenPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{filepath.Join(home, "test"), "~/test"},
		{filepath.Join(home, ".config", "salesforce-cli"), "~/.config/salesforce-cli"},
		{"/some/other/path", "/some/other/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ShortenPath(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadSaveClear(t *testing.T) {
	// Set up temp config directory
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Clear any env overrides
	os.Unsetenv("SFDC_INSTANCE_URL")
	os.Unsetenv("SFDC_CLIENT_ID")
	os.Unsetenv("SALESFORCE_INSTANCE_URL")
	os.Unsetenv("SALESFORCE_CLIENT_ID")

	t.Run("load empty config", func(t *testing.T) {
		cfg, err := Load()
		require.NoError(t, err)
		assert.Empty(t, cfg.InstanceURL)
		assert.Empty(t, cfg.ClientID)
	})

	t.Run("save and load", func(t *testing.T) {
		cfg := &Config{
			InstanceURL: "https://test.salesforce.com",
			ClientID:    "test-client-id",
		}

		err := Save(cfg)
		require.NoError(t, err)

		loaded, err := Load()
		require.NoError(t, err)
		assert.Equal(t, cfg.InstanceURL, loaded.InstanceURL)
		assert.Equal(t, cfg.ClientID, loaded.ClientID)
	})

	t.Run("clear", func(t *testing.T) {
		err := Clear()
		require.NoError(t, err)

		cfg, err := Load()
		require.NoError(t, err)
		assert.Empty(t, cfg.InstanceURL)
	})
}

func TestLoadWithEnvOverrides(t *testing.T) {
	// Set up temp config directory
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Save config file
	cfg := &Config{
		InstanceURL: "https://file.salesforce.com",
		ClientID:    "file-client-id",
	}
	err := Save(cfg)
	require.NoError(t, err)

	t.Run("SFDC_ overrides file", func(t *testing.T) {
		os.Setenv("SFDC_INSTANCE_URL", "https://sfdc.salesforce.com")
		defer os.Unsetenv("SFDC_INSTANCE_URL")

		loaded, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "https://sfdc.salesforce.com", loaded.InstanceURL)
		assert.Equal(t, "file-client-id", loaded.ClientID) // Not overridden
	})

	t.Run("SALESFORCE_ overrides file", func(t *testing.T) {
		os.Unsetenv("SFDC_CLIENT_ID")
		os.Setenv("SALESFORCE_CLIENT_ID", "salesforce-client-id")
		defer os.Unsetenv("SALESFORCE_CLIENT_ID")

		loaded, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "salesforce-client-id", loaded.ClientID)
	})

	t.Run("SFDC_ takes precedence over SALESFORCE_", func(t *testing.T) {
		os.Setenv("SFDC_INSTANCE_URL", "https://sfdc.salesforce.com")
		os.Setenv("SALESFORCE_INSTANCE_URL", "https://salesforce.salesforce.com")
		defer os.Unsetenv("SFDC_INSTANCE_URL")
		defer os.Unsetenv("SALESFORCE_INSTANCE_URL")

		loaded, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "https://sfdc.salesforce.com", loaded.InstanceURL)
	})
}

func TestIsConfigured(t *testing.T) {
	// Set up temp config directory
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Clear env vars
	os.Unsetenv("SFDC_INSTANCE_URL")
	os.Unsetenv("SFDC_CLIENT_ID")
	os.Unsetenv("SALESFORCE_INSTANCE_URL")
	os.Unsetenv("SALESFORCE_CLIENT_ID")

	t.Run("not configured when empty", func(t *testing.T) {
		assert.False(t, IsConfigured())
	})

	t.Run("not configured when partial", func(t *testing.T) {
		cfg := &Config{InstanceURL: "https://test.salesforce.com"}
		err := Save(cfg)
		require.NoError(t, err)

		assert.False(t, IsConfigured())
	})

	t.Run("configured when complete", func(t *testing.T) {
		cfg := &Config{
			InstanceURL: "https://test.salesforce.com",
			ClientID:    "test-client-id",
		}
		err := Save(cfg)
		require.NoError(t, err)

		assert.True(t, IsConfigured())
	})
}
