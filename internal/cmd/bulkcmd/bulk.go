// Package bulkcmd provides commands for Salesforce Bulk API 2.0 operations.
package bulkcmd

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the bulk command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk API 2.0 operations for large data import/export",
		Long: `Bulk API 2.0 commands for handling large data operations.

Use bulk operations when working with thousands or millions of records.
For smaller datasets, use the record command instead.

Examples:
  sfdc bulk import Account --file accounts.csv --operation insert
  sfdc bulk export "SELECT Id, Name FROM Account" --output accounts.csv
  sfdc bulk job list
  sfdc bulk job status 750xx000000001`,
	}

	cmd.AddCommand(newImportCommand(opts))
	cmd.AddCommand(newExportCommand(opts))
	cmd.AddCommand(newJobCommand(opts))

	parent.AddCommand(cmd)
}
