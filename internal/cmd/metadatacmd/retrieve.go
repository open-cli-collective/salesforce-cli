package metadatacmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newRetrieveCommand(opts *root.Options) *cobra.Command {
	var (
		metadataType string
		name         string
		outputDir    string
	)

	cmd := &cobra.Command{
		Use:   "retrieve",
		Short: "Retrieve metadata from the org",
		Long: `Retrieve metadata components from the org.

Supported types for direct retrieve:
  ApexClass, ApexTrigger, ApexPage, ApexComponent

For complex retrieves with package.xml, use the official Salesforce CLI (sf).

Examples:
  sfdc metadata retrieve --type ApexClass --name MyController --output ./src
  sfdc metadata retrieve --type ApexClass --output ./src  # all classes
  sfdc metadata retrieve --type ApexTrigger --output ./src`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if metadataType == "" {
				return fmt.Errorf("--type is required")
			}
			if outputDir == "" {
				return fmt.Errorf("--output is required")
			}
			return runRetrieve(cmd.Context(), opts, metadataType, name, outputDir)
		},
	}

	cmd.Flags().StringVar(&metadataType, "type", "", "Metadata type (required)")
	cmd.Flags().StringVar(&name, "name", "", "Component name (optional, retrieves all if not specified)")
	cmd.Flags().StringVarP(&outputDir, "output", "f", "", "Output directory (required)")

	return cmd
}

func runRetrieve(ctx context.Context, opts *root.Options, metadataType, name, outputDir string) error {
	client, err := opts.MetadataClient()
	if err != nil {
		return fmt.Errorf("failed to create metadata client: %w", err)
	}

	v := opts.View()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get file extension for the metadata type
	ext := getFileExtension(metadataType)

	if name != "" {
		// Retrieve single component
		v.Info("Retrieving %s: %s", metadataType, name)

		content, err := client.Retrieve(ctx, metadataType, name)
		if err != nil {
			return fmt.Errorf("failed to retrieve: %w", err)
		}

		filename := filepath.Join(outputDir, name+ext)
		if err := os.WriteFile(filename, content, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		v.Success("Retrieved to %s", filename)
		return nil
	}

	// Retrieve all components
	v.Info("Retrieving all %s components...", metadataType)

	components, err := client.RetrieveAll(ctx, metadataType)
	if err != nil {
		return fmt.Errorf("failed to retrieve: %w", err)
	}

	if len(components) == 0 {
		v.Info("No %s components found to retrieve", metadataType)
		return nil
	}

	for compName, content := range components {
		filename := filepath.Join(outputDir, compName+ext)
		if err := os.WriteFile(filename, content, 0644); err != nil {
			v.Error("Failed to write %s: %v", compName, err)
			continue
		}
	}

	v.Success("Retrieved %d component(s) to %s", len(components), outputDir)
	return nil
}

func getFileExtension(metadataType string) string {
	switch metadataType {
	case "ApexClass":
		return ".cls"
	case "ApexTrigger":
		return ".trigger"
	case "ApexPage":
		return ".page"
	case "ApexComponent":
		return ".component"
	default:
		return ".txt"
	}
}
