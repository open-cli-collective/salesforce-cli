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

	// Filter if needed
	objects := resp.SObjects
	if customOnly {
		filtered := make([]struct {
			Name        string
			Label       string
			LabelPlural string
			KeyPrefix   string
			Custom      bool
			Queryable   bool
		}, 0)
		for _, obj := range objects {
			if obj.Custom {
				filtered = append(filtered, struct {
					Name        string
					Label       string
					LabelPlural string
					KeyPrefix   string
					Custom      bool
					Queryable   bool
				}{
					Name:        obj.Name,
					Label:       obj.Label,
					LabelPlural: obj.LabelPlural,
					KeyPrefix:   obj.KeyPrefix,
					Custom:      obj.Custom,
					Queryable:   obj.Queryable,
				})
			}
		}

		if opts.Output == "json" {
			return v.JSON(filtered)
		}

		headers := []string{"Name", "Label", "Key Prefix", "Queryable"}
		rows := make([][]string, 0, len(filtered))
		for _, obj := range filtered {
			queryable := "No"
			if obj.Queryable {
				queryable = "Yes"
			}
			rows = append(rows, []string{obj.Name, obj.Label, obj.KeyPrefix, queryable})
		}

		if len(rows) == 0 {
			v.Info("No custom objects found")
			return nil
		}

		return v.Table(headers, rows)
	}

	if opts.Output == "json" {
		return v.JSON(objects)
	}

	headers := []string{"Name", "Label", "Key Prefix", "Custom", "Queryable"}
	rows := make([][]string, 0, len(objects))
	for _, obj := range objects {
		custom := "No"
		if obj.Custom {
			custom = "Yes"
		}
		queryable := "No"
		if obj.Queryable {
			queryable = "Yes"
		}
		rows = append(rows, []string{obj.Name, obj.Label, obj.KeyPrefix, custom, queryable})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}

	v.Info("\n%d object(s)", len(objects))
	return nil
}
