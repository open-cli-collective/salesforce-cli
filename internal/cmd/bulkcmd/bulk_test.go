package bulkcmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-cli-collective/salesforce-cli/api/bulk"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func TestImportCommand(t *testing.T) {
	expectedJob := bulk.JobInfo{
		ID:        "750xx000000001",
		Operation: bulk.OperationInsert,
		Object:    "Account",
		State:     bulk.StateJobComplete,
	}

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/services/data/v62.0/jobs/ingest":
			// Create job
			expectedJob.State = bulk.StateOpen
			_ = json.NewEncoder(w).Encode(expectedJob)
		case r.Method == http.MethodPut:
			// Upload data
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPatch:
			// Close job
			expectedJob.State = bulk.StateUploadComplete
			_ = json.NewEncoder(w).Encode(expectedJob)
		}
	}))
	defer server.Close()

	client, err := bulk.New(bulk.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	// Create temp CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "accounts.csv")
	err = os.WriteFile(csvFile, []byte("Name,Industry\nAcme,Technology"), 0644)
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: stderr,
	}
	opts.SetBulkClient(client)

	cmd := newImportCommand(opts)
	cmd.SetArgs([]string{"Account", "--file", csvFile, "--operation", "insert"})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Creating bulk insert job")
	assert.Contains(t, output, "750xx000000001")
}

func TestImportCommand_UpsertRequiresExternalID(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: stderr,
	}

	// Create temp CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "contacts.csv")
	err := os.WriteFile(csvFile, []byte("Email,Name\ntest@test.com,Test"), 0644)
	require.NoError(t, err)

	cmd := newImportCommand(opts)
	cmd.SetArgs([]string{"Contact", "--file", csvFile, "--operation", "upsert"})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--external-id is required")
}

func TestExportCommand(t *testing.T) {
	csvData := "Id,Name\n001xx000001,Acme\n001xx000002,Test"
	expectedJob := bulk.QueryJobInfo{
		ID:                     "750xx000000001",
		Operation:              bulk.OperationQuery,
		State:                  bulk.StateJobComplete,
		NumberRecordsProcessed: 2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/services/data/v62.0/jobs/query":
			// Create query job
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(expectedJob)
		case r.Method == http.MethodGet && r.URL.Path == "/services/data/v62.0/jobs/query/750xx000000001":
			// Get job status
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(expectedJob)
		case r.Method == http.MethodGet && r.URL.Path == "/services/data/v62.0/jobs/query/750xx000000001/results":
			// Get results
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write([]byte(csvData))
		}
	}))
	defer server.Close()

	client, err := bulk.New(bulk.ClientConfig{
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
	opts.SetBulkClient(client)

	cmd := newExportCommand(opts)
	cmd.SetArgs([]string{"SELECT Id, Name FROM Account"})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Id,Name")
	assert.Contains(t, output, "Acme")
}

func TestJobListCommand(t *testing.T) {
	expected := bulk.JobsResponse{
		Done: true,
		Records: []bulk.JobInfo{
			{ID: "750xx000000001", Object: "Account", Operation: bulk.OperationInsert, State: bulk.StateJobComplete},
			{ID: "750xx000000002", Object: "Contact", Operation: bulk.OperationUpdate, State: bulk.StateInProgress},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client, err := bulk.New(bulk.ClientConfig{
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
	opts.SetBulkClient(client)

	cmd := newJobListCommand(opts)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "750xx000000001")
	assert.Contains(t, output, "Account")
	assert.Contains(t, output, "2 job(s)")
}

func TestJobStatusCommand(t *testing.T) {
	expectedJob := bulk.JobInfo{
		ID:                     "750xx000000001",
		Object:                 "Account",
		Operation:              bulk.OperationInsert,
		State:                  bulk.StateJobComplete,
		NumberRecordsProcessed: 100,
		NumberRecordsFailed:    2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedJob)
	}))
	defer server.Close()

	client, err := bulk.New(bulk.ClientConfig{
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
	opts.SetBulkClient(client)

	cmd := newJobStatusCommand(opts)
	cmd.SetArgs([]string{"750xx000000001"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "JobComplete")
	assert.Contains(t, output, "100")
	assert.Contains(t, output, "2")
}

func TestJobAbortCommand(t *testing.T) {
	expectedJob := bulk.JobInfo{
		ID:    "750xx000000001",
		State: bulk.StateAborted,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedJob)
	}))
	defer server.Close()

	client, err := bulk.New(bulk.ClientConfig{
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
	opts.SetBulkClient(client)

	cmd := newJobAbortCommand(opts)
	cmd.SetArgs([]string{"750xx000000001"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "aborted")
}

func TestJobResultsCommand(t *testing.T) {
	csvData := "sf__Id,sf__Created,Name\n001xx000001,true,Acme"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(csvData))
	}))
	defer server.Close()

	client, err := bulk.New(bulk.ClientConfig{
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
	opts.SetBulkClient(client)

	cmd := newJobResultsCommand(opts)
	cmd.SetArgs([]string{"750xx000000001"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "sf__Id")
	assert.Contains(t, output, "Acme")
}

func TestJobErrorsCommand(t *testing.T) {
	csvData := "sf__Id,sf__Error,Name\n,REQUIRED_FIELD_MISSING,Acme"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(csvData))
	}))
	defer server.Close()

	client, err := bulk.New(bulk.ClientConfig{
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
	opts.SetBulkClient(client)

	cmd := newJobErrorsCommand(opts)
	cmd.SetArgs([]string{"750xx000000001"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "REQUIRED_FIELD_MISSING")
}
