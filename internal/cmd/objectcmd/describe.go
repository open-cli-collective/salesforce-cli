package objectcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newDescribeCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <object>",
		Short: "Describe an object's metadata",
		Long: `Display detailed metadata about a Salesforce object.

Examples:
  sfdc object describe Account
  sfdc object describe Account -o json
  sfdc object describe MyCustomObject__c`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDescribe(cmd.Context(), opts, args[0])
		},
	}

	return cmd
}

func runDescribe(ctx context.Context, opts *root.Options, objectName string) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	desc, err := client.DescribeSObject(ctx, objectName)
	if err != nil {
		return fmt.Errorf("failed to describe object: %w", err)
	}

	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(desc)
	}

	// Display object info
	v.Info("Object: %s", desc.Name)
	v.Info("Label: %s (%s)", desc.Label, desc.LabelPlural)
	if desc.KeyPrefix != "" {
		v.Info("Key Prefix: %s", desc.KeyPrefix)
	}
	v.Info("")

	// Display capabilities
	v.Info("Capabilities:")
	v.Info("  Custom:     %v", desc.Custom)
	v.Info("  Createable: %v", desc.Createable)
	v.Info("  Updateable: %v", desc.Updateable)
	v.Info("  Deletable:  %v", desc.Deletable)
	v.Info("  Queryable:  %v", desc.Queryable)
	v.Info("  Searchable: %v", desc.Searchable)
	v.Info("")

	// Display field count
	v.Info("Fields: %d (use 'sfdc object fields %s' to list)", len(desc.Fields), objectName)

	return nil
}
