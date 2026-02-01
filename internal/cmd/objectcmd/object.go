// Package objectcmd provides commands for working with Salesforce objects.
package objectcmd

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the object command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the object command with subcommands.
func NewCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "object",
		Short: "Work with Salesforce objects",
		Long:  "List, describe, and inspect Salesforce objects and their fields.",
	}

	cmd.AddCommand(newListCommand(opts))
	cmd.AddCommand(newDescribeCommand(opts))
	cmd.AddCommand(newFieldsCommand(opts))

	return cmd
}
