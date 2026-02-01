package metadatacmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newListCommand(opts *root.Options) *cobra.Command {
	var metadataType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List components of a metadata type",
		Long: `List components of a specific metadata type.

Supported types:
  ApexClass, ApexTrigger, ApexPage, ApexComponent, StaticResource,
  AuraDefinitionBundle, LightningComponentBundle

Examples:
  sfdc metadata list --type ApexClass
  sfdc metadata list --type ApexTrigger
  sfdc metadata list --type ApexClass -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if metadataType == "" {
				return fmt.Errorf("--type is required")
			}
			return runList(cmd.Context(), opts, metadataType)
		},
	}

	cmd.Flags().StringVar(&metadataType, "type", "", "Metadata type (required)")

	return cmd
}

func runList(ctx context.Context, opts *root.Options, metadataType string) error {
	client, err := opts.MetadataClient()
	if err != nil {
		return fmt.Errorf("failed to create metadata client: %w", err)
	}

	components, err := client.ListMetadata(ctx, metadataType)
	if err != nil {
		return fmt.Errorf("failed to list metadata: %w", err)
	}

	v := opts.View()

	if len(components) == 0 {
		v.Info("No %s components found", metadataType)
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(components)
	}

	headers := []string{"ID", "Name", "Namespace"}
	rows := make([][]string, 0, len(components))
	for _, comp := range components {
		ns := comp.NamespacePrefix
		if ns == "" {
			ns = "-"
		}
		rows = append(rows, []string{
			comp.ID,
			comp.FullName,
			ns,
		})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}
	v.Info("\n%d component(s)", len(components))
	return nil
}
