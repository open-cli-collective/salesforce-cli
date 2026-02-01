package objectcmd

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

func TestListCommand(t *testing.T) {
	sobjectsResp := api.SObjectsResponse{
		Encoding:     "UTF-8",
		MaxBatchSize: 200,
		SObjects: []api.SObjectDescribe{
			{Name: "Account", Label: "Account", LabelPlural: "Accounts", KeyPrefix: "001", Queryable: true},
			{Name: "MyCustom__c", Label: "My Custom", LabelPlural: "My Customs", KeyPrefix: "a00", Custom: true, Queryable: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sobjectsResp)
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	t.Run("list all objects", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		opts := &root.Options{
			Output: "table",
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}
		opts.SetAPIClient(client)

		cmd := newListCommand(opts)
		cmd.SetOut(stdout)
		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Account")
		assert.Contains(t, output, "MyCustom__c")
	})

	t.Run("list custom only", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		opts := &root.Options{
			Output: "table",
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}
		opts.SetAPIClient(client)

		cmd := newListCommand(opts)
		cmd.SetArgs([]string{"--custom-only"})
		cmd.SetOut(stdout)
		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.NotContains(t, output, "Account")
		assert.Contains(t, output, "MyCustom__c")
	})
}

func TestDescribeCommand(t *testing.T) {
	describe := api.SObjectDescribe{
		Name:        "Account",
		Label:       "Account",
		LabelPlural: "Accounts",
		KeyPrefix:   "001",
		Createable:  true,
		Updateable:  true,
		Deletable:   true,
		Queryable:   true,
		Searchable:  true,
		Fields: []api.Field{
			{Name: "Id", Label: "Account ID", Type: "id"},
			{Name: "Name", Label: "Account Name", Type: "string", Length: 255},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/sobjects/Account/describe")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(describe)
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

	cmd := newDescribeCommand(opts)
	cmd.SetArgs([]string{"Account"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Account")
	assert.Contains(t, output, "Createable: true")
	assert.Contains(t, output, "Fields: 2")
}

func TestFieldsCommand(t *testing.T) {
	describe := api.SObjectDescribe{
		Name: "Account",
		Fields: []api.Field{
			{Name: "Id", Label: "Account ID", Type: "id", Nillable: false, Createable: false},
			{Name: "Name", Label: "Account Name", Type: "string", Length: 255, Nillable: false, Createable: true},
			{Name: "Description", Label: "Description", Type: "textarea", Nillable: true, Createable: true},
			{Name: "CustomField__c", Label: "Custom Field", Type: "string", Custom: true, Nillable: true, Createable: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(describe)
	}))
	defer server.Close()

	client, err := api.New(api.ClientConfig{
		InstanceURL: server.URL,
		HTTPClient:  server.Client(),
	})
	require.NoError(t, err)

	t.Run("list all fields", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		opts := &root.Options{
			Output: "table",
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}
		opts.SetAPIClient(client)

		cmd := newFieldsCommand(opts)
		cmd.SetArgs([]string{"Account"})
		cmd.SetOut(stdout)

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Id")
		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "Description")
		assert.Contains(t, output, "CustomField__c")
		assert.Contains(t, output, "4 field")
	})

	t.Run("list required only", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		opts := &root.Options{
			Output: "table",
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}
		opts.SetAPIClient(client)

		cmd := newFieldsCommand(opts)
		cmd.SetArgs([]string{"Account", "--required-only"})
		cmd.SetOut(stdout)

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Name")
		assert.NotContains(t, output, "Description")
		assert.Contains(t, output, "1 field")
	})
}

func TestFieldsCommand_JSONOutput(t *testing.T) {
	describe := api.SObjectDescribe{
		Name: "Account",
		Fields: []api.Field{
			{Name: "Id", Label: "Account ID", Type: "id"},
			{Name: "Name", Label: "Account Name", Type: "string", Length: 255},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(describe)
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

	cmd := newFieldsCommand(opts)
	cmd.SetArgs([]string{"Account"})
	cmd.SetOut(stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	var fields []api.Field
	err = json.Unmarshal(stdout.Bytes(), &fields)
	require.NoError(t, err)
	assert.Len(t, fields, 2)
}
