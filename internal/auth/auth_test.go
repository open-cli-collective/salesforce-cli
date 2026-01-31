package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOAuthConfig(t *testing.T) {
	config := GetOAuthConfig("https://login.salesforce.com", "test-client-id")

	assert.Equal(t, "test-client-id", config.ClientID)
	assert.Equal(t, "https://login.salesforce.com/services/oauth2/authorize", config.Endpoint.AuthURL)
	assert.Equal(t, "https://login.salesforce.com/services/oauth2/token", config.Endpoint.TokenURL)
	assert.Equal(t, "http://localhost:8080/callback", config.RedirectURL)
	assert.Contains(t, config.Scopes, "api")
	assert.Contains(t, config.Scopes, "refresh_token")
}

func TestGetOAuthConfig_CustomDomain(t *testing.T) {
	config := GetOAuthConfig("mycompany.my.salesforce.com", "test-client-id")

	assert.Equal(t, "https://mycompany.my.salesforce.com/services/oauth2/authorize", config.Endpoint.AuthURL)
	assert.Equal(t, "https://mycompany.my.salesforce.com/services/oauth2/token", config.Endpoint.TokenURL)
}

func TestNormalizeInstanceURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"login.salesforce.com", "https://login.salesforce.com"},
		{"https://login.salesforce.com", "https://login.salesforce.com"},
		{"https://login.salesforce.com/", "https://login.salesforce.com"},
		{"test.salesforce.com", "https://test.salesforce.com"},
		{"mycompany.my.salesforce.com", "https://mycompany.my.salesforce.com"},
		{"  login.salesforce.com  ", "https://login.salesforce.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeInstanceURL(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsProductionURL(t *testing.T) {
	assert.True(t, IsProductionURL("login.salesforce.com"))
	assert.True(t, IsProductionURL("https://login.salesforce.com"))
	assert.False(t, IsProductionURL("test.salesforce.com"))
	assert.False(t, IsProductionURL("mycompany.my.salesforce.com"))
}

func TestIsSandboxURL(t *testing.T) {
	assert.True(t, IsSandboxURL("test.salesforce.com"))
	assert.True(t, IsSandboxURL("https://test.salesforce.com"))
	assert.False(t, IsSandboxURL("login.salesforce.com"))
	assert.False(t, IsSandboxURL("mycompany.my.salesforce.com"))
}

func TestGetAuthURL(t *testing.T) {
	config := GetOAuthConfig("https://login.salesforce.com", "test-client-id")
	url := GetAuthURL(config)

	assert.Contains(t, url, "https://login.salesforce.com/services/oauth2/authorize")
	assert.Contains(t, url, "client_id=test-client-id")
	assert.Contains(t, url, "redirect_uri=")
	assert.Contains(t, url, "response_type=code")
}
