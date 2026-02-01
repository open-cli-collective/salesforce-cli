package logcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-cli-collective/salesforce-cli/api/tooling"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func TestLogList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "07L000000000001",
					"LogUserId":            "005000000000001",
					"Operation":            "/aura",
					"Request":              "API",
					"Status":               "Success",
					"LogLength":            float64(5000),
					"DurationMilliseconds": float64(150),
					"StartTime":            "2024-01-15T10:30:00.000+0000",
					"Location":             "MonitoringService",
				},
				{
					"Id":                   "07L000000000002",
					"LogUserId":            "005000000000001",
					"Operation":            "ApexTrigger",
					"Request":              "API",
					"Status":               "Success",
					"LogLength":            float64(2500),
					"DurationMilliseconds": float64(75),
					"StartTime":            "2024-01-15T10:25:00.000+0000",
					"Location":             "MonitoringService",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := tooling.New(tooling.ClientConfig{
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
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"list"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "07L000000000001")
	assert.Contains(t, output, "/aura")
	assert.Contains(t, output, "2 log(s)")
}

func TestLogListWithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify limit is in query
		assert.Contains(t, r.URL.RawQuery, "LIMIT+5")

		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "07L000000000001",
					"LogUserId":            "005000000000001",
					"Operation":            "/aura",
					"Request":              "API",
					"Status":               "Success",
					"LogLength":            float64(5000),
					"DurationMilliseconds": float64(150),
					"StartTime":            "2024-01-15T10:30:00.000+0000",
					"Location":             "MonitoringService",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := tooling.New(tooling.ClientConfig{
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
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"list", "--limit", "5"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)
}

func TestLogGet(t *testing.T) {
	logContent := "DEBUG|Hello World\nUSER_DEBUG|Test message"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/Body") {
			w.Write([]byte(logContent))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := tooling.New(tooling.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "plain",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"get", "07L000000000001"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "DEBUG|Hello World")
	assert.Contains(t, output, "USER_DEBUG|Test message")
}

func TestLogListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "07L000000000001",
					"LogUserId":            "005000000000001",
					"Operation":            "/aura",
					"Request":              "API",
					"Status":               "Success",
					"LogLength":            float64(5000),
					"DurationMilliseconds": float64(150),
					"StartTime":            "2024-01-15T10:30:00.000+0000",
					"Location":             "MonitoringService",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := tooling.New(tooling.ClientConfig{
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
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"list"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	// Should be valid JSON
	var result []tooling.ApexLog
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestLogListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 0,
			Done:      true,
			Records:   []tooling.Record{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := tooling.New(tooling.ClientConfig{
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
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"list"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "No debug logs found")
}

func TestLogTailContextCancellation(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := tooling.QueryResult{
			TotalSize: 0,
			Done:      true,
			Records:   []tooling.Record{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := tooling.New(tooling.ClientConfig{
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
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"tail", "--interval", "1"})
	cmd.SetOut(stdout)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Tailing debug logs")
	// Should have made at least one API call
	assert.GreaterOrEqual(t, callCount, 1)
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{2097152, "2.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatSize(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}
