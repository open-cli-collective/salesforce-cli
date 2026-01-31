// Package errors provides error types for Salesforce API responses.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors for common HTTP status codes.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized: check your credentials")
	ErrForbidden    = errors.New("forbidden: insufficient permissions")
	ErrBadRequest   = errors.New("bad request")
	ErrRateLimited  = errors.New("rate limited: too many requests")
	ErrServerError  = errors.New("server error")
)

// SalesforceError represents a single error from the Salesforce API.
// Salesforce returns errors as an array: [{"errorCode": "...", "message": "...", "fields": [...]}]
type SalesforceError struct {
	ErrorCode string   `json:"errorCode"`
	Message   string   `json:"message"`
	Fields    []string `json:"fields,omitempty"`
}

// APIError represents an error response from the Salesforce API.
type APIError struct {
	StatusCode int
	Errors     []SalesforceError
}

// Error returns a human-readable error message.
func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("API error (status %d)", e.StatusCode)
	}

	var parts []string
	for _, err := range e.Errors {
		msg := err.Message
		if err.ErrorCode != "" {
			msg = fmt.Sprintf("[%s] %s", err.ErrorCode, err.Message)
		}
		if len(err.Fields) > 0 {
			msg = fmt.Sprintf("%s (fields: %s)", msg, strings.Join(err.Fields, ", "))
		}
		parts = append(parts, msg)
	}

	return strings.Join(parts, "; ")
}

// ParseAPIError parses an HTTP response body into an appropriate error type.
// It returns sentinel errors for common status codes, wrapping the APIError
// details when additional information is available.
func ParseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}

	if len(body) > 0 {
		// Salesforce returns errors as an array
		var sfErrors []SalesforceError
		if err := json.Unmarshal(body, &sfErrors); err == nil {
			apiErr.Errors = sfErrors
		}
	}

	// Determine the base sentinel error
	var sentinel error
	switch statusCode {
	case http.StatusUnauthorized:
		sentinel = ErrUnauthorized
	case http.StatusForbidden:
		sentinel = ErrForbidden
	case http.StatusNotFound:
		sentinel = ErrNotFound
	case http.StatusBadRequest:
		sentinel = ErrBadRequest
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		if statusCode >= 500 {
			sentinel = ErrServerError
		} else {
			// For other 4xx errors, just return the APIError
			return apiErr
		}
	}

	// If we have additional details, wrap the sentinel error
	details := apiErr.Error()
	genericMsg := fmt.Sprintf("API error (status %d)", statusCode)
	if details != genericMsg {
		return fmt.Errorf("%w: %s", sentinel, details)
	}

	return sentinel
}

// IsNotFound returns true if the error is or wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized returns true if the error is or wraps ErrUnauthorized.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden returns true if the error is or wraps ErrForbidden.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsBadRequest returns true if the error is or wraps ErrBadRequest.
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

// IsRateLimited returns true if the error is or wraps ErrRateLimited.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsServerError returns true if the error is or wraps ErrServerError.
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError)
}
