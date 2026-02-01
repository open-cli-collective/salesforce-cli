package tooling

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultAPIVersion is the default Salesforce API version.
const DefaultAPIVersion = "v62.0"

// Client is a Salesforce Tooling API client.
type Client struct {
	httpClient  *http.Client
	instanceURL string
	apiVersion  string
	baseURL     string
}

// ClientConfig contains configuration for creating a new Tooling API client.
type ClientConfig struct {
	InstanceURL string
	HTTPClient  *http.Client
	APIVersion  string
}

// New creates a new Tooling API client.
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
		baseURL:     fmt.Sprintf("%s/services/data/%s/tooling", instanceURL, apiVersion),
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

// Query executes a SOQL query against the Tooling API.
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

// ListApexClasses returns all Apex classes.
func (c *Client) ListApexClasses(ctx context.Context) ([]ApexClass, error) {
	soql := "SELECT Id, Name, Status, IsValid, ApiVersion, LengthWithoutComments, NamespacePrefix FROM ApexClass ORDER BY Name"
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	classes := make([]ApexClass, 0, len(result.Records))
	for _, rec := range result.Records {
		class := recordToApexClass(rec)
		classes = append(classes, class)
	}

	return classes, nil
}

// ListApexTriggers returns all Apex triggers.
func (c *Client) ListApexTriggers(ctx context.Context) ([]ApexTrigger, error) {
	soql := "SELECT Id, Name, Status, IsValid, ApiVersion, TableEnumOrId, NamespacePrefix FROM ApexTrigger ORDER BY Name"
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	triggers := make([]ApexTrigger, 0, len(result.Records))
	for _, rec := range result.Records {
		trigger := recordToApexTrigger(rec)
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

// GetApexClass returns an Apex class by name, including body.
func (c *Client) GetApexClass(ctx context.Context, name string) (*ApexClass, error) {
	soql := fmt.Sprintf("SELECT Id, Name, Body, Status, IsValid, ApiVersion, LengthWithoutComments, NamespacePrefix FROM ApexClass WHERE Name = '%s'", name)
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("apex class not found: %s", name)
	}

	class := recordToApexClass(result.Records[0])
	return &class, nil
}

// GetApexTrigger returns an Apex trigger by name, including body.
func (c *Client) GetApexTrigger(ctx context.Context, name string) (*ApexTrigger, error) {
	soql := fmt.Sprintf("SELECT Id, Name, Body, Status, IsValid, ApiVersion, TableEnumOrId, NamespacePrefix FROM ApexTrigger WHERE Name = '%s'", name)
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("apex trigger not found: %s", name)
	}

	trigger := recordToApexTrigger(result.Records[0])
	return &trigger, nil
}

// ExecuteAnonymous executes anonymous Apex code.
func (c *Client) ExecuteAnonymous(ctx context.Context, code string) (*ExecuteAnonymousResult, error) {
	path := fmt.Sprintf("/executeAnonymous?anonymousBody=%s", url.QueryEscape(code))
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result ExecuteAnonymousResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse execute result: %w", err)
	}

	return &result, nil
}

// RunTestsAsync enqueues Apex tests to run asynchronously.
func (c *Client) RunTestsAsync(ctx context.Context, classIDs []string) (string, error) {
	req := RunTestsRequest{
		ClassIDs: classIDs,
	}
	body, err := c.Post(ctx, "/runTestsAsynchronous", req)
	if err != nil {
		return "", err
	}

	// Response is a quoted string with the job ID
	var jobID string
	if err := json.Unmarshal(body, &jobID); err != nil {
		// Try unquoting directly
		jobID = strings.Trim(string(body), "\"")
	}

	return jobID, nil
}

// GetTestResults returns test results for a given async job.
func (c *Client) GetTestResults(ctx context.Context, asyncJobID string) ([]ApexTestResult, error) {
	soql := fmt.Sprintf(
		"SELECT Id, ApexClassId, ApexClass.Name, MethodName, Outcome, Message, StackTrace, RunTime, AsyncApexJobId FROM ApexTestResult WHERE AsyncApexJobId = '%s'",
		asyncJobID,
	)
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	results := make([]ApexTestResult, 0, len(result.Records))
	for _, rec := range result.Records {
		tr := recordToApexTestResult(rec)
		results = append(results, tr)
	}

	return results, nil
}

// GetAsyncJobStatus returns the status of an async Apex job.
func (c *Client) GetAsyncJobStatus(ctx context.Context, jobID string) (*AsyncApexJob, error) {
	soql := fmt.Sprintf(
		"SELECT Id, Status, JobItemsProcessed, TotalJobItems, NumberOfErrors, ExtendedStatus, CompletedDate FROM AsyncApexJob WHERE Id = '%s'",
		jobID,
	)
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("async job not found: %s", jobID)
	}

	job := recordToAsyncApexJob(result.Records[0])
	return &job, nil
}

// ListApexLogs returns debug logs.
func (c *Client) ListApexLogs(ctx context.Context, userID string, limit int) ([]ApexLog, error) {
	soql := "SELECT Id, LogUserId, Operation, Request, Status, LogLength, DurationMilliseconds, StartTime, Location, Application FROM ApexLog"
	if userID != "" {
		soql += fmt.Sprintf(" WHERE LogUserId = '%s'", userID)
	}
	soql += " ORDER BY StartTime DESC"
	if limit > 0 {
		soql += fmt.Sprintf(" LIMIT %d", limit)
	}

	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	logs := make([]ApexLog, 0, len(result.Records))
	for _, rec := range result.Records {
		log := recordToApexLog(rec)
		logs = append(logs, log)
	}

	return logs, nil
}

// GetApexLogBody returns the body content of a debug log.
func (c *Client) GetApexLogBody(ctx context.Context, logID string) (string, error) {
	// Log body is retrieved from the REST API, not Tooling API
	path := fmt.Sprintf("%s/services/data/%s/sobjects/ApexLog/%s/Body", c.instanceURL, c.apiVersion, logID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// GetCodeCoverage returns aggregate code coverage for the org.
func (c *Client) GetCodeCoverage(ctx context.Context) ([]ApexCodeCoverageAggregate, error) {
	soql := "SELECT Id, ApexClassOrTriggerId, ApexClassOrTrigger.Name, NumLinesCovered, NumLinesUncovered FROM ApexCodeCoverageAggregate ORDER BY ApexClassOrTrigger.Name"
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	coverage := make([]ApexCodeCoverageAggregate, 0, len(result.Records))
	for _, rec := range result.Records {
		cov := recordToApexCodeCoverageAggregate(rec)
		coverage = append(coverage, cov)
	}

	return coverage, nil
}

// GetCodeCoverageForClass returns aggregate code coverage for a specific class.
func (c *Client) GetCodeCoverageForClass(ctx context.Context, className string) (*ApexCodeCoverageAggregate, error) {
	soql := fmt.Sprintf(
		"SELECT Id, ApexClassOrTriggerId, ApexClassOrTrigger.Name, NumLinesCovered, NumLinesUncovered FROM ApexCodeCoverageAggregate WHERE ApexClassOrTrigger.Name = '%s'",
		className,
	)
	result, err := c.Query(ctx, soql)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("no coverage data found for: %s", className)
	}

	cov := recordToApexCodeCoverageAggregate(result.Records[0])
	return &cov, nil
}

// GetApexClassID returns the ID of an Apex class by name.
func (c *Client) GetApexClassID(ctx context.Context, className string) (string, error) {
	soql := fmt.Sprintf("SELECT Id FROM ApexClass WHERE Name = '%s'", className)
	result, err := c.Query(ctx, soql)
	if err != nil {
		return "", err
	}

	if len(result.Records) == 0 {
		return "", fmt.Errorf("apex class not found: %s", className)
	}

	id, _ := result.Records[0]["Id"].(string)
	return id, nil
}

// Helper functions to convert generic records to typed structs

func recordToApexClass(rec Record) ApexClass {
	class := ApexClass{}
	if v, ok := rec["Id"].(string); ok {
		class.ID = v
	}
	if v, ok := rec["Name"].(string); ok {
		class.Name = v
	}
	if v, ok := rec["Body"].(string); ok {
		class.Body = v
	}
	if v, ok := rec["Status"].(string); ok {
		class.Status = v
	}
	if v, ok := rec["IsValid"].(bool); ok {
		class.IsValid = v
	}
	if v, ok := rec["ApiVersion"].(float64); ok {
		class.APIVersion = v
	}
	if v, ok := rec["LengthWithoutComments"].(float64); ok {
		class.LengthWithoutComments = int(v)
	}
	if v, ok := rec["NamespacePrefix"].(string); ok {
		class.NamespacePrefix = v
	}
	return class
}

func recordToApexTrigger(rec Record) ApexTrigger {
	trigger := ApexTrigger{}
	if v, ok := rec["Id"].(string); ok {
		trigger.ID = v
	}
	if v, ok := rec["Name"].(string); ok {
		trigger.Name = v
	}
	if v, ok := rec["Body"].(string); ok {
		trigger.Body = v
	}
	if v, ok := rec["Status"].(string); ok {
		trigger.Status = v
	}
	if v, ok := rec["IsValid"].(bool); ok {
		trigger.IsValid = v
	}
	if v, ok := rec["ApiVersion"].(float64); ok {
		trigger.APIVersion = v
	}
	if v, ok := rec["TableEnumOrId"].(string); ok {
		trigger.TableEnumOrID = v
	}
	if v, ok := rec["NamespacePrefix"].(string); ok {
		trigger.NamespacePrefix = v
	}
	return trigger
}

func recordToApexLog(rec Record) ApexLog {
	log := ApexLog{}
	if v, ok := rec["Id"].(string); ok {
		log.ID = v
	}
	if v, ok := rec["LogUserId"].(string); ok {
		log.LogUserID = v
	}
	if v, ok := rec["Operation"].(string); ok {
		log.Operation = v
	}
	if v, ok := rec["Request"].(string); ok {
		log.Request = v
	}
	if v, ok := rec["Status"].(string); ok {
		log.Status = v
	}
	if v, ok := rec["LogLength"].(float64); ok {
		log.LogLength = int(v)
	}
	if v, ok := rec["DurationMilliseconds"].(float64); ok {
		log.DurationMS = int(v)
	}
	if v, ok := rec["StartTime"].(string); ok {
		log.StartTime, _ = parseTime(v)
	}
	if v, ok := rec["Location"].(string); ok {
		log.Location = v
	}
	if v, ok := rec["Application"].(string); ok {
		log.Application = v
	}
	return log
}

func recordToApexTestResult(rec Record) ApexTestResult {
	result := ApexTestResult{}
	if v, ok := rec["Id"].(string); ok {
		result.ID = v
	}
	if v, ok := rec["ApexClassId"].(string); ok {
		result.ApexClassID = v
	}
	if nested, ok := rec["ApexClass"].(map[string]interface{}); ok {
		if v, ok := nested["Name"].(string); ok {
			result.ClassName = v
		}
	}
	if v, ok := rec["MethodName"].(string); ok {
		result.MethodName = v
	}
	if v, ok := rec["Outcome"].(string); ok {
		result.Outcome = v
	}
	if v, ok := rec["Message"].(string); ok {
		result.Message = v
	}
	if v, ok := rec["StackTrace"].(string); ok {
		result.StackTrace = v
	}
	if v, ok := rec["RunTime"].(float64); ok {
		result.RunTime = int(v)
	}
	if v, ok := rec["AsyncApexJobId"].(string); ok {
		result.AsyncApexJobID = v
	}
	return result
}

func recordToAsyncApexJob(rec Record) AsyncApexJob {
	job := AsyncApexJob{}
	if v, ok := rec["Id"].(string); ok {
		job.ID = v
	}
	if v, ok := rec["Status"].(string); ok {
		job.Status = v
	}
	if v, ok := rec["JobItemsProcessed"].(float64); ok {
		job.JobItemsProcessed = int(v)
	}
	if v, ok := rec["TotalJobItems"].(float64); ok {
		job.TotalJobItems = int(v)
	}
	if v, ok := rec["NumberOfErrors"].(float64); ok {
		job.NumberOfErrors = int(v)
	}
	if v, ok := rec["ExtendedStatus"].(string); ok {
		job.ExtendedStatus = v
	}
	if v, ok := rec["CompletedDate"].(string); ok {
		job.CompletedDate = v
	}
	return job
}

func recordToApexCodeCoverageAggregate(rec Record) ApexCodeCoverageAggregate {
	cov := ApexCodeCoverageAggregate{}
	if v, ok := rec["Id"].(string); ok {
		cov.ID = v
	}
	if v, ok := rec["ApexClassOrTriggerId"].(string); ok {
		cov.ApexClassOrTriggerID = v
	}
	if nested, ok := rec["ApexClassOrTrigger"].(map[string]interface{}); ok {
		if v, ok := nested["Name"].(string); ok {
			cov.ApexClassOrTrigger.Name = v
		}
	}
	if v, ok := rec["NumLinesCovered"].(float64); ok {
		cov.NumLinesCovered = int(v)
	}
	if v, ok := rec["NumLinesUncovered"].(float64); ok {
		cov.NumLinesUncovered = int(v)
	}
	return cov
}

func parseTime(s string) (time.Time, error) {
	// Salesforce datetime format
	formats := []string{
		"2006-01-02T15:04:05.000+0000",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
