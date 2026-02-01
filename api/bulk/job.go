package bulk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CreateJob creates a new bulk ingest job.
func (c *Client) CreateJob(ctx context.Context, cfg JobConfig) (*JobInfo, error) {
	contentType := cfg.ContentType
	if contentType == "" {
		contentType = ContentTypeCSV
	}

	req := CreateJobRequest{
		Object:              cfg.Object,
		Operation:           cfg.Operation,
		ExternalIDFieldName: cfg.ExternalID,
		ContentType:         contentType,
	}

	body, err := c.doRequest(ctx, http.MethodPost, "/jobs/ingest", req)
	if err != nil {
		return nil, err
	}

	var job JobInfo
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}

	return &job, nil
}

// UploadJobData uploads CSV data to a bulk job.
func (c *Client) UploadJobData(ctx context.Context, jobID string, data []byte) error {
	path := fmt.Sprintf("/jobs/ingest/%s/batches", jobID)
	_, err := c.doRequest(ctx, http.MethodPut, path, data)
	return err
}

// CloseJob marks a job as UploadComplete to start processing.
func (c *Client) CloseJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/jobs/ingest/%s", jobID)
	req := UpdateJobRequest{State: StateUploadComplete}

	body, err := c.doRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}

	var job JobInfo
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}

	return &job, nil
}

// GetJob retrieves information about a bulk ingest job.
func (c *Client) GetJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/jobs/ingest/%s", jobID)
	body, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var job JobInfo
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}

	return &job, nil
}

// AbortJob aborts a bulk job.
func (c *Client) AbortJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/jobs/ingest/%s", jobID)
	req := UpdateJobRequest{State: StateAborted}

	body, err := c.doRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}

	var job JobInfo
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}

	return &job, nil
}

// DeleteJob deletes a bulk job.
func (c *Client) DeleteJob(ctx context.Context, jobID string) error {
	path := fmt.Sprintf("/jobs/ingest/%s", jobID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// ListJobs lists bulk ingest jobs.
func (c *Client) ListJobs(ctx context.Context) (*JobsResponse, error) {
	body, err := c.doRequest(ctx, http.MethodGet, "/jobs/ingest", nil)
	if err != nil {
		return nil, err
	}

	var resp JobsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse jobs response: %w", err)
	}

	return &resp, nil
}

// GetSuccessfulResults retrieves successful results from a completed job.
func (c *Client) GetSuccessfulResults(ctx context.Context, jobID string) ([]byte, error) {
	path := fmt.Sprintf("/jobs/ingest/%s/successfulResults", jobID)
	return c.doCSVRequest(ctx, http.MethodGet, path)
}

// GetFailedResults retrieves failed results from a completed job.
func (c *Client) GetFailedResults(ctx context.Context, jobID string) ([]byte, error) {
	path := fmt.Sprintf("/jobs/ingest/%s/failedResults", jobID)
	return c.doCSVRequest(ctx, http.MethodGet, path)
}

// GetUnprocessedRecords retrieves unprocessed records from a job.
func (c *Client) GetUnprocessedRecords(ctx context.Context, jobID string) ([]byte, error) {
	path := fmt.Sprintf("/jobs/ingest/%s/unprocessedrecords", jobID)
	return c.doCSVRequest(ctx, http.MethodGet, path)
}

// PollJob polls a job until it reaches a terminal state or timeout.
func (c *Client) PollJob(ctx context.Context, jobID string, cfg PollConfig) (*JobInfo, error) {
	if cfg.Interval == 0 {
		cfg = DefaultPollConfig()
	}

	deadline := time.Now().Add(cfg.Timeout)
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for job to complete")
			}

			job, err := c.GetJob(ctx, jobID)
			if err != nil {
				return nil, err
			}

			switch job.State {
			case StateJobComplete, StateFailed, StateAborted:
				return job, nil
			}
		}
	}
}

// CreateQueryJob creates a new bulk query job.
func (c *Client) CreateQueryJob(ctx context.Context, cfg QueryConfig) (*QueryJobInfo, error) {
	contentType := cfg.ContentType
	if contentType == "" {
		contentType = ContentTypeCSV
	}

	req := CreateQueryJobRequest{
		Operation:   OperationQuery,
		Query:       cfg.Query,
		ContentType: contentType,
	}

	body, err := c.doRequest(ctx, http.MethodPost, "/jobs/query", req)
	if err != nil {
		return nil, err
	}

	var job QueryJobInfo
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse query job response: %w", err)
	}

	return &job, nil
}

// GetQueryJob retrieves information about a bulk query job.
func (c *Client) GetQueryJob(ctx context.Context, jobID string) (*QueryJobInfo, error) {
	path := fmt.Sprintf("/jobs/query/%s", jobID)
	body, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var job QueryJobInfo
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse query job response: %w", err)
	}

	return &job, nil
}

// GetQueryResults retrieves results from a bulk query job.
func (c *Client) GetQueryResults(ctx context.Context, jobID string) ([]byte, error) {
	path := fmt.Sprintf("/jobs/query/%s/results", jobID)
	return c.doCSVRequest(ctx, http.MethodGet, path)
}

// AbortQueryJob aborts a bulk query job.
func (c *Client) AbortQueryJob(ctx context.Context, jobID string) (*QueryJobInfo, error) {
	path := fmt.Sprintf("/jobs/query/%s", jobID)
	req := UpdateJobRequest{State: StateAborted}

	body, err := c.doRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}

	var job QueryJobInfo
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse query job response: %w", err)
	}

	return &job, nil
}

// DeleteQueryJob deletes a bulk query job.
func (c *Client) DeleteQueryJob(ctx context.Context, jobID string) error {
	path := fmt.Sprintf("/jobs/query/%s", jobID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// ListQueryJobs lists bulk query jobs.
func (c *Client) ListQueryJobs(ctx context.Context) (*QueryJobsResponse, error) {
	body, err := c.doRequest(ctx, http.MethodGet, "/jobs/query", nil)
	if err != nil {
		return nil, err
	}

	var resp QueryJobsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse query jobs response: %w", err)
	}

	return &resp, nil
}

// PollQueryJob polls a query job until it reaches a terminal state or timeout.
func (c *Client) PollQueryJob(ctx context.Context, jobID string, cfg PollConfig) (*QueryJobInfo, error) {
	if cfg.Interval == 0 {
		cfg = DefaultPollConfig()
	}

	deadline := time.Now().Add(cfg.Timeout)
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for query job to complete")
			}

			job, err := c.GetQueryJob(ctx, jobID)
			if err != nil {
				return nil, err
			}

			switch job.State {
			case StateJobComplete, StateFailed, StateAborted:
				return job, nil
			}
		}
	}
}
