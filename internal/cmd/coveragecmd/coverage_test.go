package coveragecmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-cli-collective/salesforce-cli/api/tooling"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func TestCoverageList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "500000000000001",
					"ApexClassOrTriggerId": "01p000000000001",
					"ApexClassOrTrigger":   map[string]interface{}{"Name": "MyController"},
					"NumLinesCovered":      float64(80),
					"NumLinesUncovered":    float64(20),
				},
				{
					"Id":                   "500000000000002",
					"ApexClassOrTriggerId": "01p000000000002",
					"ApexClassOrTrigger":   map[string]interface{}{"Name": "MyHelper"},
					"NumLinesCovered":      float64(50),
					"NumLinesUncovered":    float64(50),
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
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "MyController")
	assert.Contains(t, output, "MyHelper")
	assert.Contains(t, output, "80.0%")
	assert.Contains(t, output, "50.0%")
	assert.Contains(t, output, "Overall")
}

func TestCoverageForClass(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "MyController")

		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "500000000000001",
					"ApexClassOrTriggerId": "01p000000000001",
					"ApexClassOrTrigger":   map[string]interface{}{"Name": "MyController"},
					"NumLinesCovered":      float64(80),
					"NumLinesUncovered":    float64(20),
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
	cmd.SetArgs([]string{"--class", "MyController"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "MyController")
	assert.Contains(t, output, "80.0%")
}

func TestCoverageMinimumPass(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "500000000000001",
					"ApexClassOrTriggerId": "01p000000000001",
					"ApexClassOrTrigger":   map[string]interface{}{"Name": "MyController"},
					"NumLinesCovered":      float64(80),
					"NumLinesUncovered":    float64(20),
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
	cmd.SetArgs([]string{"--min", "75"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err) // 80% > 75%, should pass
}

func TestCoverageMinimumFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "500000000000001",
					"ApexClassOrTriggerId": "01p000000000001",
					"ApexClassOrTrigger":   map[string]interface{}{"Name": "MyController"},
					"NumLinesCovered":      float64(50),
					"NumLinesUncovered":    float64(50),
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
	cmd.SetArgs([]string{"--min", "75"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	assert.Error(t, err) // 50% < 75%, should fail
	assert.Contains(t, err.Error(), "below minimum")
}

func TestCoverageJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                   "500000000000001",
					"ApexClassOrTriggerId": "01p000000000001",
					"ApexClassOrTrigger":   map[string]interface{}{"Name": "MyController"},
					"NumLinesCovered":      float64(80),
					"NumLinesUncovered":    float64(20),
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
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	// Should be valid JSON
	var result []tooling.ApexCodeCoverageAggregate
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestCoverageEmpty(t *testing.T) {
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
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "No code coverage data found")
}

func TestCoverageClassNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "NonExistent") {
			response := tooling.QueryResult{
				TotalSize: 0,
				Done:      true,
				Records:   []tooling.Record{},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
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
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"--class", "NonExistent"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no coverage data found")
}
