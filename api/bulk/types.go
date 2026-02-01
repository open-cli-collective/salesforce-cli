// Package bulk provides a client for the Salesforce Bulk API 2.0.
package bulk

import "time"

// Operation represents a bulk job operation type.
type Operation string

// Bulk job operations.
const (
	OperationInsert Operation = "insert"
	OperationUpdate Operation = "update"
	OperationUpsert Operation = "upsert"
	OperationDelete Operation = "delete"
	OperationQuery  Operation = "query"
)

// State represents a bulk job state.
type State string

// Bulk job states.
const (
	StateOpen           State = "Open"
	StateUploadComplete State = "UploadComplete"
	StateInProgress     State = "InProgress"
	StateJobComplete    State = "JobComplete"
	StateFailed         State = "Failed"
	StateAborted        State = "Aborted"
)

// ContentType represents the content type for bulk data.
type ContentType string

// Content types.
const (
	ContentTypeCSV  ContentType = "CSV"
	ContentTypeJSON ContentType = "JSON"
)

// JobInfo represents information about a bulk job.
type JobInfo struct {
	ID                      string      `json:"id,omitempty"`
	Operation               Operation   `json:"operation"`
	Object                  string      `json:"object"`
	CreatedByID             string      `json:"createdById,omitempty"`
	CreatedDate             string      `json:"createdDate,omitempty"`
	SystemModstamp          string      `json:"systemModstamp,omitempty"`
	State                   State       `json:"state,omitempty"`
	ExternalIDFieldName     string      `json:"externalIdFieldName,omitempty"`
	ConcurrencyMode         string      `json:"concurrencyMode,omitempty"`
	ContentType             ContentType `json:"contentType,omitempty"`
	APIVersion              float64     `json:"apiVersion,omitempty"`
	JobType                 string      `json:"jobType,omitempty"`
	LineEnding              string      `json:"lineEnding,omitempty"`
	ColumnDelimiter         string      `json:"columnDelimiter,omitempty"`
	NumberRecordsProcessed  int         `json:"numberRecordsProcessed,omitempty"`
	NumberRecordsFailed     int         `json:"numberRecordsFailed,omitempty"`
	Retries                 int         `json:"retries,omitempty"`
	TotalProcessingTime     int         `json:"totalProcessingTime,omitempty"`
	APIActiveProcessingTime int         `json:"apiActiveProcessingTime,omitempty"`
	ApexProcessingTime      int         `json:"apexProcessingTime,omitempty"`
	ErrorMessage            string      `json:"errorMessage,omitempty"`
}

// QueryJobInfo represents information about a bulk query job.
type QueryJobInfo struct {
	ID                     string      `json:"id,omitempty"`
	Operation              Operation   `json:"operation"`
	Object                 string      `json:"object,omitempty"`
	CreatedByID            string      `json:"createdById,omitempty"`
	CreatedDate            string      `json:"createdDate,omitempty"`
	SystemModstamp         string      `json:"systemModstamp,omitempty"`
	State                  State       `json:"state,omitempty"`
	ConcurrencyMode        string      `json:"concurrencyMode,omitempty"`
	ContentType            ContentType `json:"contentType,omitempty"`
	APIVersion             float64     `json:"apiVersion,omitempty"`
	LineEnding             string      `json:"lineEnding,omitempty"`
	ColumnDelimiter        string      `json:"columnDelimiter,omitempty"`
	NumberRecordsProcessed int         `json:"numberRecordsProcessed,omitempty"`
	Retries                int         `json:"retries,omitempty"`
	TotalProcessingTime    int         `json:"totalProcessingTime,omitempty"`
	Query                  string      `json:"query,omitempty"`
}

// JobsResponse represents a list of bulk jobs.
type JobsResponse struct {
	Done           bool      `json:"done"`
	Records        []JobInfo `json:"records"`
	NextRecordsURL string    `json:"nextRecordsUrl,omitempty"`
}

// QueryJobsResponse represents a list of bulk query jobs.
type QueryJobsResponse struct {
	Done           bool           `json:"done"`
	Records        []QueryJobInfo `json:"records"`
	NextRecordsURL string         `json:"nextRecordsUrl,omitempty"`
}

// CreateJobRequest represents a request to create a bulk ingest job.
type CreateJobRequest struct {
	Object              string      `json:"object"`
	Operation           Operation   `json:"operation"`
	ExternalIDFieldName string      `json:"externalIdFieldName,omitempty"`
	ContentType         ContentType `json:"contentType,omitempty"`
	LineEnding          string      `json:"lineEnding,omitempty"`
	ColumnDelimiter     string      `json:"columnDelimiter,omitempty"`
}

// CreateQueryJobRequest represents a request to create a bulk query job.
type CreateQueryJobRequest struct {
	Operation       Operation   `json:"operation"`
	Query           string      `json:"query"`
	ContentType     ContentType `json:"contentType,omitempty"`
	LineEnding      string      `json:"lineEnding,omitempty"`
	ColumnDelimiter string      `json:"columnDelimiter,omitempty"`
}

// UpdateJobRequest represents a request to update a bulk job state.
type UpdateJobRequest struct {
	State State `json:"state"`
}

// JobConfig contains configuration for creating a bulk job.
type JobConfig struct {
	Object      string
	Operation   Operation
	ExternalID  string
	ContentType ContentType
}

// QueryConfig contains configuration for creating a bulk query job.
type QueryConfig struct {
	Query       string
	ContentType ContentType
}

// PollConfig contains configuration for polling job status.
type PollConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

// DefaultPollConfig returns default polling configuration.
func DefaultPollConfig() PollConfig {
	return PollConfig{
		Interval: 5 * time.Second,
		Timeout:  10 * time.Minute,
	}
}
