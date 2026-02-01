package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DefaultAPIVersion is the default Salesforce API version
const DefaultAPIVersion = "v62.0"

// Client is a Salesforce REST API client
type Client struct {
	// HTTPClient is the underlying HTTP client (should have OAuth token)
	HTTPClient *http.Client

	// InstanceURL is the Salesforce instance URL (e.g., https://mycompany.my.salesforce.com)
	InstanceURL string

	// APIVersion is the API version to use (e.g., v62.0)
	APIVersion string

	// BaseURL is the full REST API base URL
	BaseURL string
}

// ClientConfig contains configuration for creating a new client
type ClientConfig struct {
	// InstanceURL is the Salesforce instance URL
	InstanceURL string

	// HTTPClient is an authenticated HTTP client (e.g., from auth.GetHTTPClient)
	HTTPClient *http.Client

	// APIVersion is the API version to use (optional, defaults to DefaultAPIVersion)
	APIVersion string
}

// New creates a new Salesforce API client
func New(cfg ClientConfig) (*Client, error) {
	if cfg.InstanceURL == "" {
		return nil, ErrInstanceURLRequired
	}
	if cfg.HTTPClient == nil {
		return nil, ErrHTTPClientRequired
	}

	instanceURL := normalizeURL(cfg.InstanceURL)

	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	return &Client{
		HTTPClient:  cfg.HTTPClient,
		InstanceURL: instanceURL,
		APIVersion:  apiVersion,
		BaseURL:     fmt.Sprintf("%s/services/data/%s", instanceURL, apiVersion),
	}, nil
}

// normalizeURL ensures the URL has proper format
func normalizeURL(urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	return strings.TrimSuffix(urlStr, "/")
}

// Get performs a GET request to the specified path
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request to the specified path
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// Patch performs a PATCH request to the specified path
func (c *Client) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body)
}

// Put performs a PUT request to the specified path
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPut, path, body)
}

// Delete performs a DELETE request to the specified path
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	fullURL := c.buildURL(path)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, ParseAPIError(resp)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, nil
}

func (c *Client) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	if strings.HasPrefix(path, "/services/") {
		return c.InstanceURL + path
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.BaseURL + path
}

// GetAPIVersions returns available API versions
func (c *Client) GetAPIVersions(ctx context.Context) ([]APIVersion, error) {
	body, err := c.doRequest(ctx, http.MethodGet, c.InstanceURL+"/services/data/", nil)
	if err != nil {
		return nil, err
	}

	var versions []APIVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse API versions: %w", err)
	}

	return versions, nil
}

// GetSObjects returns metadata about all SObjects in the org
func (c *Client) GetSObjects(ctx context.Context) (*SObjectsResponse, error) {
	body, err := c.Get(ctx, "/sobjects/")
	if err != nil {
		return nil, err
	}

	var resp SObjectsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse sobjects response: %w", err)
	}

	return &resp, nil
}

// DescribeSObject returns detailed metadata about an SObject type
func (c *Client) DescribeSObject(ctx context.Context, objectName string) (*SObjectDescribe, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/sobjects/%s/describe", objectName))
	if err != nil {
		return nil, err
	}

	var desc SObjectDescribe
	if err := json.Unmarshal(body, &desc); err != nil {
		return nil, fmt.Errorf("failed to parse describe response: %w", err)
	}

	return &desc, nil
}

// GetLimits returns the org's API limits
func (c *Client) GetLimits(ctx context.Context) (Limits, error) {
	body, err := c.Get(ctx, "/limits/")
	if err != nil {
		return nil, err
	}

	var limits Limits
	if err := json.Unmarshal(body, &limits); err != nil {
		return nil, fmt.Errorf("failed to parse limits response: %w", err)
	}

	return limits, nil
}

// Query executes a SOQL query and returns the results
func (c *Client) Query(ctx context.Context, soql string) (*QueryResult, error) {
	path := fmt.Sprintf("/query?q=%s", url.QueryEscape(soql))
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result QueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	return &result, nil
}

// QueryMore retrieves the next batch of query results
func (c *Client) QueryMore(ctx context.Context, nextRecordsURL string) (*QueryResult, error) {
	body, err := c.Get(ctx, nextRecordsURL)
	if err != nil {
		return nil, err
	}

	var result QueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	return &result, nil
}

// QueryAll executes a query and retrieves all results (handles pagination)
func (c *Client) QueryAll(ctx context.Context, soql string) (*QueryResult, error) {
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	for !result.Done && result.NextRecordsURL != "" {
		nextPage, err := c.QueryMore(ctx, result.NextRecordsURL)
		if err != nil {
			return nil, err
		}
		result.Records = append(result.Records, nextPage.Records...)
		result.Done = nextPage.Done
		result.NextRecordsURL = nextPage.NextRecordsURL
	}

	return result, nil
}

// GetRecord retrieves a single record by ID
func (c *Client) GetRecord(ctx context.Context, objectName, recordID string, fields []string) (*SObject, error) {
	path := fmt.Sprintf("/sobjects/%s/%s", objectName, recordID)
	if len(fields) > 0 {
		path += "?fields=" + strings.Join(fields, ",")
	}

	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var record SObject
	if err := json.Unmarshal(body, &record); err != nil {
		return nil, fmt.Errorf("failed to parse record: %w", err)
	}

	return &record, nil
}

// CreateRecord creates a new record and returns the result
func (c *Client) CreateRecord(ctx context.Context, objectName string, record map[string]interface{}) (*RecordResult, error) {
	path := fmt.Sprintf("/sobjects/%s/", objectName)
	body, err := c.Post(ctx, path, record)
	if err != nil {
		return nil, err
	}

	var result RecordResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create result: %w", err)
	}

	return &result, nil
}

// UpdateRecord updates an existing record
func (c *Client) UpdateRecord(ctx context.Context, objectName, recordID string, record map[string]interface{}) error {
	path := fmt.Sprintf("/sobjects/%s/%s", objectName, recordID)
	_, err := c.Patch(ctx, path, record)
	return err
}

// DeleteRecord deletes a record
func (c *Client) DeleteRecord(ctx context.Context, objectName, recordID string) error {
	path := fmt.Sprintf("/sobjects/%s/%s", objectName, recordID)
	_, err := c.Delete(ctx, path)
	return err
}

// RecordURL returns the web URL for a record
func (c *Client) RecordURL(recordID string) string {
	return fmt.Sprintf("%s/%s", c.InstanceURL, recordID)
}

// Search executes a SOSL search and returns the results
func (c *Client) Search(ctx context.Context, sosl string) (*SearchResult, error) {
	path := fmt.Sprintf("/search?q=%s", url.QueryEscape(sosl))
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search result: %w", err)
	}

	return &result, nil
}
