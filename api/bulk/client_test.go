package bulk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ClientConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: ClientConfig{
				InstanceURL: "https://test.salesforce.com",
				HTTPClient:  &http.Client{},
			},
			wantErr: false,
		},
		{
			name: "missing instance URL",
			cfg: ClientConfig{
				HTTPClient: &http.Client{},
			},
			wantErr: true,
		},
		{
			name: "missing HTTP client",
			cfg: ClientConfig{
				InstanceURL: "https://test.salesforce.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}

func TestCreateJob(t *testing.T) {
	expectedJob := JobInfo{
		ID:        "750xx000000001",
		Operation: OperationInsert,
		Object:    "Account",
		State:     StateOpen,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/jobs/ingest")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedJob)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	job, err := client.CreateJob(context.Background(), JobConfig{
		Object:    "Account",
		Operation: OperationInsert,
	})
	require.NoError(t, err)
	assert.Equal(t, expectedJob.ID, job.ID)
	assert.Equal(t, StateOpen, job.State)
}

func TestUploadJobData(t *testing.T) {
	csvData := []byte("Name,Industry\nAcme,Technology")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Contains(t, r.URL.Path, "/jobs/ingest/750xx000000001/batches")
		assert.Equal(t, "text/csv", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	err = client.UploadJobData(context.Background(), "750xx000000001", csvData)
	require.NoError(t, err)
}

func TestCloseJob(t *testing.T) {
	expectedJob := JobInfo{
		ID:    "750xx000000001",
		State: StateUploadComplete,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Contains(t, r.URL.Path, "/jobs/ingest/750xx000000001")

		var req UpdateJobRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, StateUploadComplete, req.State)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedJob)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	job, err := client.CloseJob(context.Background(), "750xx000000001")
	require.NoError(t, err)
	assert.Equal(t, StateUploadComplete, job.State)
}

func TestGetJob(t *testing.T) {
	expectedJob := JobInfo{
		ID:                     "750xx000000001",
		State:                  StateJobComplete,
		NumberRecordsProcessed: 100,
		NumberRecordsFailed:    2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedJob)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	job, err := client.GetJob(context.Background(), "750xx000000001")
	require.NoError(t, err)
	assert.Equal(t, 100, job.NumberRecordsProcessed)
	assert.Equal(t, 2, job.NumberRecordsFailed)
}

func TestListJobs(t *testing.T) {
	expected := JobsResponse{
		Done: true,
		Records: []JobInfo{
			{ID: "750xx000000001", State: StateJobComplete},
			{ID: "750xx000000002", State: StateInProgress},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/jobs/ingest")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	resp, err := client.ListJobs(context.Background())
	require.NoError(t, err)
	assert.Len(t, resp.Records, 2)
}

func TestGetSuccessfulResults(t *testing.T) {
	csvData := "sf__Id,sf__Created,Name\n001xx000001,true,Acme"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/successfulResults")
		assert.Equal(t, "text/csv", r.Header.Get("Accept"))
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(csvData))
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	data, err := client.GetSuccessfulResults(context.Background(), "750xx000000001")
	require.NoError(t, err)
	assert.Equal(t, csvData, string(data))
}

func TestGetFailedResults(t *testing.T) {
	csvData := "sf__Id,sf__Error,Name\n,REQUIRED_FIELD_MISSING,Acme"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/failedResults")
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(csvData))
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	data, err := client.GetFailedResults(context.Background(), "750xx000000001")
	require.NoError(t, err)
	assert.Contains(t, string(data), "REQUIRED_FIELD_MISSING")
}

func TestCreateQueryJob(t *testing.T) {
	expectedJob := QueryJobInfo{
		ID:        "750xx000000001",
		Operation: OperationQuery,
		Query:     "SELECT Id FROM Account",
		State:     StateUploadComplete,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/jobs/query")

		var req CreateQueryJobRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "SELECT Id FROM Account", req.Query)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedJob)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	job, err := client.CreateQueryJob(context.Background(), QueryConfig{
		Query: "SELECT Id FROM Account",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedJob.ID, job.ID)
	assert.Equal(t, OperationQuery, job.Operation)
}

func TestGetQueryResults(t *testing.T) {
	csvData := "Id,Name\n001xx000001,Acme\n001xx000002,Test"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/jobs/query/750xx000000001/results")
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(csvData))
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	data, err := client.GetQueryResults(context.Background(), "750xx000000001")
	require.NoError(t, err)
	assert.Contains(t, string(data), "Acme")
	assert.Contains(t, string(data), "Test")
}

func TestAbortJob(t *testing.T) {
	expectedJob := JobInfo{
		ID:    "750xx000000001",
		State: StateAborted,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)

		var req UpdateJobRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, StateAborted, req.State)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedJob)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	job, err := client.AbortJob(context.Background(), "750xx000000001")
	require.NoError(t, err)
	assert.Equal(t, StateAborted, job.State)
}
