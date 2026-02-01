// Package querycmd provides the query command for executing SOQL queries.
package querycmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the query command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the query command.
func NewCommand(opts *root.Options) *cobra.Command {
	var (
		all     bool
		noLimit bool
	)

	cmd := &cobra.Command{
		Use:   "query <soql>",
		Short: "Execute a SOQL query",
		Long: `Execute a SOQL query against Salesforce and display the results.

Examples:
  sfdc query "SELECT Id, Name FROM Account LIMIT 10"
  sfdc query "SELECT Id, Name FROM Account" --all
  sfdc query "SELECT Id, Name, Phone FROM Contact" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuery(cmd.Context(), opts, args[0], all, noLimit)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Include deleted and archived records (queryAll)")
	cmd.Flags().BoolVar(&noLimit, "no-limit", false, "Fetch all pages of results (may be slow for large datasets)")

	return cmd
}

func runQuery(ctx context.Context, opts *root.Options, soql string, all, noLimit bool) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	var result *api.QueryResult

	if all {
		// Use queryAll to include deleted/archived records
		result, err = queryAllRecords(ctx, client, soql)
	} else if noLimit {
		// Fetch all pages
		result, err = client.QueryAll(ctx, soql)
	} else {
		// Single page query
		result, err = client.Query(ctx, soql)
	}

	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	return renderQueryResult(opts, result)
}

// queryAllRecords uses the queryAll endpoint to include deleted/archived records
func queryAllRecords(ctx context.Context, client *api.Client, soql string) (*api.QueryResult, error) {
	// The queryAll endpoint is at /queryAll instead of /query
	// We need to make a direct request since the client doesn't have this method
	path := fmt.Sprintf("/queryAll?q=%s", url.QueryEscape(soql))

	// Use the client's Get method with URL encoding handled
	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result api.QueryResult
	if err := parseJSON(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	return &result, nil
}

func renderQueryResult(opts *root.Options, result *api.QueryResult) error {
	v := opts.View()

	if len(result.Records) == 0 {
		v.Info("No records found (totalSize: %d)", result.TotalSize)
		return nil
	}

	// For JSON output, render the full result
	if opts.Output == "json" {
		return v.JSON(result)
	}

	// For table/plain output, extract field names from first record
	headers := extractHeaders(result.Records)
	rows := extractRows(result.Records, headers)

	// Add record count footer
	if err := v.Table(headers, rows); err != nil {
		return err
	}

	// Show pagination info if not all records fetched
	if !result.Done {
		v.Info("\nShowing %d of %d records (use --no-limit to fetch all)", len(result.Records), result.TotalSize)
	} else {
		v.Info("\n%d record(s)", result.TotalSize)
	}

	return nil
}

// extractHeaders gets column headers from the first record
func extractHeaders(records []api.SObject) []string {
	if len(records) == 0 {
		return nil
	}

	headers := []string{"Id"}

	// Get field names from first record and sort for consistency
	first := records[0]
	fieldNames := make([]string, 0, len(first.Fields))
	for name := range first.Fields {
		if name != "Id" { // Id is handled separately
			fieldNames = append(fieldNames, name)
		}
	}
	sort.Strings(fieldNames)

	return append(headers, fieldNames...)
}

// extractRows converts records to string rows for table output
func extractRows(records []api.SObject, headers []string) [][]string {
	rows := make([][]string, 0, len(records))

	for _, rec := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			if header == "Id" {
				row[i] = rec.ID
			} else {
				row[i] = formatFieldValue(rec.Fields[header])
			}
		}
		rows = append(rows, row)
	}

	return rows
}

// formatFieldValue converts a field value to a string for display
func formatFieldValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case float64:
		// Check if it's a whole number
		if val == float64(int64(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case map[string]interface{}:
		// Nested object (e.g., relationship)
		if name, ok := val["Name"].(string); ok {
			return name
		}
		return "[object]"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// parseJSON is a helper to parse JSON response
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
