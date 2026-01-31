package api

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "empty errors",
			err: &APIError{
				StatusCode: 400,
				Errors:     nil,
			},
			expected: "API error: HTTP 400",
		},
		{
			name: "single error",
			err: &APIError{
				StatusCode: 400,
				Errors: []SalesforceError{
					{ErrorCode: "MALFORMED_QUERY", Message: "Invalid SOQL"},
				},
			},
			expected: "MALFORMED_QUERY: Invalid SOQL",
		},
		{
			name: "single error with fields",
			err: &APIError{
				StatusCode: 400,
				Errors: []SalesforceError{
					{ErrorCode: "INVALID_FIELD", Message: "No such column 'foo'", Fields: []string{"foo"}},
				},
			},
			expected: "INVALID_FIELD: No such column 'foo' (fields: foo)",
		},
		{
			name: "multiple errors",
			err: &APIError{
				StatusCode: 400,
				Errors: []SalesforceError{
					{ErrorCode: "REQUIRED_FIELD_MISSING", Message: "Required field missing: Name"},
					{ErrorCode: "INVALID_FIELD", Message: "Invalid field: BadField", Fields: []string{"BadField"}},
				},
			},
			expected: "REQUIRED_FIELD_MISSING: Required field missing: Name; INVALID_FIELD: Invalid field: BadField (fields: BadField)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		name    string
		err     *APIError
		wantErr error
	}{
		{
			name:    "401 unauthorized",
			err:     &APIError{StatusCode: 401},
			wantErr: ErrUnauthorized,
		},
		{
			name: "401 invalid session",
			err: &APIError{
				StatusCode: 401,
				Errors:     []SalesforceError{{ErrorCode: "INVALID_SESSION_ID", Message: "Session expired"}},
			},
			wantErr: ErrInvalidSession,
		},
		{
			name:    "403 forbidden",
			err:     &APIError{StatusCode: 403},
			wantErr: ErrForbidden,
		},
		{
			name:    "404 not found",
			err:     &APIError{StatusCode: 404},
			wantErr: ErrNotFound,
		},
		{
			name:    "400 bad request",
			err:     &APIError{StatusCode: 400},
			wantErr: ErrBadRequest,
		},
		{
			name:    "429 rate limited",
			err:     &APIError{StatusCode: 429},
			wantErr: ErrRateLimited,
		},
		{
			name:    "500 server error",
			err:     &APIError{StatusCode: 500},
			wantErr: ErrServerError,
		},
		{
			name:    "503 server error",
			err:     &APIError{StatusCode: 503},
			wantErr: ErrServerError,
		},
		{
			name:    "200 no error",
			err:     &APIError{StatusCode: 200},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Unwrap()
			assert.Equal(t, tt.wantErr, got)
		})
	}
}

func TestIsHelpers(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		check  func(error) bool
		expect bool
	}{
		{"IsNotFound true", &APIError{StatusCode: 404}, IsNotFound, true},
		{"IsNotFound false", &APIError{StatusCode: 400}, IsNotFound, false},
		{"IsUnauthorized true", &APIError{StatusCode: 401}, IsUnauthorized, true},
		{"IsUnauthorized invalid session", &APIError{StatusCode: 401, Errors: []SalesforceError{{ErrorCode: "INVALID_SESSION_ID"}}}, IsUnauthorized, true},
		{"IsForbidden true", &APIError{StatusCode: 403}, IsForbidden, true},
		{"IsBadRequest true", &APIError{StatusCode: 400}, IsBadRequest, true},
		{"IsRateLimited true", &APIError{StatusCode: 429}, IsRateLimited, true},
		{"IsServerError true", &APIError{StatusCode: 500}, IsServerError, true},
		{"IsInvalidSession true", &APIError{StatusCode: 401, Errors: []SalesforceError{{ErrorCode: "INVALID_SESSION_ID"}}}, IsInvalidSession, true},
		{"IsInvalidSession false", &APIError{StatusCode: 401}, IsInvalidSession, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.check(tt.err)
			assert.Equal(t, tt.expect, got)
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		body          string
		expectErrors  int
		expectCode    string
		expectMessage string
	}{
		{
			name:          "array format",
			statusCode:    400,
			body:          `[{"errorCode": "MALFORMED_QUERY", "message": "Invalid SOQL"}]`,
			expectErrors:  1,
			expectCode:    "MALFORMED_QUERY",
			expectMessage: "Invalid SOQL",
		},
		{
			name:          "single object format",
			statusCode:    400,
			body:          `{"errorCode": "INVALID_FIELD", "message": "No such column"}`,
			expectErrors:  1,
			expectCode:    "INVALID_FIELD",
			expectMessage: "No such column",
		},
		{
			name:          "multiple errors",
			statusCode:    400,
			body:          `[{"errorCode": "ERR1", "message": "First"}, {"errorCode": "ERR2", "message": "Second"}]`,
			expectErrors:  2,
			expectCode:    "ERR1",
			expectMessage: "First",
		},
		{
			name:          "plain text fallback",
			statusCode:    500,
			body:          `Internal Server Error`,
			expectErrors:  1,
			expectCode:    "",
			expectMessage: "Internal Server Error",
		},
		{
			name:          "with fields",
			statusCode:    400,
			body:          `[{"errorCode": "INVALID_FIELD", "message": "Bad field", "fields": ["Name", "Email"]}]`,
			expectErrors:  1,
			expectCode:    "INVALID_FIELD",
			expectMessage: "Bad field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}

			err := ParseAPIError(resp)
			apiErr, ok := err.(*APIError)
			assert.True(t, ok)
			assert.Equal(t, tt.statusCode, apiErr.StatusCode)
			assert.Len(t, apiErr.Errors, tt.expectErrors)
			if tt.expectErrors > 0 {
				assert.Equal(t, tt.expectCode, apiErr.Errors[0].ErrorCode)
				assert.Equal(t, tt.expectMessage, apiErr.Errors[0].Message)
			}
		})
	}
}

func TestErrorsIs(t *testing.T) {
	// Test that errors.Is works correctly with APIError
	apiErr := &APIError{StatusCode: 404}

	assert.True(t, errors.Is(apiErr, ErrNotFound))
	assert.False(t, errors.Is(apiErr, ErrUnauthorized))
}
