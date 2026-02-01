// Package recordcmd provides commands for working with Salesforce records.
package recordcmd

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the record command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the record command with subcommands.
func NewCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Work with Salesforce records",
		Long:  "Get, create, update, and delete Salesforce records.",
	}

	cmd.AddCommand(newGetCommand(opts))
	cmd.AddCommand(newCreateCommand(opts))
	cmd.AddCommand(newUpdateCommand(opts))
	cmd.AddCommand(newDeleteCommand(opts))

	return cmd
}
