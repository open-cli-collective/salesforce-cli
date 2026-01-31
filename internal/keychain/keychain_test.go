package keychain

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestFileStorage(t *testing.T) {
	// Set up temp config directory
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	t.Run("set and get from file", func(t *testing.T) {
		err := setInConfigFile(token)
		require.NoError(t, err)

		retrieved, err := getFromConfigFile()
		require.NoError(t, err)
		assert.Equal(t, token.AccessToken, retrieved.AccessToken)
		assert.Equal(t, token.RefreshToken, retrieved.RefreshToken)
	})

	t.Run("delete from file", func(t *testing.T) {
		err := deleteFromConfigFile()
		require.NoError(t, err)

		_, err = getFromConfigFile()
		assert.ErrorIs(t, err, ErrTokenNotFound)
	})

	t.Run("get nonexistent returns error", func(t *testing.T) {
		_, err := getFromConfigFile()
		assert.ErrorIs(t, err, ErrTokenNotFound)
	})
}

func TestSecureDelete(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-token.json")

	// Create a test file
	content := []byte(`{"access_token":"secret"}`)
	err := os.WriteFile(testFile, content, 0600)
	require.NoError(t, err)

	// Securely delete it
	err = secureDelete(testFile)
	require.NoError(t, err)

	// File should be gone
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

func TestSecureDeleteNonexistent(t *testing.T) {
	// Should not error on nonexistent file
	err := secureDelete("/nonexistent/path/to/file")
	assert.NoError(t, err)
}

func TestHasStoredToken(t *testing.T) {
	// Set up temp config directory
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	t.Run("false when no token", func(t *testing.T) {
		assert.False(t, HasStoredToken())
	})

	t.Run("true when token exists", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: "test-token",
			Expiry:      time.Now().Add(time.Hour),
		}
		err := setInConfigFile(token)
		require.NoError(t, err)

		assert.True(t, HasStoredToken())

		// Cleanup
		deleteFromConfigFile()
	})
}

func TestIsSecureStorage(t *testing.T) {
	// This test just verifies the function doesn't panic
	// Actual behavior depends on platform
	result := IsSecureStorage()
	// Should be a boolean
	assert.IsType(t, true, result)
}

func TestGetStorageBackend(t *testing.T) {
	// This test just verifies the function returns a valid backend
	backend := GetStorageBackend()
	validBackends := []StorageBackend{BackendKeychain, BackendSecretTool, BackendFile}
	assert.Contains(t, validBackends, backend)
}
