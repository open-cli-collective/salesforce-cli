package searchcmd

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

func TestSearchCommand(t *testing.T) {
	searchResult := api.SearchResult{
		SearchRecords: []api.SearchRecord{
			{
				Attributes: api.SObjectAttributes{Type: "Account"},
				ID:         "001xx000001",
				Fields:     map[string]interface{}{"Name": "Acme Corp"},
			},
			{
				Attributes: api.SObjectAttributes{Type: "Contact"},
				ID:         "003xx000001",
				Fields:     map[string]interface{}{"Name": "John Doe"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/search")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResult)
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
	cmd.SetArgs([]string{"Acme"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Account")
	assert.Contains(t, output, "001xx000001")
	assert.Contains(t, output, "Contact")
	assert.Contains(t, output, "003xx000001")
	assert.Contains(t, output, "2 record(s) found")
}

func TestSearchCommand_NoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.SearchResult{
			SearchRecords: []api.SearchRecord{},
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
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "No records found")
}

func TestSearchCommand_JSONOutput(t *testing.T) {
	searchResult := api.SearchResult{
		SearchRecords: []api.SearchRecord{
			{
				Attributes: api.SObjectAttributes{Type: "Account"},
				ID:         "001xx000001",
				Fields:     map[string]interface{}{"Name": "Test"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResult)
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
	cmd.SetArgs([]string{"Test"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	var result api.SearchResult
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result.SearchRecords, 1)
}

func TestBuildSOSL(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		inObjects string
		returning string
		want      string
	}{
		{
			name:  "simple query",
			query: "Acme",
			want:  "FIND {Acme}",
		},
		{
			name:      "with in objects",
			query:     "Acme",
			inObjects: "Account,Contact",
			want:      "FIND {Acme} IN ALL FIELDS RETURNING Account,Contact",
		},
		{
			name:      "with returning",
			query:     "Acme",
			returning: "Account(Id,Name),Contact(Id,Email)",
			want:      "FIND {Acme} RETURNING Account(Id,Name),Contact(Id,Email)",
		},
		{
			name:  "raw SOSL passthrough",
			query: "FIND {Acme} IN NAME FIELDS RETURNING Account(Id,Name)",
			want:  "FIND {Acme} IN NAME FIELDS RETURNING Account(Id,Name)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSOSL(tt.query, tt.inObjects, tt.returning)
			assert.Equal(t, tt.want, got)
		})
	}
}
