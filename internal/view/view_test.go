package view

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidFormats(t *testing.T) {
	formats := ValidFormats()
	assert.Contains(t, formats, "table")
	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "plain")
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		format  string
		wantErr bool
	}{
		{"", false},
		{"table", false},
		{"json", false},
		{"plain", false},
		{"invalid", true},
		{"XML", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	v := New(FormatTable, true) // noColor for consistent testing
	v.SetOutput(&buf)

	headers := []string{"ID", "NAME", "STATUS"}
	rows := [][]string{
		{"001", "Account 1", "Active"},
		{"002", "Account 2", "Inactive"},
	}

	err := v.Table(headers, rows)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "Account 1")
	assert.Contains(t, output, "Account 2")
}

func TestTableAsJSON(t *testing.T) {
	var buf bytes.Buffer
	v := New(FormatJSON, true)
	v.SetOutput(&buf)

	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"001", "Test"},
	}

	err := v.Table(headers, rows)
	require.NoError(t, err)

	var result []map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, "001", result[0]["id"])
	assert.Equal(t, "Test", result[0]["name"])
}

func TestPlain(t *testing.T) {
	var buf bytes.Buffer
	v := New(FormatPlain, true)
	v.SetOutput(&buf)

	rows := [][]string{
		{"001", "Test"},
		{"002", "Test2"},
	}

	err := v.Plain(rows)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "001")
	assert.Contains(t, lines[1], "002")
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	v := New(FormatJSON, true)
	v.SetOutput(&buf)

	data := map[string]string{"key": "value"}
	err := v.JSON(data)
	require.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

func TestRender(t *testing.T) {
	headers := []string{"ID", "NAME"}
	rows := [][]string{{"001", "Test"}}
	jsonData := map[string]string{"id": "001"}

	tests := []struct {
		name   string
		format Format
		check  func(t *testing.T, output string)
	}{
		{
			name:   "table",
			format: FormatTable,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "ID")
				assert.Contains(t, output, "001")
			},
		},
		{
			name:   "json",
			format: FormatJSON,
			check: func(t *testing.T, output string) {
				var result map[string]string
				err := json.Unmarshal([]byte(output), &result)
				require.NoError(t, err)
				assert.Equal(t, "001", result["id"])
			},
		},
		{
			name:   "plain",
			format: FormatPlain,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "001")
				assert.NotContains(t, output, "ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			v := New(tt.format, true)
			v.SetOutput(&buf)

			err := v.Render(headers, rows, jsonData)
			require.NoError(t, err)
			tt.check(t, buf.String())
		})
	}
}

func TestSuccessErrorWarningInfo(t *testing.T) {
	tests := []struct {
		name   string
		method func(v *View)
		want   string
		isErr  bool
	}{
		{
			name:   "Success",
			method: func(v *View) { v.Success("done") },
			want:   "✓ done",
			isErr:  false,
		},
		{
			name:   "Error",
			method: func(v *View) { v.Error("failed") },
			want:   "✗ failed",
			isErr:  true,
		},
		{
			name:   "Warning",
			method: func(v *View) { v.Warning("caution") },
			want:   "⚠ caution",
			isErr:  true,
		},
		{
			name:   "Info",
			method: func(v *View) { v.Info("info") },
			want:   "info",
			isErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, err bytes.Buffer
			v := New(FormatTable, true)
			v.SetOutput(&out)
			v.SetError(&err)

			tt.method(v)

			if tt.isErr {
				assert.Contains(t, err.String(), tt.want)
			} else {
				assert.Contains(t, out.String(), tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"hello", 5, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}
