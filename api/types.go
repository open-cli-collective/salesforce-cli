// Package api provides a Go client for the Salesforce REST API.
package api

import (
	"encoding/json"
	"time"
)

// APIVersion represents a Salesforce API version
type APIVersion struct {
	Label   string `json:"label"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

// SObject represents a generic Salesforce object record
type SObject struct {
	// Attributes contains metadata about the record
	Attributes SObjectAttributes `json:"attributes,omitempty"`

	// ID is the Salesforce record ID (18-character)
	ID string `json:"Id,omitempty"`

	// Fields contains all other fields as a map
	// Use GetString, GetBool, etc. helpers to access typed values
	Fields map[string]interface{} `json:"-"`
}

// SObjectAttributes contains metadata about an SObject record
type SObjectAttributes struct {
	Type string `json:"type"`
	URL  string `json:"url,omitempty"`
}

// UnmarshalJSON custom unmarshaler to capture all fields
func (s *SObject) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to get all fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Fields = make(map[string]interface{})

	for key, value := range raw {
		switch key {
		case "attributes":
			if err := json.Unmarshal(value, &s.Attributes); err != nil {
				return err
			}
		case "Id":
			if err := json.Unmarshal(value, &s.ID); err != nil {
				return err
			}
		default:
			var v interface{}
			if err := json.Unmarshal(value, &v); err != nil {
				return err
			}
			s.Fields[key] = v
		}
	}

	return nil
}

// MarshalJSON custom marshaler to include all fields
func (s SObject) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})

	if s.Attributes.Type != "" {
		result["attributes"] = s.Attributes
	}
	if s.ID != "" {
		result["Id"] = s.ID
	}

	for key, value := range s.Fields {
		result[key] = value
	}

	return json.Marshal(result)
}

// GetString returns a string field value
func (s *SObject) GetString(field string) string {
	if s.Fields == nil {
		return ""
	}
	if v, ok := s.Fields[field].(string); ok {
		return v
	}
	return ""
}

// GetBool returns a boolean field value
func (s *SObject) GetBool(field string) bool {
	if s.Fields == nil {
		return false
	}
	if v, ok := s.Fields[field].(bool); ok {
		return v
	}
	return false
}

// GetFloat returns a numeric field value as float64
func (s *SObject) GetFloat(field string) float64 {
	if s.Fields == nil {
		return 0
	}
	if v, ok := s.Fields[field].(float64); ok {
		return v
	}
	return 0
}

// GetInt returns a numeric field value as int
func (s *SObject) GetInt(field string) int {
	return int(s.GetFloat(field))
}

// GetTime parses a datetime field value
func (s *SObject) GetTime(field string) time.Time {
	str := s.GetString(field)
	if str == "" {
		return time.Time{}
	}
	// Salesforce datetime format: 2023-01-15T10:30:00.000+0000
	t, _ := time.Parse("2006-01-02T15:04:05.000-0700", str)
	return t
}

// QueryResult represents the result of a SOQL query
type QueryResult struct {
	TotalSize      int       `json:"totalSize"`
	Done           bool      `json:"done"`
	NextRecordsURL string    `json:"nextRecordsUrl,omitempty"`
	Records        []SObject `json:"records"`
}

// RecordResult represents the result of a record create/update operation
type RecordResult struct {
	ID      string        `json:"id,omitempty"`
	Success bool          `json:"success"`
	Errors  []RecordError `json:"errors,omitempty"`
	Created bool          `json:"created,omitempty"`
}

// RecordError represents an error from a record operation
type RecordError struct {
	StatusCode string   `json:"statusCode"`
	Message    string   `json:"message"`
	Fields     []string `json:"fields,omitempty"`
}

// CompositeRequest represents a composite API request
type CompositeRequest struct {
	AllOrNone          bool                  `json:"allOrNone"`
	CollateSubrequests bool                  `json:"collateSubrequests,omitempty"`
	CompositeRequest   []CompositeSubrequest `json:"compositeRequest"`
}

// CompositeSubrequest represents a single request within a composite batch
type CompositeSubrequest struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	ReferenceID string                 `json:"referenceId"`
	Body        map[string]interface{} `json:"body,omitempty"`
}

// CompositeResponse represents a composite API response
type CompositeResponse struct {
	CompositeResponse []CompositeSubresponse `json:"compositeResponse"`
}

// CompositeSubresponse represents a single response within a composite batch
type CompositeSubresponse struct {
	Body           json.RawMessage   `json:"body"`
	HTTPHeaders    map[string]string `json:"httpHeaders"`
	HTTPStatusCode int               `json:"httpStatusCode"`
	ReferenceID    string            `json:"referenceId"`
}

// SObjectDescribe represents metadata about an SObject type
type SObjectDescribe struct {
	Name        string  `json:"name"`
	Label       string  `json:"label"`
	LabelPlural string  `json:"labelPlural"`
	KeyPrefix   string  `json:"keyPrefix,omitempty"`
	Custom      bool    `json:"custom"`
	Createable  bool    `json:"createable"`
	Updateable  bool    `json:"updateable"`
	Deletable   bool    `json:"deletable"`
	Queryable   bool    `json:"queryable"`
	Searchable  bool    `json:"searchable"`
	Fields      []Field `json:"fields,omitempty"`
}

// Field represents a field on an SObject
type Field struct {
	Name              string          `json:"name"`
	Label             string          `json:"label"`
	Type              string          `json:"type"`
	Length            int             `json:"length,omitempty"`
	Precision         int             `json:"precision,omitempty"`
	Scale             int             `json:"scale,omitempty"`
	Nillable          bool            `json:"nillable"`
	Createable        bool            `json:"createable"`
	Updateable        bool            `json:"updateable"`
	Custom            bool            `json:"custom"`
	CalculatedFormula string          `json:"calculatedFormula,omitempty"`
	DefaultValue      interface{}     `json:"defaultValue,omitempty"`
	PicklistValues    []PicklistValue `json:"picklistValues,omitempty"`
	ReferenceTo       []string        `json:"referenceTo,omitempty"`
	RelationshipName  string          `json:"relationshipName,omitempty"`
}

// PicklistValue represents a picklist option
type PicklistValue struct {
	Value        string `json:"value"`
	Label        string `json:"label"`
	Active       bool   `json:"active"`
	DefaultValue bool   `json:"defaultValue"`
}

// SObjectsResponse represents the response from /sobjects/
type SObjectsResponse struct {
	Encoding     string            `json:"encoding"`
	MaxBatchSize int               `json:"maxBatchSize"`
	SObjects     []SObjectDescribe `json:"sobjects"`
}

// Limits represents Salesforce org limits
type Limits map[string]LimitInfo

// LimitInfo represents a single limit value
type LimitInfo struct {
	Max       int `json:"Max"`
	Remaining int `json:"Remaining"`
}

// SearchResult represents the result of a SOSL search
type SearchResult struct {
	SearchRecords []SearchRecord `json:"searchRecords"`
}

// SearchRecord represents a single search result record
type SearchRecord struct {
	Attributes SObjectAttributes      `json:"attributes"`
	ID         string                 `json:"Id"`
	Fields     map[string]interface{} `json:"-"`
}

// UnmarshalJSON custom unmarshaler for SearchRecord to capture all fields
func (s *SearchRecord) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Fields = make(map[string]interface{})

	for key, value := range raw {
		switch key {
		case "attributes":
			if err := json.Unmarshal(value, &s.Attributes); err != nil {
				return err
			}
		case "Id":
			if err := json.Unmarshal(value, &s.ID); err != nil {
				return err
			}
		default:
			var v interface{}
			if err := json.Unmarshal(value, &v); err != nil {
				return err
			}
			s.Fields[key] = v
		}
	}

	return nil
}
