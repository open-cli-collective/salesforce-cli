// Package logcmd provides commands for debug log operations.
package logcmd

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the log command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the log command.
func NewCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Debug log operations",
		Long: `Manage and view Salesforce debug logs.

Examples:
  sfdc log list                     # List recent debug logs
  sfdc log list --limit 20          # List last 20 logs
  sfdc log get 07L1x000000ABCD      # Get log content
  sfdc log tail                     # Stream new logs`,
	}

	cmd.AddCommand(newListCommand(opts))
	cmd.AddCommand(newGetCommand(opts))
	cmd.AddCommand(newTailCommand(opts))

	return cmd
}
