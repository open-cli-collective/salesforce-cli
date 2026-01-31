package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Sentinel errors for common API error conditions
var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized - check your OAuth token")
	ErrForbidden      = errors.New("forbidden - insufficient permissions")
	ErrBadRequest     = errors.New("bad request")
	ErrRateLimited    = errors.New("rate limited - try again later")
	ErrServerError    = errors.New("server error")
	ErrInvalidSession = errors.New("invalid session - token may be expired")
)

// Validation errors
var (
	ErrInstanceURLRequired = errors.New("instance URL is required")
	ErrHTTPClientRequired  = errors.New("HTTP client is required")
)

// APIError represents a Salesforce API error response.
// Salesforce returns errors as an array: [{"errorCode": "...", "message": "...", "fields": [...]}]
type APIError struct {
	StatusCode int
	Errors     []SalesforceError
}

// SalesforceError represents a single error from the Salesforce API
type SalesforceError struct {
	ErrorCode string   `json:"errorCode"`
	Message   string   `json:"message"`
	Fields    []string `json:"fields,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("API error: HTTP %d", e.StatusCode)
	}

	if len(e.Errors) == 1 {
		err := e.Errors[0]
		if len(err.Fields) > 0 {
			return fmt.Sprintf("%s: %s (fields: %s)", err.ErrorCode, err.Message, strings.Join(err.Fields, ", "))
		}
		return fmt.Sprintf("%s: %s", err.ErrorCode, err.Message)
	}

	// Multiple errors
	var msgs []string
	for _, err := range e.Errors {
		if len(err.Fields) > 0 {
			msgs = append(msgs, fmt.Sprintf("%s: %s (fields: %s)", err.ErrorCode, err.Message, strings.Join(err.Fields, ", ")))
		} else {
			msgs = append(msgs, fmt.Sprintf("%s: %s", err.ErrorCode, err.Message))
		}
	}
	return strings.Join(msgs, "; ")
}

// Unwrap returns the underlying sentinel error based on status code
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusUnauthorized:
		// Check for invalid session
		for _, err := range e.Errors {
			if err.ErrorCode == "INVALID_SESSION_ID" {
				return ErrInvalidSession
			}
		}
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		if e.StatusCode >= 500 {
			return ErrServerError
		}
		return nil
	}
}

// IsNotFound returns true if the error indicates a resource was not found
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized returns true if the error indicates an authentication failure
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrInvalidSession)
}

// IsForbidden returns true if the error indicates insufficient permissions
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsBadRequest returns true if the error indicates a bad request
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

// IsRateLimited returns true if the error indicates rate limiting
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsServerError returns true if the error indicates a server error
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError)
}

// IsInvalidSession returns true if the error indicates an expired/invalid session
func IsInvalidSession(err error) bool {
	return errors.Is(err, ErrInvalidSession)
}

// ParseAPIError creates an APIError from an HTTP response.
// The response body is read and closed.
func ParseAPIError(resp *http.Response) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Errors:     []SalesforceError{{Message: "failed to read error response"}},
		}
	}

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
	}

	// Try to parse as Salesforce error array
	var sfErrors []SalesforceError
	if err := json.Unmarshal(body, &sfErrors); err == nil && len(sfErrors) > 0 {
		apiErr.Errors = sfErrors
		return apiErr
	}

	// Try to parse as single error object
	var sfError SalesforceError
	if err := json.Unmarshal(body, &sfError); err == nil && sfError.ErrorCode != "" {
		apiErr.Errors = []SalesforceError{sfError}
		return apiErr
	}

	// Fall back to raw body as message
	if len(body) > 0 {
		apiErr.Errors = []SalesforceError{{Message: string(body)}}
	}

	return apiErr
}

// Common Salesforce error codes for reference:
// - INVALID_SESSION_ID: Session expired or invalid
// - INVALID_FIELD: Field doesn't exist or is inaccessible
// - INVALID_TYPE: SObject type doesn't exist
// - MALFORMED_QUERY: Invalid SOQL syntax
// - MALFORMED_ID: Invalid record ID format
// - INSUFFICIENT_ACCESS_ON_CROSS_REFERENCE_ENTITY: Insufficient permissions
// - DUPLICATE_VALUE: Unique constraint violation
// - REQUIRED_FIELD_MISSING: Required field not provided
// - FIELD_CUSTOM_VALIDATION_EXCEPTION: Custom validation rule failed
// - STRING_TOO_LONG: Field value exceeds maximum length
// - NUMBER_OUTSIDE_VALID_RANGE: Numeric value out of range
