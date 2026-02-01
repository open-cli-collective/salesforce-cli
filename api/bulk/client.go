package bulk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is a Salesforce Bulk API 2.0 client.
type Client struct {
	httpClient  *http.Client
	instanceURL string
	apiVersion  string
	baseURL     string
}

// ClientConfig contains configuration for creating a new Bulk API client.
type ClientConfig struct {
	InstanceURL string
	HTTPClient  *http.Client
	APIVersion  string
}

// New creates a new Bulk API client.
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
		apiVersion = "v62.0"
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
	contentType := "application/json"

	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
			contentType = "text/csv"
		case []byte:
			bodyReader = bytes.NewReader(v)
			contentType = "text/csv"
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}
	}

	fullURL := path
	if !strings.HasPrefix(path, "http") {
		fullURL = c.baseURL + path
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
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

// doCSVRequest performs an HTTP request expecting CSV response.
func (c *Client) doCSVRequest(ctx context.Context, method, path string) ([]byte, error) {
	fullURL := path
	if !strings.HasPrefix(path, "http") {
		fullURL = c.baseURL + path
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/csv")

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
