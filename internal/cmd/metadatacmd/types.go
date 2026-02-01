package metadatacmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newTypesCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "List available metadata types",
		Long: `List metadata types available in the org.

Examples:
  sfdc metadata types
  sfdc metadata types -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTypes(cmd.Context(), opts)
		},
	}

	return cmd
}

func runTypes(ctx context.Context, opts *root.Options) error {
	client, err := opts.MetadataClient()
	if err != nil {
		return fmt.Errorf("failed to create metadata client: %w", err)
	}

	result, err := client.DescribeMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to describe metadata: %w", err)
	}

	v := opts.View()

	if len(result.MetadataObjects) == 0 {
		v.Info("No metadata types found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(result.MetadataObjects)
	}

	// Sort by name for consistent output
	sort.Slice(result.MetadataObjects, func(i, j int) bool {
		return result.MetadataObjects[i].XMLName < result.MetadataObjects[j].XMLName
	})

	headers := []string{"Type Name"}
	rows := make([][]string, 0, len(result.MetadataObjects))
	for _, mt := range result.MetadataObjects {
		rows = append(rows, []string{mt.XMLName})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}
	v.Info("\n%d type(s)", len(result.MetadataObjects))
	return nil
}
