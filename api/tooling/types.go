// Package tooling provides a client for the Salesforce Tooling API.
package tooling

import "time"

// ApexClass represents an Apex class in Salesforce.
type ApexClass struct {
	ID                    string  `json:"Id"`
	Name                  string  `json:"Name"`
	Body                  string  `json:"Body,omitempty"`
	Status                string  `json:"Status"`
	IsValid               bool    `json:"IsValid"`
	APIVersion            float64 `json:"ApiVersion"`
	LengthWithoutComments int     `json:"LengthWithoutComments"`
	NamespacePrefix       string  `json:"NamespacePrefix,omitempty"`
}

// ApexTrigger represents an Apex trigger in Salesforce.
type ApexTrigger struct {
	ID              string  `json:"Id"`
	Name            string  `json:"Name"`
	Body            string  `json:"Body,omitempty"`
	Status          string  `json:"Status"`
	IsValid         bool    `json:"IsValid"`
	APIVersion      float64 `json:"ApiVersion"`
	TableEnumOrID   string  `json:"TableEnumOrId"`
	NamespacePrefix string  `json:"NamespacePrefix,omitempty"`
}

// ApexLog represents a debug log entry.
type ApexLog struct {
	ID             string    `json:"Id"`
	LogUserID      string    `json:"LogUserId"`
	LogUserName    string    `json:"LogUser.Name,omitempty"`
	Operation      string    `json:"Operation"`
	Request        string    `json:"Request"`
	Status         string    `json:"Status"`
	LogLength      int       `json:"LogLength"`
	DurationMS     int       `json:"DurationMilliseconds"`
	StartTime      time.Time `json:"StartTime"`
	Location       string    `json:"Location"`
	Application    string    `json:"Application,omitempty"`
	LastModified   time.Time `json:"LastModifiedDate,omitempty"`
	SystemModstamp time.Time `json:"SystemModstamp,omitempty"`
}

// ApexTestQueueItem represents a test class in the test queue.
type ApexTestQueueItem struct {
	ID             string `json:"Id"`
	ApexClassID    string `json:"ApexClassId"`
	Status         string `json:"Status"`
	ExtendedStatus string `json:"ExtendedStatus,omitempty"`
	ParentJobID    string `json:"ParentJobId,omitempty"`
}

// ApexTestResult represents the result of running an Apex test method.
type ApexTestResult struct {
	ID             string `json:"Id"`
	ApexClassID    string `json:"ApexClassId"`
	ClassName      string `json:"ApexClass.Name,omitempty"`
	MethodName     string `json:"MethodName"`
	Outcome        string `json:"Outcome"` // Pass, Fail, CompileFail, Skip
	Message        string `json:"Message,omitempty"`
	StackTrace     string `json:"StackTrace,omitempty"`
	RunTime        int    `json:"RunTime"` // milliseconds
	AsyncApexJobID string `json:"AsyncApexJobId"`
	TestTimestamp  string `json:"TestTimestamp,omitempty"`
}

// ApexCodeCoverage represents code coverage for an Apex class.
type ApexCodeCoverage struct {
	ID                   string `json:"Id"`
	ApexClassOrTriggerID string `json:"ApexClassOrTriggerId"`
	ApexClassOrTrigger   struct {
		Name string `json:"Name"`
	} `json:"ApexClassOrTrigger,omitempty"`
	ApexTestClassID   string `json:"ApexTestClassId"`
	NumLinesCovered   int    `json:"NumLinesCovered"`
	NumLinesUncovered int    `json:"NumLinesUncovered"`
}

// ApexCodeCoverageAggregate represents aggregate code coverage.
type ApexCodeCoverageAggregate struct {
	ID                   string `json:"Id"`
	ApexClassOrTriggerID string `json:"ApexClassOrTriggerId"`
	ApexClassOrTrigger   struct {
		Name string `json:"Name"`
	} `json:"ApexClassOrTrigger,omitempty"`
	NumLinesCovered   int `json:"NumLinesCovered"`
	NumLinesUncovered int `json:"NumLinesUncovered"`
}

// ExecuteAnonymousResult represents the result of executing anonymous Apex.
type ExecuteAnonymousResult struct {
	Line                int    `json:"line"`
	Column              int    `json:"column"`
	Compiled            bool   `json:"compiled"`
	Success             bool   `json:"success"`
	CompileProblem      string `json:"compileProblem,omitempty"`
	ExceptionMessage    string `json:"exceptionMessage,omitempty"`
	ExceptionStackTrace string `json:"exceptionStackTrace,omitempty"`
}

// AsyncApexJob represents an asynchronous Apex job (for test runs).
type AsyncApexJob struct {
	ID                string `json:"Id"`
	Status            string `json:"Status"` // Queued, Processing, Completed, Aborted, Failed
	JobItemsProcessed int    `json:"JobItemsProcessed"`
	TotalJobItems     int    `json:"TotalJobItems"`
	NumberOfErrors    int    `json:"NumberOfErrors"`
	MethodName        string `json:"MethodName,omitempty"`
	ExtendedStatus    string `json:"ExtendedStatus,omitempty"`
	ParentJobID       string `json:"ParentJobId,omitempty"`
	ApexClassID       string `json:"ApexClassId,omitempty"`
	CompletedDate     string `json:"CompletedDate,omitempty"`
}

// QueryResult represents the result of a Tooling API query.
type QueryResult struct {
	TotalSize      int      `json:"totalSize"`
	Done           bool     `json:"done"`
	Records        []Record `json:"records"`
	NextRecordsURL string   `json:"nextRecordsUrl,omitempty"`
}

// Record represents a generic record from a Tooling API query.
type Record map[string]interface{}

// RunTestsRequest represents a request to run Apex tests.
type RunTestsRequest struct {
	ClassIDs       []string `json:"classids,omitempty"`
	SuiteIDs       []string `json:"suiteids,omitempty"`
	MaxFailedTests int      `json:"maxFailedTests,omitempty"`
	TestLevel      string   `json:"testLevel,omitempty"`
}

// RunTestsAsyncResult represents the result of enqueuing tests.
type RunTestsAsyncResult string
