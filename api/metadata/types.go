// Package metadata provides a client for the Salesforce Metadata API.
package metadata

import "time"

// MetadataType represents a metadata type available in the org.
type MetadataType struct {
	XMLName       string   `json:"xmlName"`
	DirectoryName string   `json:"directoryName"`
	Suffix        string   `json:"suffix"`
	InFolder      bool     `json:"inFolder"`
	MetaFile      bool     `json:"metaFile"`
	ChildNames    []string `json:"childXmlNames,omitempty"`
}

// MetadataComponent represents a metadata component in the org.
type MetadataComponent struct {
	ID               string    `json:"id,omitempty"`
	Type             string    `json:"type"`
	FullName         string    `json:"fullName"`
	FileName         string    `json:"fileName,omitempty"`
	NamespacePrefix  string    `json:"namespacePrefix,omitempty"`
	Creatable        bool      `json:"creatable"`
	Updateable       bool      `json:"updateable"`
	Deletable        bool      `json:"deletable"`
	LastModifiedBy   string    `json:"lastModifiedById,omitempty"`
	LastModifiedDate time.Time `json:"lastModifiedDate,omitempty"`
}

// DeployRequest represents a request to deploy metadata.
type DeployRequest struct {
	// Zip file as base64-encoded string
	ZipFile string `json:"zipFile"`
	// DeployOptions contains deployment options
	DeployOptions DeployOptions `json:"deployOptions,omitempty"`
}

// DeployOptions contains options for deployment.
type DeployOptions struct {
	AllowMissingFiles bool     `json:"allowMissingFiles,omitempty"`
	AutoUpdatePackage bool     `json:"autoUpdatePackage,omitempty"`
	CheckOnly         bool     `json:"checkOnly,omitempty"`
	IgnoreWarnings    bool     `json:"ignoreWarnings,omitempty"`
	PerformRetrieve   bool     `json:"performRetrieve,omitempty"`
	PurgeOnDelete     bool     `json:"purgeOnDelete,omitempty"`
	RollbackOnError   bool     `json:"rollbackOnError,omitempty"`
	SinglePackage     bool     `json:"singlePackage,omitempty"`
	TestLevel         string   `json:"testLevel,omitempty"` // NoTestRun, RunSpecifiedTests, RunLocalTests, RunAllTestsInOrg
	RunTests          []string `json:"runTests,omitempty"`
}

// DeployResult represents the result of a deployment.
type DeployResult struct {
	ID                       string         `json:"id"`
	Status                   string         `json:"status"` // Pending, InProgress, Succeeded, Failed, Canceling, Canceled
	Done                     bool           `json:"done"`
	Success                  bool           `json:"success"`
	CheckOnly                bool           `json:"checkOnly"`
	IgnoreWarnings           bool           `json:"ignoreWarnings"`
	NumberComponentsTotal    int            `json:"numberComponentsTotal"`
	NumberComponentsDeployed int            `json:"numberComponentsDeployed"`
	NumberComponentErrors    int            `json:"numberComponentErrors"`
	NumberTestsTotal         int            `json:"numberTestsTotal"`
	NumberTestsCompleted     int            `json:"numberTestsCompleted"`
	NumberTestErrors         int            `json:"numberTestErrors"`
	StartDate                time.Time      `json:"startDate,omitempty"`
	CompletedDate            time.Time      `json:"completedDate,omitempty"`
	ErrorMessage             string         `json:"errorMessage,omitempty"`
	ErrorStatusCode          string         `json:"errorStatusCode,omitempty"`
	StateDetail              string         `json:"stateDetail,omitempty"`
	DeployDetails            *DeployDetails `json:"details,omitempty"`
}

// DeployDetails contains detailed information about deployment results.
type DeployDetails struct {
	ComponentSuccesses []ComponentResult `json:"componentSuccesses,omitempty"`
	ComponentFailures  []ComponentResult `json:"componentFailures,omitempty"`
	RunTestResult      *RunTestResult    `json:"runTestResult,omitempty"`
}

// ComponentResult represents the result of deploying a single component.
type ComponentResult struct {
	ComponentType string `json:"componentType"`
	FullName      string `json:"fullName"`
	FileName      string `json:"fileName,omitempty"`
	Success       bool   `json:"success"`
	Changed       bool   `json:"changed"`
	Created       bool   `json:"created"`
	Deleted       bool   `json:"deleted"`
	Problem       string `json:"problem,omitempty"`
	ProblemType   string `json:"problemType,omitempty"`
	LineNumber    int    `json:"lineNumber,omitempty"`
	ColumnNumber  int    `json:"columnNumber,omitempty"`
}

// RunTestResult contains test execution results from deployment.
type RunTestResult struct {
	NumTestsRun int `json:"numTestsRun"`
	NumFailures int `json:"numFailures"`
	TotalTime   int `json:"totalTime"` // milliseconds
}

// RetrieveRequest represents a request to retrieve metadata.
type RetrieveRequest struct {
	APIVersion    string   `json:"apiVersion"`
	SinglePackage bool     `json:"singlePackage"`
	PackageNames  []string `json:"packageNames,omitempty"`
	Unpackaged    *Package `json:"unpackaged,omitempty"`
}

// Package represents a package.xml structure.
type Package struct {
	Types   []PackageType `json:"types"`
	Version string        `json:"version"`
}

// PackageType represents a type section in package.xml.
type PackageType struct {
	Members []string `json:"members"`
	Name    string   `json:"name"`
}

// RetrieveResult represents the result of a retrieve operation.
type RetrieveResult struct {
	ID              string `json:"id"`
	Status          string `json:"status"` // Pending, InProgress, Succeeded, Failed
	Done            bool   `json:"done"`
	Success         bool   `json:"success"`
	ZipFile         string `json:"zipFile,omitempty"` // Base64-encoded zip
	ErrorMessage    string `json:"errorMessage,omitempty"`
	ErrorStatusCode string `json:"errorStatusCode,omitempty"`
}

// DescribeMetadataResult represents the result of describing metadata.
type DescribeMetadataResult struct {
	MetadataObjects       []MetadataType `json:"metadataObjects"`
	OrganizationNamespace string         `json:"organizationNamespace"`
	PartialSaveAllowed    bool           `json:"partialSaveAllowed"`
	TestRequired          bool           `json:"testRequired"`
}

// ListMetadataQuery represents a query for listing metadata.
type ListMetadataQuery struct {
	Type   string `json:"type"`
	Folder string `json:"folder,omitempty"`
}
