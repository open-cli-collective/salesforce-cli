// Package metadatacmd provides commands for metadata operations.
package metadatacmd

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the metadata command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the metadata command.
func NewCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Metadata operations",
		Long: `Manage Salesforce metadata.

This is a thin wrapper for basic metadata operations. For complex
workflows, use the official Salesforce CLI (sf).

Examples:
  sfdc metadata types                           # List metadata types
  sfdc metadata list --type ApexClass           # List Apex classes
  sfdc metadata retrieve --type ApexClass       # Retrieve all classes
  sfdc metadata deploy --source ./src           # Deploy from directory`,
	}

	cmd.AddCommand(newTypesCommand(opts))
	cmd.AddCommand(newListCommand(opts))
	cmd.AddCommand(newRetrieveCommand(opts))
	cmd.AddCommand(newDeployCommand(opts))

	return cmd
}
