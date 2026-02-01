package metadata

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// DefaultAPIVersion is the default Salesforce API version.
const DefaultAPIVersion = "v62.0"

// Client is a Salesforce Metadata API client.
type Client struct {
	httpClient  *http.Client
	instanceURL string
	apiVersion  string
	baseURL     string
}

// ClientConfig contains configuration for creating a new Metadata API client.
type ClientConfig struct {
	InstanceURL string
	HTTPClient  *http.Client
	APIVersion  string
}

// New creates a new Metadata API client.
func New(cfg ClientConfig) (*Client, error) {
	if cfg.InstanceURL == "" {
		return nil, fmt.Errorf("instance URL is required")
	}
	if cfg.HTTPClient == nil {
		return nil, fmt.Errorf("HTTP client is required")
	}

	instanceURL := strings.TrimSuffix(cfg.InstanceURL, "/")
	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	return &Client{
		httpClient:  cfg.HTTPClient,
		instanceURL: instanceURL,
		apiVersion:  apiVersion,
		baseURL:     fmt.Sprintf("%s/services/data/%s", instanceURL, apiVersion),
	}, nil
}

// doRequest performs an HTTP request and returns the response body.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	fullURL := path
	if !strings.HasPrefix(path, "http") {
		fullURL = c.baseURL + path
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// DescribeMetadata returns available metadata types.
func (c *Client) DescribeMetadata(ctx context.Context) (*DescribeMetadataResult, error) {
	path := "/tooling/describe"
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var describeResult struct {
		Sobjects []struct {
			Name       string `json:"name"`
			Createable bool   `json:"createable"`
			Updateable bool   `json:"updateable"`
			Deletable  bool   `json:"deletable"`
			Queryable  bool   `json:"queryable"`
			KeyPrefix  string `json:"keyPrefix"`
		} `json:"sobjects"`
	}

	if err := json.Unmarshal(body, &describeResult); err != nil {
		return nil, fmt.Errorf("failed to parse describe result: %w", err)
	}

	metadataTypeNames := map[string]bool{
		"ApexClass":                true,
		"ApexTrigger":              true,
		"ApexComponent":            true,
		"ApexPage":                 true,
		"AuraDefinition":           true,
		"LightningComponentBundle": true,
		"StaticResource":           true,
		"CustomObject":             true,
		"CustomField":              true,
		"ValidationRule":           true,
		"WorkflowRule":             true,
		"Flow":                     true,
		"FlowDefinition":           true,
	}

	result := &DescribeMetadataResult{
		MetadataObjects: make([]MetadataType, 0),
	}

	for _, obj := range describeResult.Sobjects {
		if metadataTypeNames[obj.Name] {
			result.MetadataObjects = append(result.MetadataObjects, MetadataType{
				XMLName: obj.Name,
			})
		}
	}

	return result, nil
}

// ListMetadata lists components of a specific metadata type.
func (c *Client) ListMetadata(ctx context.Context, metadataType string) ([]MetadataComponent, error) {
	var soql string
	switch metadataType {
	case "ApexClass":
		soql = "SELECT Id, Name, NamespacePrefix, CreatedById, LastModifiedById, LastModifiedDate FROM ApexClass ORDER BY Name"
	case "ApexTrigger":
		soql = "SELECT Id, Name, NamespacePrefix, TableEnumOrId, CreatedById, LastModifiedById, LastModifiedDate FROM ApexTrigger ORDER BY Name"
	case "ApexPage":
		soql = "SELECT Id, Name, NamespacePrefix, MasterLabel, CreatedById, LastModifiedById, LastModifiedDate FROM ApexPage ORDER BY Name"
	case "ApexComponent":
		soql = "SELECT Id, Name, NamespacePrefix, MasterLabel, CreatedById, LastModifiedById, LastModifiedDate FROM ApexComponent ORDER BY Name"
	case "StaticResource":
		soql = "SELECT Id, Name, NamespacePrefix, ContentType, CreatedById, LastModifiedById, LastModifiedDate FROM StaticResource ORDER BY Name"
	case "AuraDefinitionBundle":
		soql = "SELECT Id, DeveloperName, NamespacePrefix, MasterLabel, CreatedById, LastModifiedById, LastModifiedDate FROM AuraDefinitionBundle ORDER BY DeveloperName"
	case "LightningComponentBundle":
		soql = "SELECT Id, DeveloperName, NamespacePrefix, MasterLabel FROM LightningComponentBundle ORDER BY DeveloperName"
	default:
		return nil, fmt.Errorf("unsupported metadata type: %s", metadataType)
	}

	path := fmt.Sprintf("/tooling/query?q=%s", url.QueryEscape(soql))
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var queryResult struct {
		TotalSize int                      `json:"totalSize"`
		Done      bool                     `json:"done"`
		Records   []map[string]interface{} `json:"records"`
	}

	if err := json.Unmarshal(body, &queryResult); err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	components := make([]MetadataComponent, 0, len(queryResult.Records))
	for _, rec := range queryResult.Records {
		comp := MetadataComponent{
			Type: metadataType,
		}

		if id, ok := rec["Id"].(string); ok {
			comp.ID = id
		}

		if name, ok := rec["Name"].(string); ok {
			comp.FullName = name
		} else if name, ok := rec["DeveloperName"].(string); ok {
			comp.FullName = name
		}

		if ns, ok := rec["NamespacePrefix"].(string); ok {
			comp.NamespacePrefix = ns
		}

		components = append(components, comp)
	}

	return components, nil
}

// Deploy deploys metadata to the org.
func (c *Client) Deploy(ctx context.Context, zipData []byte, options DeployOptions) (*DeployResult, error) {
	zipBase64 := base64.StdEncoding.EncodeToString(zipData)

	request := DeployRequest{
		ZipFile:       zipBase64,
		DeployOptions: options,
	}

	body, err := c.Post(ctx, "/metadata/deployRequest", request)
	if err != nil {
		return nil, err
	}

	var result DeployResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse deploy result: %w", err)
	}

	return &result, nil
}

// GetDeployStatus gets the status of a deployment.
func (c *Client) GetDeployStatus(ctx context.Context, deployID string, includeDetails bool) (*DeployResult, error) {
	path := fmt.Sprintf("/metadata/deployRequest/%s", deployID)
	if includeDetails {
		path += "?includeDetails=true"
	}

	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result DeployResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse deploy status: %w", err)
	}

	return &result, nil
}

// CreateZipFromDirectory creates a zip file from a directory.
func CreateZipFromDirectory(sourceDir string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		zipPath := filepath.ToSlash(relPath)

		if info.IsDir() {
			_, err := zipWriter.Create(zipPath + "/")
			return err
		}

		writer, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExtractZipToDirectory extracts a zip file to a directory.
func ExtractZipToDirectory(zipData []byte, destDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to read zip: %w", err)
	}

	for _, file := range reader.File {
		destPath := filepath.Join(destDir, file.Name)

		// Prevent zip slip vulnerability
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		if err := extractFile(file, destPath); err != nil {
			return err
		}
	}

	return nil
}

// extractFile extracts a single file from a zip archive.
func extractFile(file *zip.File, destPath string) error {
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// Retrieve retrieves metadata from the org using the Tooling API.
// For complex retrieves with package.xml, use the official Salesforce CLI.
func (c *Client) Retrieve(ctx context.Context, metadataType, componentName string) ([]byte, error) {
	var soql string
	switch metadataType {
	case "ApexClass":
		soql = fmt.Sprintf("SELECT Id, Name, Body FROM ApexClass WHERE Name = '%s'", componentName)
	case "ApexTrigger":
		soql = fmt.Sprintf("SELECT Id, Name, Body FROM ApexTrigger WHERE Name = '%s'", componentName)
	case "ApexPage":
		soql = fmt.Sprintf("SELECT Id, Name, Markup FROM ApexPage WHERE Name = '%s'", componentName)
	case "ApexComponent":
		soql = fmt.Sprintf("SELECT Id, Name, Markup FROM ApexComponent WHERE Name = '%s'", componentName)
	default:
		return nil, fmt.Errorf("direct retrieve not supported for type: %s (use sf CLI for complex retrieves)", metadataType)
	}

	path := fmt.Sprintf("/tooling/query?q=%s", url.QueryEscape(soql))
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var queryResult struct {
		Records []map[string]interface{} `json:"records"`
	}

	if err := json.Unmarshal(body, &queryResult); err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	if len(queryResult.Records) == 0 {
		return nil, fmt.Errorf("%s not found: %s", metadataType, componentName)
	}

	rec := queryResult.Records[0]
	var content string
	if body, ok := rec["Body"].(string); ok {
		content = body
	} else if markup, ok := rec["Markup"].(string); ok {
		content = markup
	} else {
		return nil, fmt.Errorf("no content found for %s: %s", metadataType, componentName)
	}

	return []byte(content), nil
}

// RetrieveAll retrieves all components of a type from the org.
func (c *Client) RetrieveAll(ctx context.Context, metadataType string) (map[string][]byte, error) {
	components, err := c.ListMetadata(ctx, metadataType)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]byte)
	for _, comp := range components {
		if comp.NamespacePrefix != "" {
			continue
		}

		content, err := c.Retrieve(ctx, metadataType, comp.FullName)
		if err != nil {
			continue
		}
		results[comp.FullName] = content
	}

	return results, nil
}
