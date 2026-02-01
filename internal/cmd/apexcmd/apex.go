// Package apexcmd provides commands for Apex class operations.
package apexcmd

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the apex command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the apex command.
func NewCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apex",
		Short: "Apex class operations",
		Long: `Manage Apex classes, triggers, and execute anonymous Apex.

Examples:
  sfdc apex list                          # List all Apex classes
  sfdc apex list --triggers               # List all Apex triggers
  sfdc apex get MyController              # Get class source code
  sfdc apex execute "System.debug('Hi');" # Execute anonymous Apex
  sfdc apex test --class MyTest           # Run Apex tests`,
	}

	cmd.AddCommand(newListCommand(opts))
	cmd.AddCommand(newGetCommand(opts))
	cmd.AddCommand(newExecuteCommand(opts))
	cmd.AddCommand(newTestCommand(opts))

	return cmd
}
