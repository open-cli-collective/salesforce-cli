package limitscmd

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

func TestLimitsCommand(t *testing.T) {
	limits := api.Limits{
		"DailyApiRequests":     api.LimitInfo{Max: 100000, Remaining: 99500},
		"DailyBulkApiRequests": api.LimitInfo{Max: 10000, Remaining: 10000},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/limits")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(limits)
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
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "DailyApiRequests")
	assert.Contains(t, output, "100000")
	assert.Contains(t, output, "99500")
}

func TestLimitsCommand_ShowSpecific(t *testing.T) {
	limits := api.Limits{
		"DailyApiRequests":     api.LimitInfo{Max: 100000, Remaining: 99500},
		"DailyBulkApiRequests": api.LimitInfo{Max: 10000, Remaining: 10000},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(limits)
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
	cmd.SetArgs([]string{"--show", "DailyApiRequests"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "DailyApiRequests")
	assert.Contains(t, output, "Max:")
	assert.Contains(t, output, "Remaining:")
	assert.Contains(t, output, "Used:")
}

func TestLimitsCommand_ShowNotFound(t *testing.T) {
	limits := api.Limits{
		"DailyApiRequests": api.LimitInfo{Max: 100000, Remaining: 99500},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(limits)
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	opts := &root.Options{
		Output: "table",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	opts.SetAPIClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"--show", "NonExistentLimit"})

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLimitsCommand_JSONOutput(t *testing.T) {
	limits := api.Limits{
		"DailyApiRequests": api.LimitInfo{Max: 100000, Remaining: 99500},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(limits)
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
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	var result api.Limits
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 100000, result["DailyApiRequests"].Max)
}
