package querycmd

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

func TestQueryCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		serverResponse api.QueryResult
		wantErr        bool
		wantContains   []string
	}{
		{
			name: "simple query with results",
			args: []string{"SELECT Id, Name FROM Account"},
			serverResponse: api.QueryResult{
				TotalSize: 2,
				Done:      true,
				Records: []api.SObject{
					{ID: "001xx000001", Fields: map[string]interface{}{"Name": "Acme Corp"}},
					{ID: "001xx000002", Fields: map[string]interface{}{"Name": "Test Inc"}},
				},
			},
			wantContains: []string{"001xx000001", "Acme Corp", "001xx000002", "Test Inc"},
		},
		{
			name: "query with no results",
			args: []string{"SELECT Id FROM Account WHERE Name = 'NonExistent'"},
			serverResponse: api.QueryResult{
				TotalSize: 0,
				Done:      true,
				Records:   []api.SObject{},
			},
			wantContains: []string{"No records found"},
		},
		{
			name: "query with pagination indicator",
			args: []string{"SELECT Id FROM Account"},
			serverResponse: api.QueryResult{
				TotalSize:      1000,
				Done:           false,
				NextRecordsURL: "/services/data/v62.0/query/01gxx0000000001-500",
				Records: []api.SObject{
					{ID: "001xx000001", Fields: map[string]interface{}{}},
				},
			},
			wantContains: []string{"Showing 1 of 1000 records"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			// Create client pointing to test server
			client, err := api.New(api.ClientConfig{
				InstanceURL: server.URL,
				HTTPClient:  server.Client(),
			})
			require.NoError(t, err)

			// Set up options with test client
			stdout := &bytes.Buffer{}
			opts := &root.Options{
				Output: "table",
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
			}
			opts.SetAPIClient(client)

			// Create and execute command
			cmd := NewCommand(opts)
			cmd.SetArgs(tt.args)
			cmd.SetOut(stdout)
			cmd.SetErr(&bytes.Buffer{})

			err = cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			output := stdout.String()
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "output should contain %q", want)
			}
		})
	}
}

func TestQueryCommand_JSONOutput(t *testing.T) {
	serverResponse := api.QueryResult{
		TotalSize: 1,
		Done:      true,
		Records: []api.SObject{
			{ID: "001xx000001", Fields: map[string]interface{}{"Name": "Test"}},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(serverResponse)
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "json",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetAPIClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"SELECT Id, Name FROM Account"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify JSON output
	var result api.QueryResult
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalSize)
	assert.Len(t, result.Records, 1)
}

func TestQueryCommand_AllFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it hits the queryAll endpoint
		assert.Contains(t, r.URL.Path, "/queryAll")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []api.SObject{
				{ID: "001xx000001", Fields: map[string]interface{}{"IsDeleted": true}},
			},
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

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"SELECT Id FROM Account", "--all"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)
}

func TestFormatFieldValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"integer float", float64(42), "42"},
		{"decimal float", 3.14, "3.14"},
		{"true bool", true, "true"},
		{"false bool", false, "false"},
		{"nested object with Name", map[string]interface{}{"Name": "Related"}, "Related"},
		{"nested object without Name", map[string]interface{}{"Id": "123"}, "[object]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFieldValue(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}
