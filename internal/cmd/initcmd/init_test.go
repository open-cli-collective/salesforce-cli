package initcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractAuthCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "raw code",
			input: "abc123xyz",
			want:  "abc123xyz",
		},
		{
			name:  "localhost URL with code",
			input: "http://localhost:8080/callback?code=abc123xyz",
			want:  "abc123xyz",
		},
		{
			name:  "localhost URL with code and state",
			input: "http://localhost:8080/callback?code=abc123xyz&state=state-token",
			want:  "abc123xyz",
		},
		{
			name:  "URL with error",
			input: "http://localhost:8080/callback?error=access_denied",
			want:  "",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace",
			input: "  abc123  ",
			want:  "abc123",
		},
		{
			name:  "localhost URL without port",
			input: "http://localhost/?code=abc123xyz",
			want:  "abc123xyz",
		},
		{
			name:  "https localhost URL",
			input: "https://localhost/?code=SecureCode456",
			want:  "SecureCode456",
		},
		{
			name:  "code with special characters",
			input: "http://localhost:8080/callback?code=4/P-abc_123.xyz~456",
			want:  "4/P-abc_123.xyz~456",
		},
		{
			name:  "URL encoded code",
			input: "http://localhost:8080/callback?code=4%2F0AQSTgQ",
			want:  "4/0AQSTgQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAuthCode(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "init", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Check flags exist
	assert.NotNil(t, cmd.Flags().Lookup("instance-url"))
	assert.NotNil(t, cmd.Flags().Lookup("client-id"))
	assert.NotNil(t, cmd.Flags().Lookup("no-verify"))

	// --no-browser flag was removed (no more callback server or auto-browser-opening)
	assert.Nil(t, cmd.Flags().Lookup("no-browser"))
}
