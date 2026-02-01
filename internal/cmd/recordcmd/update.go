package recordcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newUpdateCommand(opts *root.Options) *cobra.Command {
	var setFlags []string

	cmd := &cobra.Command{
		Use:   "update <object> <id>",
		Short: "Update an existing record",
		Long: `Update an existing Salesforce record.

Examples:
  sfdc record update Account 001xx000003DGbYAAW --set Name="New Name"
  sfdc record update Contact 003xx000001abcd --set Phone="555-1234" --set Email=new@example.com`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := parseSetFlags(setFlags)
			if err != nil {
				return err
			}
			if len(fields) == 0 {
				return fmt.Errorf("at least one --set flag is required")
			}
			return runUpdate(cmd.Context(), opts, args[0], args[1], fields)
		},
	}

	cmd.Flags().StringArrayVar(&setFlags, "set", nil, "Set field value (format: Field=Value)")

	return cmd
}

func runUpdate(ctx context.Context, opts *root.Options, objectName, recordID string, fields map[string]interface{}) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	err = client.UpdateRecord(ctx, objectName, recordID, fields)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(map[string]interface{}{
			"success": true,
			"id":      recordID,
			"object":  objectName,
		})
	}

	v.Success("Updated %s record: %s", objectName, recordID)
	return nil
}
