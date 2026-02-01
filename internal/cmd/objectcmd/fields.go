package objectcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newFieldsCommand(opts *root.Options) *cobra.Command {
	var requiredOnly bool

	cmd := &cobra.Command{
		Use:   "fields <object>",
		Short: "List fields for an object",
		Long: `List all fields for a Salesforce object.

Examples:
  sfdc object fields Account
  sfdc object fields Account --required-only
  sfdc object fields Contact -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFields(cmd.Context(), opts, args[0], requiredOnly)
		},
	}

	cmd.Flags().BoolVar(&requiredOnly, "required-only", false, "Show only required fields")

	return cmd
}

func runFields(ctx context.Context, opts *root.Options, objectName string, requiredOnly bool) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	desc, err := client.DescribeSObject(ctx, objectName)
	if err != nil {
		return fmt.Errorf("failed to describe object: %w", err)
	}

	v := opts.View()

	// Filter fields if needed
	fields := desc.Fields
	if requiredOnly {
		filtered := make([]api.Field, 0)
		for _, f := range fields {
			// Required = not nillable AND createable (can be set on create)
			if !f.Nillable && f.Createable {
				filtered = append(filtered, f)
			}
		}
		fields = filtered
	}

	if opts.Output == "json" {
		return v.JSON(fields)
	}

	if len(fields) == 0 {
		v.Info("No fields found")
		return nil
	}

	headers := []string{"Name", "Label", "Type", "Length", "Required", "Custom"}
	rows := make([][]string, 0, len(fields))

	for _, f := range fields {
		isRequired := !f.Nillable && f.Createable
		length := ""
		if f.Length > 0 {
			length = fmt.Sprintf("%d", f.Length)
		}

		rows = append(rows, []string{
			f.Name,
			f.Label,
			f.Type,
			length,
			boolToYesNo(isRequired),
			boolToYesNo(f.Custom),
		})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}

	v.Info("\n%d field(s)", len(fields))
	return nil
}
