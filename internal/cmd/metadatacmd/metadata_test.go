package metadatacmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-cli-collective/salesforce-cli/api/metadata"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func TestMetadataTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Sobjects []struct {
				Name string `json:"name"`
			} `json:"sobjects"`
		}{
			Sobjects: []struct {
				Name string `json:"name"`
			}{
				{Name: "ApexClass"},
				{Name: "ApexTrigger"},
				{Name: "CustomObject"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := metadata.New(metadata.ClientConfig{
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
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"types"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "ApexClass")
	assert.Contains(t, output, "ApexTrigger")
}

func TestMetadataTypesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Sobjects []struct {
				Name string `json:"name"`
			} `json:"sobjects"`
		}{
			Sobjects: []struct {
				Name string `json:"name"`
			}{
				{Name: "ApexClass"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := metadata.New(metadata.ClientConfig{
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
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"types"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	var result []metadata.MetadataType
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
}

func TestMetadataList(t *testing.T) {
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
				{"Id": "01p000000000001", "Name": "MyController"},
				{"Id": "01p000000000002", "Name": "MyHelper"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := metadata.New(metadata.ClientConfig{
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
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"list", "--type", "ApexClass"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "MyController")
	assert.Contains(t, output, "MyHelper")
	assert.Contains(t, output, "2 component(s)")
}

func TestMetadataListMissingType(t *testing.T) {
	client, err := metadata.New(metadata.ClientConfig{
		InstanceURL: "https://test.salesforce.com",
		HTTPClient:  http.DefaultClient,
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"list"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--type is required")
}

func TestMetadataRetrieve(t *testing.T) {
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

	client, err := metadata.New(metadata.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"retrieve", "--type", "ApexClass", "--name", "MyController", "--output", tmpDir})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(filepath.Join(tmpDir, "MyController.cls"))
	require.NoError(t, err)
	assert.Equal(t, "public class MyController { }", string(content))
}

func TestMetadataRetrieveMissingFlags(t *testing.T) {
	client, err := metadata.New(metadata.ClientConfig{
		InstanceURL: "https://test.salesforce.com",
		HTTPClient:  http.DefaultClient,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing type",
			args:    []string{"retrieve", "--output", "./src"},
			wantErr: "--type is required",
		},
		{
			name:    "missing output",
			args:    []string{"retrieve", "--type", "ApexClass"},
			wantErr: "--output is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			opts := &root.Options{
				Output: "table",
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
			}
			opts.SetMetadataClient(client)

			cmd := NewCommand(opts)
			cmd.SetArgs(tt.args)
			cmd.SetOut(stdout)

			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMetadataDeploy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/metadata/deployRequest") && r.Method == http.MethodPost {
			result := metadata.DeployResult{
				ID:     "0Af000000000001",
				Status: "Pending",
				Done:   false,
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)
		}
	}))
	defer server.Close()

	client, err := metadata.New(metadata.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	// Create temp source directory with a file
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "classes"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "classes", "MyClass.cls"),
		[]byte("public class MyClass {}"),
		0644,
	))

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"deploy", "--source", tmpDir})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "0Af000000000001")
	assert.Contains(t, output, "Deployment started")
}

func TestMetadataDeployCheckOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var request metadata.DeployRequest
			_ = json.NewDecoder(r.Body).Decode(&request)
			assert.True(t, request.DeployOptions.CheckOnly)

			result := metadata.DeployResult{
				ID:        "0Af000000000001",
				Status:    "Pending",
				Done:      false,
				CheckOnly: true,
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)
		}
	}))
	defer server.Close()

	client, err := metadata.New(metadata.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "classes"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "classes", "MyClass.cls"),
		[]byte("public class MyClass {}"),
		0644,
	))

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"deploy", "--source", tmpDir, "--check-only"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Validating")
}

func TestMetadataDeployMissingSource(t *testing.T) {
	client, err := metadata.New(metadata.ClientConfig{
		InstanceURL: "https://test.salesforce.com",
		HTTPClient:  http.DefaultClient,
	})
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	opts := &root.Options{
		Output: "table",
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	opts.SetMetadataClient(client)

	cmd := NewCommand(opts)
	cmd.SetArgs([]string{"deploy"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--source is required")
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		metadataType string
		want         string
	}{
		{"ApexClass", ".cls"},
		{"ApexTrigger", ".trigger"},
		{"ApexPage", ".page"},
		{"ApexComponent", ".component"},
		{"CustomObject", ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.metadataType, func(t *testing.T) {
			got := getFileExtension(tt.metadataType)
			assert.Equal(t, tt.want, got)
		})
	}
}
