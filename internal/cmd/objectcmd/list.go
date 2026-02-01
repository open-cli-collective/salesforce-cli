package objectcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newListCommand(opts *root.Options) *cobra.Command {
	var customOnly bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all objects in the org",
		Long: `List all Salesforce objects (SObjects) in the org.

Examples:
  sfdc object list
  sfdc object list --custom-only
  sfdc object list -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), opts, customOnly)
		},
	}

	cmd.Flags().BoolVar(&customOnly, "custom-only", false, "Show only custom objects")

	return cmd
}

func runList(ctx context.Context, opts *root.Options, customOnly bool) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	resp, err := client.GetSObjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get objects: %w", err)
	}

	v := opts.View()

	// Filter custom objects if requested
	objects := resp.SObjects
	if customOnly {
		filtered := objects[:0]
		for _, obj := range objects {
			if obj.Custom {
				filtered = append(filtered, obj)
			}
		}
		objects = filtered

		if len(objects) == 0 {
			v.Info("No custom objects found")
			return nil
		}
	}

	if opts.Output == "json" {
		return v.JSON(objects)
	}

	// Build table rows
	var headers []string
	if customOnly {
		headers = []string{"Name", "Label", "Key Prefix", "Queryable"}
	} else {
		headers = []string{"Name", "Label", "Key Prefix", "Custom", "Queryable"}
	}

	rows := make([][]string, 0, len(objects))
	for _, obj := range objects {
		queryable := boolToYesNo(obj.Queryable)
		if customOnly {
			rows = append(rows, []string{obj.Name, obj.Label, obj.KeyPrefix, queryable})
		} else {
			custom := boolToYesNo(obj.Custom)
			rows = append(rows, []string{obj.Name, obj.Label, obj.KeyPrefix, custom, queryable})
		}
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}

	v.Info("\n%d object(s)", len(objects))
	return nil
}

// boolToYesNo converts a boolean to "Yes" or "No" for display.
func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
