package recordcmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newDeleteCommand(opts *root.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <object> <id>",
		Short: "Delete a record",
		Long: `Delete a Salesforce record.

Examples:
  sfdc record delete Account 001xx000003DGbYAAW --confirm
  sfdc record delete Contact 003xx000001abcd`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd.Context(), opts, args[0], args[1], confirm)
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(ctx context.Context, opts *root.Options, objectName, recordID string, confirm bool) error {
	v := opts.View()

	// Prompt for confirmation if not confirmed
	if !confirm {
		fmt.Printf("Delete %s record %s? [y/N]: ", objectName, recordID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			v.Info("Cancelled")
			return nil
		}
	}

	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	err = client.DeleteRecord(ctx, objectName, recordID)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	if opts.Output == "json" {
		return v.JSON(map[string]interface{}{
			"success": true,
			"id":      recordID,
			"object":  objectName,
			"deleted": true,
		})
	}

	v.Success("Deleted %s record: %s", objectName, recordID)
	return nil
}
