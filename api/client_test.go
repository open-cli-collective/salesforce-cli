package api

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
		wantErr error
	}{
		{
			name: "valid config",
			cfg: ClientConfig{
				InstanceURL: "https://test.salesforce.com",
				HTTPClient:  http.DefaultClient,
			},
			wantErr: nil,
		},
		{
			name: "with custom API version",
			cfg: ClientConfig{
				InstanceURL: "https://test.salesforce.com",
				HTTPClient:  http.DefaultClient,
				APIVersion:  "v59.0",
			},
			wantErr: nil,
		},
		{
			name: "missing instance URL",
			cfg: ClientConfig{
				HTTPClient: http.DefaultClient,
			},
			wantErr: ErrInstanceURLRequired,
		},
		{
			name: "missing HTTP client",
			cfg: ClientConfig{
				InstanceURL: "https://test.salesforce.com",
			},
			wantErr: ErrHTTPClientRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_NormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.salesforce.com", "https://test.salesforce.com"},
		{"https://test.salesforce.com", "https://test.salesforce.com"},
		{"https://test.salesforce.com/", "https://test.salesforce.com"},
		{"  test.salesforce.com  ", "https://test.salesforce.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeURL(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestClient_BuildURL(t *testing.T) {
	client := &Client{
		InstanceURL: "https://test.salesforce.com",
		APIVersion:  "v62.0",
		BaseURL:     "https://test.salesforce.com/services/data/v62.0",
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "relative path",
			path:     "/query",
			expected: "https://test.salesforce.com/services/data/v62.0/query",
		},
		{
			name:     "relative path without slash",
			path:     "query",
			expected: "https://test.salesforce.com/services/data/v62.0/query",
		},
		{
			name:     "full services path",
			path:     "/services/data/",
			expected: "https://test.salesforce.com/services/data/",
		},
		{
			name:     "absolute URL",
			path:     "https://other.salesforce.com/path",
			expected: "https://other.salesforce.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.buildURL(tt.path)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestClient_Query(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/query")
		assert.Contains(t, r.URL.RawQuery, "q=SELECT+Id%2C+Name+FROM+Account")

		resp := QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []SObject{
				{ID: "001xx000003ABCDEF", Attributes: SObjectAttributes{Type: "Account"}},
				{ID: "001xx000003GHIJKL", Attributes: SObjectAttributes{Type: "Account"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	result, err := client.Query(context.Background(), "SELECT Id, Name FROM Account")
	require.NoError(t, err)

	assert.Equal(t, 2, result.TotalSize)
	assert.True(t, result.Done)
	assert.Len(t, result.Records, 2)
}

func TestClient_QueryAll(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First page
			resp := QueryResult{
				TotalSize:      3,
				Done:           false,
				NextRecordsURL: "/services/data/v62.0/query/01gxx0000000001-2000",
				Records: []SObject{
					{ID: "001"},
					{ID: "002"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			// Second page
			resp := QueryResult{
				TotalSize: 3,
				Done:      true,
				Records: []SObject{
					{ID: "003"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	result, err := client.QueryAll(context.Background(), "SELECT Id FROM Account")
	require.NoError(t, err)

	assert.Equal(t, 2, callCount)
	assert.True(t, result.Done)
	assert.Len(t, result.Records, 3)
}

func TestClient_GetRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/services/data/v62.0/sobjects/Account/001xx000003ABCDEF", r.URL.Path)

		resp := map[string]interface{}{
			"attributes": map[string]string{"type": "Account"},
			"Id":         "001xx000003ABCDEF",
			"Name":       "Test Account",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	record, err := client.GetRecord(context.Background(), "Account", "001xx000003ABCDEF", nil)
	require.NoError(t, err)

	assert.Equal(t, "001xx000003ABCDEF", record.ID)
	assert.Equal(t, "Account", record.Attributes.Type)
	assert.Equal(t, "Test Account", record.GetString("Name"))
}

func TestClient_CreateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/services/data/v62.0/sobjects/Account/", r.URL.Path)

		w.WriteHeader(http.StatusCreated)
		resp := RecordResult{
			ID:      "001xx000003NEWREC",
			Success: true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	result, err := client.CreateRecord(context.Background(), "Account", map[string]interface{}{
		"Name": "New Account",
	})
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "001xx000003NEWREC", result.ID)
}

func TestClient_UpdateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/services/data/v62.0/sobjects/Account/001xx000003ABCDEF", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	err = client.UpdateRecord(context.Background(), "Account", "001xx000003ABCDEF", map[string]interface{}{
		"Name": "Updated Account",
	})
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/services/data/v62.0/sobjects/Account/001xx000003ABCDEF", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	err = client.DeleteRecord(context.Background(), "Account", "001xx000003ABCDEF")
	require.NoError(t, err)
}

func TestClient_GetAPIVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services/data/", r.URL.Path)

		resp := []APIVersion{
			{Label: "Winter '24", URL: "/services/data/v59.0", Version: "59.0"},
			{Label: "Spring '24", URL: "/services/data/v60.0", Version: "60.0"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	versions, err := client.GetAPIVersions(context.Background())
	require.NoError(t, err)

	assert.Len(t, versions, 2)
	assert.Equal(t, "59.0", versions[0].Version)
}

func TestClient_GetSObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services/data/v62.0/sobjects/", r.URL.Path)

		resp := SObjectsResponse{
			Encoding:     "UTF-8",
			MaxBatchSize: 200,
			SObjects: []SObjectDescribe{
				{Name: "Account", Label: "Account", Queryable: true},
				{Name: "Contact", Label: "Contact", Queryable: true},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	resp, err := client.GetSObjects(context.Background())
	require.NoError(t, err)

	assert.Len(t, resp.SObjects, 2)
	assert.Equal(t, "Account", resp.SObjects[0].Name)
}

func TestClient_DescribeSObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services/data/v62.0/sobjects/Account/describe", r.URL.Path)

		resp := SObjectDescribe{
			Name:       "Account",
			Label:      "Account",
			Createable: true,
			Updateable: true,
			Fields: []Field{
				{Name: "Id", Label: "Account ID", Type: "id"},
				{Name: "Name", Label: "Account Name", Type: "string"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	desc, err := client.DescribeSObject(context.Background(), "Account")
	require.NoError(t, err)

	assert.Equal(t, "Account", desc.Name)
	assert.Len(t, desc.Fields, 2)
}

func TestClient_GetLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services/data/v62.0/limits/", r.URL.Path)

		resp := Limits{
			"DailyApiRequests":     LimitInfo{Max: 15000, Remaining: 14500},
			"DailyBulkApiRequests": LimitInfo{Max: 5000, Remaining: 4999},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	limits, err := client.GetLimits(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 15000, limits["DailyApiRequests"].Max)
	assert.Equal(t, 14500, limits["DailyApiRequests"].Remaining)
}

func TestClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := []SalesforceError{
			{ErrorCode: "MALFORMED_QUERY", Message: "Invalid SOQL syntax"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	_, err = client.Query(context.Background(), "INVALID SOQL")
	require.Error(t, err)

	assert.True(t, IsBadRequest(err))
	assert.Contains(t, err.Error(), "MALFORMED_QUERY")
}

func TestClient_RecordURL(t *testing.T) {
	client := &Client{
		InstanceURL: "https://mycompany.my.salesforce.com",
	}

	url := client.RecordURL("001xx000003ABCDEF")
	assert.Equal(t, "https://mycompany.my.salesforce.com/001xx000003ABCDEF", url)
}
