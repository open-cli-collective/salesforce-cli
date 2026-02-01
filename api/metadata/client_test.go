package metadata

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestDescribeMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/tooling/describe")

		response := struct {
			Sobjects []struct {
				Name       string `json:"name"`
				Createable bool   `json:"createable"`
				Updateable bool   `json:"updateable"`
				Deletable  bool   `json:"deletable"`
				Queryable  bool   `json:"queryable"`
			} `json:"sobjects"`
		}{
			Sobjects: []struct {
				Name       string `json:"name"`
				Createable bool   `json:"createable"`
				Updateable bool   `json:"updateable"`
				Deletable  bool   `json:"deletable"`
				Queryable  bool   `json:"queryable"`
			}{
				{Name: "ApexClass", Createable: true, Updateable: true, Deletable: true, Queryable: true},
				{Name: "ApexTrigger", Createable: true, Updateable: true, Deletable: true, Queryable: true},
				{Name: "SomeOtherObject", Createable: true, Updateable: true, Deletable: true, Queryable: true},
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

	result, err := client.DescribeMetadata(context.Background())
	require.NoError(t, err)

	// Should only include known metadata types
	assert.GreaterOrEqual(t, len(result.MetadataObjects), 2)

	found := false
	for _, obj := range result.MetadataObjects {
		if obj.XMLName == "ApexClass" {
			found = true
			break
		}
	}
	assert.True(t, found, "ApexClass should be in metadata types")
}

func TestListMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/tooling/query")
		assert.Contains(t, r.URL.RawQuery, "ApexClass")

		response := struct {
			TotalSize int                      `json:"totalSize"`
			Done      bool                     `json:"done"`
			Records   []map[string]interface{} `json:"records"`
		}{
			TotalSize: 2,
			Done:      true,
			Records: []map[string]interface{}{
				{"Id": "01p000000000001", "Name": "MyController", "NamespacePrefix": nil},
				{"Id": "01p000000000002", "Name": "MyHelper", "NamespacePrefix": nil},
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

	components, err := client.ListMetadata(context.Background(), "ApexClass")
	require.NoError(t, err)

	assert.Len(t, components, 2)
	assert.Equal(t, "MyController", components[0].FullName)
	assert.Equal(t, "ApexClass", components[0].Type)
}

func TestListMetadataUnsupportedType(t *testing.T) {
	client, err := New(ClientConfig{
		InstanceURL: "https://test.salesforce.com",
		HTTPClient:  http.DefaultClient,
	})
	require.NoError(t, err)

	_, err = client.ListMetadata(context.Background(), "UnsupportedType")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestRetrieve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Records []map[string]interface{} `json:"records"`
		}{
			Records: []map[string]interface{}{
				{
					"Id":   "01p000000000001",
					"Name": "MyController",
					"Body": "public class MyController { }",
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

	content, err := client.Retrieve(context.Background(), "ApexClass", "MyController")
	require.NoError(t, err)

	assert.Equal(t, "public class MyController { }", string(content))
}

func TestRetrieveNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Records []map[string]interface{} `json:"records"`
		}{
			Records: []map[string]interface{}{},
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

	_, err = client.Retrieve(context.Background(), "ApexClass", "NonExistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeploy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/metadata/deployRequest")
		assert.Equal(t, http.MethodPost, r.Method)

		var request DeployRequest
		_ = json.NewDecoder(r.Body).Decode(&request)
		assert.NotEmpty(t, request.ZipFile)

		result := DeployResult{
			ID:     "0Af000000000001",
			Status: "Pending",
			Done:   false,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	// Create a simple zip
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	writer, _ := zipWriter.Create("test.txt")
	_, _ = writer.Write([]byte("test content"))
	_ = zipWriter.Close()

	result, err := client.Deploy(context.Background(), buf.Bytes(), DeployOptions{CheckOnly: true})
	require.NoError(t, err)

	assert.Equal(t, "0Af000000000001", result.ID)
	assert.Equal(t, "Pending", result.Status)
}

func TestGetDeployStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/metadata/deployRequest/0Af000000000001")

		result := DeployResult{
			ID:                       "0Af000000000001",
			Status:                   "Succeeded",
			Done:                     true,
			Success:                  true,
			NumberComponentsTotal:    5,
			NumberComponentsDeployed: 5,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client, err := New(ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	result, err := client.GetDeployStatus(context.Background(), "0Af000000000001", false)
	require.NoError(t, err)

	assert.Equal(t, "Succeeded", result.Status)
	assert.True(t, result.Done)
	assert.True(t, result.Success)
	assert.Equal(t, 5, result.NumberComponentsDeployed)
}

func TestCreateZipFromDirectory(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "classes")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	testFile := filepath.Join(subDir, "MyClass.cls")
	require.NoError(t, os.WriteFile(testFile, []byte("public class MyClass {}"), 0644))

	metaFile := filepath.Join(subDir, "MyClass.cls-meta.xml")
	require.NoError(t, os.WriteFile(metaFile, []byte("<ApexClass/>"), 0644))

	zipData, err := CreateZipFromDirectory(tmpDir)
	require.NoError(t, err)
	assert.NotEmpty(t, zipData)

	// Verify zip contents
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	require.NoError(t, err)

	fileNames := make([]string, 0)
	for _, f := range reader.File {
		fileNames = append(fileNames, f.Name)
	}

	assert.Contains(t, fileNames, "classes/MyClass.cls")
	assert.Contains(t, fileNames, "classes/MyClass.cls-meta.xml")
}

func TestExtractZipToDirectory(t *testing.T) {
	// Create a test zip
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	writer, _ := zipWriter.Create("classes/MyClass.cls")
	_, _ = writer.Write([]byte("public class MyClass {}"))

	writer, _ = zipWriter.Create("classes/MyClass.cls-meta.xml")
	_, _ = writer.Write([]byte("<ApexClass/>"))

	require.NoError(t, zipWriter.Close())

	// Extract to temp directory
	destDir := t.TempDir()
	err := ExtractZipToDirectory(buf.Bytes(), destDir)
	require.NoError(t, err)

	// Verify extracted files
	content, err := os.ReadFile(filepath.Join(destDir, "classes", "MyClass.cls"))
	require.NoError(t, err)
	assert.Equal(t, "public class MyClass {}", string(content))

	content, err = os.ReadFile(filepath.Join(destDir, "classes", "MyClass.cls-meta.xml"))
	require.NoError(t, err)
	assert.Equal(t, "<ApexClass/>", string(content))
}

func TestExtractZipToDirectoryZipSlip(t *testing.T) {
	// Create a malicious zip with path traversal
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Try to escape the destination directory
	writer, _ := zipWriter.Create("../../../etc/passwd")
	_, _ = writer.Write([]byte("malicious"))

	require.NoError(t, zipWriter.Close())

	destDir := t.TempDir()
	err := ExtractZipToDirectory(buf.Bytes(), destDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "illegal file path")
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

	_, err = client.DescribeMetadata(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}
