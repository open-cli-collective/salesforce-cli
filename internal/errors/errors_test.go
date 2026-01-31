package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *APIError
		contains []string
	}{
		{
			name:     "empty errors",
			apiErr:   &APIError{StatusCode: 400},
			contains: []string{"API error", "400"},
		},
		{
			name: "single error",
			apiErr: &APIError{
				StatusCode: 400,
				Errors: []SalesforceError{
					{ErrorCode: "INVALID_FIELD", Message: "Field not found"},
				},
			},
			contains: []string{"INVALID_FIELD", "Field not found"},
		},
		{
			name: "error with fields",
			apiErr: &APIError{
				StatusCode: 400,
				Errors: []SalesforceError{
					{ErrorCode: "REQUIRED_FIELD_MISSING", Message: "Required field missing", Fields: []string{"Name", "Email"}},
				},
			},
			contains: []string{"REQUIRED_FIELD_MISSING", "Name", "Email"},
		},
		{
			name: "multiple errors",
			apiErr: &APIError{
				StatusCode: 400,
				Errors: []SalesforceError{
					{ErrorCode: "ERROR1", Message: "First error"},
					{ErrorCode: "ERROR2", Message: "Second error"},
				},
			},
			contains: []string{"ERROR1", "ERROR2", "First error", "Second error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.apiErr.Error()
			for _, want := range tt.contains {
				assert.Contains(t, errStr, want)
			}
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantErr    error
	}{
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       nil,
			wantErr:    ErrUnauthorized,
		},
		{
			name:       "403 forbidden",
			statusCode: http.StatusForbidden,
			body:       nil,
			wantErr:    ErrForbidden,
		},
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			body:       nil,
			wantErr:    ErrNotFound,
		},
		{
			name:       "400 bad request",
			statusCode: http.StatusBadRequest,
			body:       nil,
			wantErr:    ErrBadRequest,
		},
		{
			name:       "429 rate limited",
			statusCode: http.StatusTooManyRequests,
			body:       nil,
			wantErr:    ErrRateLimited,
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			body:       nil,
			wantErr:    ErrServerError,
		},
		{
			name:       "401 with details",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`[{"errorCode":"INVALID_SESSION_ID","message":"Session expired or invalid"}]`),
			wantErr:    ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseAPIError(tt.statusCode, tt.body)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestParseAPIError_WithBody(t *testing.T) {
	body := []byte(`[{"errorCode":"INVALID_FIELD","message":"No such column 'foo'"}]`)
	err := ParseAPIError(http.StatusBadRequest, body)

	assert.True(t, IsBadRequest(err))
	assert.Contains(t, err.Error(), "INVALID_FIELD")
	assert.Contains(t, err.Error(), "No such column")
}

func TestParseAPIError_OtherStatusCode(t *testing.T) {
	// 418 I'm a teapot - not a standard error
	body := []byte(`[{"errorCode":"TEAPOT","message":"I am a teapot"}]`)
	err := ParseAPIError(418, body)

	// Should return APIError directly, not wrapped
	apiErr, ok := err.(*APIError)
	assert.True(t, ok)
	assert.Equal(t, 418, apiErr.StatusCode)
}

func TestIsHelpers(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		checker func(error) bool
		want    bool
	}{
		{"IsNotFound true", ErrNotFound, IsNotFound, true},
		{"IsNotFound false", ErrUnauthorized, IsNotFound, false},
		{"IsUnauthorized true", ErrUnauthorized, IsUnauthorized, true},
		{"IsForbidden true", ErrForbidden, IsForbidden, true},
		{"IsBadRequest true", ErrBadRequest, IsBadRequest, true},
		{"IsRateLimited true", ErrRateLimited, IsRateLimited, true},
		{"IsServerError true", ErrServerError, IsServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.checker(tt.err))
		})
	}
}
