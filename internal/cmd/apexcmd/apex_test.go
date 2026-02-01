package apexcmd

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

func TestApexListClasses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                    "01p000000000001",
					"Name":                  "MyController",
					"Status":                "Active",
					"IsValid":               true,
					"ApiVersion":            float64(62),
					"LengthWithoutComments": float64(500),
				},
				{
					"Id":                    "01p000000000002",
					"Name":                  "MyHelper",
					"Status":                "Active",
					"IsValid":               true,
					"ApiVersion":            float64(62),
					"LengthWithoutComments": float64(300),
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
	assert.Contains(t, output, "MyController")
	assert.Contains(t, output, "MyHelper")
	assert.Contains(t, output, "2 class(es)")
}

func TestApexListTriggers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":            "01q000000000001",
					"Name":          "AccountTrigger",
					"Status":        "Active",
					"IsValid":       true,
					"ApiVersion":    float64(62),
					"TableEnumOrId": "Account",
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
	cmd.SetArgs([]string{"list", "--triggers"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "AccountTrigger")
	assert.Contains(t, output, "Account")
}

func TestApexGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":         "01p000000000001",
					"Name":       "MyController",
					"Body":       "public class MyController {\n    public void doSomething() { }\n}",
					"Status":     "Active",
					"IsValid":    true,
					"ApiVersion": float64(62),
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
		Output: "plain",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"get", "MyController"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "public class MyController")
	assert.Contains(t, output, "doSomething")
}

func TestApexExecuteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.ExecuteAnonymousResult{
			Compiled: true,
			Success:  true,
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
	cmd.SetArgs([]string{"execute", "System.debug('Hello');"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Executed successfully")
}

func TestApexExecuteCompileError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.ExecuteAnonymousResult{
			Line:           1,
			Column:         10,
			Compiled:       false,
			Success:        false,
			CompileProblem: "Variable does not exist: foo",
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
	stderr := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: stderr,
	}
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"execute", "System.debug(foo);"})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compilation failed")
}

func TestApexExecuteFromStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.ExecuteAnonymousResult{
			Compiled: true,
			Success:  true,
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

	stdin := strings.NewReader("System.debug('From stdin');")
	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"execute", "-"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Executed successfully")
}

func TestApexTestNoWait(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.RawQuery, "ApexClass") && strings.Contains(r.URL.RawQuery, "Id") && strings.Contains(r.URL.RawQuery, "MyTest") {
			// Get class ID
			response := tooling.QueryResult{
				TotalSize: 1,
				Done:      true,
				Records: []tooling.Record{
					{"Id": "01p000000000001"},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		} else if strings.Contains(r.URL.Path, "runTestsAsynchronous") {
			// Run tests
			w.Write([]byte(`"7071x00000ABCDE"`))
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
	cmd.SetArgs([]string{"test", "--class", "MyTest"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "7071x00000ABCDE")
	assert.Contains(t, output, "Tests enqueued")
}

func TestApexGetTrigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "ApexTrigger")

		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":            "01q000000000001",
					"Name":          "AccountTrigger",
					"Body":          "trigger AccountTrigger on Account (before insert) { }",
					"Status":        "Active",
					"IsValid":       true,
					"ApiVersion":    float64(62),
					"TableEnumOrId": "Account",
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
		Output: "plain",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"get", "AccountTrigger", "--trigger"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "trigger AccountTrigger on Account")
}

func TestApexExecuteRuntimeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.ExecuteAnonymousResult{
			Compiled:            true,
			Success:             false,
			ExceptionMessage:    "System.NullPointerException: Attempt to de-reference a null object",
			ExceptionStackTrace: "AnonymousBlock: line 1, column 1",
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
	stderr := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: stderr,
	}
	opts.SetToolingClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"execute", "String s; s.length();"})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution failed")
	assert.Contains(t, stderr.String(), "NullPointerException")
}

func TestApexListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tooling.QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []tooling.Record{
				{
					"Id":                    "01p000000000001",
					"Name":                  "MyController",
					"Status":                "Active",
					"IsValid":               true,
					"ApiVersion":            float64(62),
					"LengthWithoutComments": float64(500),
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
	assert.Contains(t, output, "MyController")
	// Should be valid JSON
	var result []tooling.ApexClass
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}
