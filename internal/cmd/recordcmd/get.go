package recordcmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newGetCommand(opts *root.Options) *cobra.Command {
	var fields string

	cmd := &cobra.Command{
		Use:   "get <object> <id>",
		Short: "Get a record by ID",
		Long: `Retrieve a Salesforce record by its ID.

Examples:
  sfdc record get Account 001xx000003DGbYAAW
  sfdc record get Contact 003xx000001abcd --fields Name,Email,Phone
  sfdc record get Account 001xx000003DGbYAAW -o json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var fieldList []string
			if fields != "" {
				fieldList = strings.Split(fields, ",")
				for i := range fieldList {
					fieldList[i] = strings.TrimSpace(fieldList[i])
				}
			}
			return runGet(cmd.Context(), opts, args[0], args[1], fieldList)
		},
	}

	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated list of fields to retrieve")

	return cmd
}

func runGet(ctx context.Context, opts *root.Options, objectName, recordID string, fields []string) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	record, err := client.GetRecord(ctx, objectName, recordID, fields)
	if err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}

	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(record)
	}

	// Display as key-value pairs
	v.Info("Object: %s", record.Attributes.Type)
	v.Info("ID: %s", record.ID)
	v.Info("")

	// Sort field names for consistent output
	fieldNames := make([]string, 0, len(record.Fields))
	for name := range record.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for _, name := range fieldNames {
		value := formatFieldValue(record.Fields[name])
		v.Info("%s: %s", name, value)
	}

	// Show record URL
	v.Info("")
	v.Info("URL: %s", client.RecordURL(record.ID))

	return nil
}

// formatFieldValue converts a field value to a string for display
func formatFieldValue(v interface{}) string {
	if v == nil {
		return "(null)"
	}

	switch val := v.(type) {
	case string:
		return val
	case float64:
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
		if name, ok := val["Name"].(string); ok {
			return name
		}
		return "[object]"
	default:
		return fmt.Sprintf("%v", val)
	}
}
