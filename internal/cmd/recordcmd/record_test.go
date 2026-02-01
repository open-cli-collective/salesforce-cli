package recordcmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-cli-collective/salesforce-cli/api"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func TestGetCommand(t *testing.T) {
	record := api.SObject{
		ID: "001xx000001",
		Attributes: api.SObjectAttributes{
			Type: "Account",
			URL:  "/services/data/v62.0/sobjects/Account/001xx000001",
		},
		Fields: map[string]interface{}{
			"Name":     "Acme Corp",
			"Industry": "Technology",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/sobjects/Account/001xx000001")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(record)
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetAPIClient(client)

	cmd := newGetCommand(opts)
	cmd.SetArgs([]string{"Account", "001xx000001"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Account")
	assert.Contains(t, output, "001xx000001")
	assert.Contains(t, output, "Acme Corp")
	assert.Contains(t, output, "Technology")
}

func TestGetCommand_WithFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify fields parameter is passed
		assert.Contains(t, r.URL.RawQuery, "fields=Name,Phone")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.SObject{
			ID:         "001xx000001",
			Attributes: api.SObjectAttributes{Type: "Account"},
			Fields:     map[string]interface{}{"Name": "Test", "Phone": "555-1234"},
		})
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetAPIClient(client)

	cmd := newGetCommand(opts)
	cmd.SetArgs([]string{"Account", "001xx000001", "--fields", "Name,Phone"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)
}

func TestCreateCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/sobjects/Account")

		// Verify request body
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Acme Corp", body["Name"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.RecordResult{
			ID:      "001xx000001",
			Success: true,
		})
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetAPIClient(client)

	cmd := newCreateCommand(opts)
	cmd.SetArgs([]string{"Account", "--set", "Name=Acme Corp"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Created")
	assert.Contains(t, output, "001xx000001")
}

func TestCreateCommand_NoFields(t *testing.T) {
	opts := &root.Options{
		Output: "table",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	cmd := newCreateCommand(opts)
	cmd.SetArgs([]string{"Account"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one --set flag")
}

func TestUpdateCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Contains(t, r.URL.Path, "/sobjects/Account/001xx000001")

		// Verify request body
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "555-1234", body["Phone"])

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetAPIClient(client)

	cmd := newUpdateCommand(opts)
	cmd.SetArgs([]string{"Account", "001xx000001", "--set", "Phone=555-1234"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Updated")
}

func TestDeleteCommand_WithConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/sobjects/Account/001xx000001")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetAPIClient(client)

	cmd := newDeleteCommand(opts)
	cmd.SetArgs([]string{"Account", "001xx000001", "--confirm"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Deleted")
}

func TestParseSetFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:  "simple string",
			flags: []string{"Name=Acme"},
			want:  map[string]interface{}{"Name": "Acme"},
		},
		{
			name:  "quoted string",
			flags: []string{`Name="Acme Corp"`},
			want:  map[string]interface{}{"Name": "Acme Corp"},
		},
		{
			name:  "boolean true",
			flags: []string{"IsActive=true"},
			want:  map[string]interface{}{"IsActive": true},
		},
		{
			name:  "boolean false",
			flags: []string{"IsActive=false"},
			want:  map[string]interface{}{"IsActive": false},
		},
		{
			name:  "null value",
			flags: []string{"Description=null"},
			want:  map[string]interface{}{"Description": nil},
		},
		{
			name:  "multiple fields",
			flags: []string{"Name=Test", "Phone=555-1234", "IsActive=true"},
			want: map[string]interface{}{
				"Name":     "Test",
				"Phone":    "555-1234",
				"IsActive": true,
			},
		},
		{
			name:    "invalid format",
			flags:   []string{"InvalidNoEquals"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSetFlags(tt.flags)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
