package recordcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newCreateCommand(opts *root.Options) *cobra.Command {
	var setFlags []string

	cmd := &cobra.Command{
		Use:   "create <object>",
		Short: "Create a new record",
		Long: `Create a new Salesforce record.

Examples:
  sfdc record create Account --set Name="Acme Corp"
  sfdc record create Contact --set FirstName=John --set LastName=Doe --set Email=john@example.com
  sfdc record create Account --set Name="Test" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := parseSetFlags(setFlags)
			if err != nil {
				return err
			}
			if len(fields) == 0 {
				return fmt.Errorf("at least one --set flag is required")
			}
			return runCreate(cmd.Context(), opts, args[0], fields)
		},
	}

	cmd.Flags().StringArrayVar(&setFlags, "set", nil, "Set field value (format: Field=Value)")

	return cmd
}

func runCreate(ctx context.Context, opts *root.Options, objectName string, fields map[string]interface{}) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	result, err := client.CreateRecord(ctx, objectName, fields)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(result)
	}

	if result.Success {
		v.Success("Created %s record: %s", objectName, result.ID)
		v.Info("URL: %s", client.RecordURL(result.ID))
	} else {
		v.Error("Failed to create record")
		for _, e := range result.Errors {
			v.Error("  %s: %s", e.StatusCode, e.Message)
		}
	}

	return nil
}

// parseSetFlags parses --set flags into a map of field values
func parseSetFlags(flags []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, flag := range flags {
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --set format: %q (expected Field=Value)", flag)
		}

		fieldName := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		// Try to parse as boolean
		switch strings.ToLower(value) {
		case "true":
			result[fieldName] = true
		case "false":
			result[fieldName] = false
		case "null", "":
			result[fieldName] = nil
		default:
			result[fieldName] = value
		}
	}

	return result, nil
}
