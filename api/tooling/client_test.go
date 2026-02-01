package tooling

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
				HTTPClient:  http.DefaultClient,
			},
			wantErr: false,
		},
		{
			name: "missing instance URL",
			cfg: ClientConfig{
				HTTPClient: http.DefaultClient,
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
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestListApexClasses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/tooling/query")
		assert.Contains(t, r.URL.RawQuery, "ApexClass")

		response := QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []Record{
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

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	classes, err := client.ListApexClasses(context.Background())
	require.NoError(t, err)
	assert.Len(t, classes, 2)
	assert.Equal(t, "MyController", classes[0].Name)
	assert.Equal(t, "MyHelper", classes[1].Name)
}

func TestListApexTriggers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/tooling/query")
		assert.Contains(t, r.URL.RawQuery, "ApexTrigger")

		response := QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []Record{
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

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	triggers, err := client.ListApexTriggers(context.Background())
	require.NoError(t, err)
	assert.Len(t, triggers, 1)
	assert.Equal(t, "AccountTrigger", triggers[0].Name)
	assert.Equal(t, "Account", triggers[0].TableEnumOrID)
}

func TestGetApexClass(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []Record{
				{
					"Id":         "01p000000000001",
					"Name":       "MyController",
					"Body":       "public class MyController { }",
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

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	class, err := client.GetApexClass(context.Background(), "MyController")
	require.NoError(t, err)
	assert.Equal(t, "MyController", class.Name)
	assert.Equal(t, "public class MyController { }", class.Body)
}

func TestGetApexClassNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := QueryResult{
			TotalSize: 0,
			Done:      true,
			Records:   []Record{},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	_, err = client.GetApexClass(context.Background(), "NonExistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExecuteAnonymous(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/executeAnonymous")
		assert.Contains(t, r.URL.RawQuery, "anonymousBody")

		response := ExecuteAnonymousResult{
			Compiled: true,
			Success:  true,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	result, err := client.ExecuteAnonymous(context.Background(), "System.debug('Hello');")
	require.NoError(t, err)
	assert.True(t, result.Compiled)
	assert.True(t, result.Success)
}

func TestExecuteAnonymousCompileError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ExecuteAnonymousResult{
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

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	result, err := client.ExecuteAnonymous(context.Background(), "System.debug(foo);")
	require.NoError(t, err)
	assert.False(t, result.Compiled)
	assert.False(t, result.Success)
	assert.Equal(t, "Variable does not exist: foo", result.CompileProblem)
}

func TestRunTestsAsync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/runTestsAsynchronous")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`"7071x00000ABCDE"`))
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	jobID, err := client.RunTestsAsync(context.Background(), []string{"01p000000000001"})
	require.NoError(t, err)
	assert.Equal(t, "7071x00000ABCDE", jobID)
}

func TestGetTestResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []Record{
				{
					"Id":             "07M000000000001",
					"ApexClassId":    "01p000000000001",
					"ApexClass":      map[string]interface{}{"Name": "MyTest"},
					"MethodName":     "testSuccess",
					"Outcome":        "Pass",
					"RunTime":        float64(150),
					"AsyncApexJobId": "7071x00000ABCDE",
				},
				{
					"Id":             "07M000000000002",
					"ApexClassId":    "01p000000000001",
					"ApexClass":      map[string]interface{}{"Name": "MyTest"},
					"MethodName":     "testFailure",
					"Outcome":        "Fail",
					"Message":        "Assertion failed",
					"RunTime":        float64(200),
					"AsyncApexJobId": "7071x00000ABCDE",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	results, err := client.GetTestResults(context.Background(), "7071x00000ABCDE")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "testSuccess", results[0].MethodName)
	assert.Equal(t, "Pass", results[0].Outcome)
	assert.Equal(t, "testFailure", results[1].MethodName)
	assert.Equal(t, "Fail", results[1].Outcome)
}

func TestListApexLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []Record{
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

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	logs, err := client.ListApexLogs(context.Background(), "", 10)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "07L000000000001", logs[0].ID)
	assert.Equal(t, "/aura", logs[0].Operation)
	assert.Equal(t, 5000, logs[0].LogLength)
}

func TestGetApexLogBody(t *testing.T) {
	logContent := "DEBUG|Hello World\nUSER_DEBUG|Test message"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/sobjects/ApexLog/")
		assert.Contains(t, r.URL.Path, "/Body")

		w.Write([]byte(logContent))
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	body, err := client.GetApexLogBody(context.Background(), "07L000000000001")
	require.NoError(t, err)
	assert.Equal(t, logContent, body)
}

func TestGetCodeCoverage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []Record{
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

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	coverage, err := client.GetCodeCoverage(context.Background())
	require.NoError(t, err)
	assert.Len(t, coverage, 2)
	assert.Equal(t, "MyController", coverage[0].ApexClassOrTrigger.Name)
	assert.Equal(t, 80, coverage[0].NumLinesCovered)
	assert.Equal(t, 20, coverage[0].NumLinesUncovered)
}

func TestGetAsyncJobStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := QueryResult{
			TotalSize: 1,
			Done:      true,
			Records: []Record{
				{
					"Id":                "7071x00000ABCDE",
					"Status":            "Completed",
					"JobItemsProcessed": float64(5),
					"TotalJobItems":     float64(5),
					"NumberOfErrors":    float64(1),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	job, err := client.GetAsyncJobStatus(context.Background(), "7071x00000ABCDE")
	require.NoError(t, err)
	assert.Equal(t, "Completed", job.Status)
	assert.Equal(t, 5, job.TotalJobItems)
	assert.Equal(t, 1, job.NumberOfErrors)
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`[{"errorCode": "INVALID_SESSION_ID", "message": "Session expired"}]`))
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	_, err = client.ListApexClasses(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}
