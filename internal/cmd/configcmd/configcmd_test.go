package configcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "config", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	// Check subcommands exist
	subcommands := cmd.Commands()
	var subNames []string
	for _, sub := range subcommands {
		subNames = append(subNames, sub.Use)
	}

	assert.Contains(t, subNames, "show")
	assert.Contains(t, subNames, "test")
	assert.Contains(t, subNames, "clear")
}

func TestMaskClientID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "long client ID",
			input:    "3MVG9nKNqSNYF2dG9y7eJzIxOtLw.abc123xyz",
			expected: "3MVG...3xyz",
		},
		{
			name:     "short client ID",
			input:    "short123",
			expected: "****",
		},
		{
			name:     "exactly 12 chars",
			input:    "123456789012",
			expected: "****",
		},
		{
			name:     "13 chars",
			input:    "1234567890123",
			expected: "1234...0123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskClientID(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"login.salesforce.com", "https://login.salesforce.com"},
		{"https://login.salesforce.com", "https://login.salesforce.com"},
		{"https://login.salesforce.com/", "https://login.salesforce.com"},
		{"  login.salesforce.com  ", "https://login.salesforce.com"},
		{"test.salesforce.com", "https://test.salesforce.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeURL(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
