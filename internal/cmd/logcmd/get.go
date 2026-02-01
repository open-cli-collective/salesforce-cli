package logcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newGetCommand(opts *root.Options) *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "get <log-id>",
		Short: "Get debug log content",
		Long: `Get the content of a debug log.

Examples:
  sfdc log get 07L1x000000ABCD              # Display log content
  sfdc log get 07L1x000000ABCD --output debug.log  # Save to file`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogGet(cmd.Context(), opts, args[0], outputFile)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "f", "", "Output file path")

	return cmd
}

func runLogGet(ctx context.Context, opts *root.Options, logID, outputFile string) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	body, err := client.GetApexLogBody(ctx, logID)
	if err != nil {
		return fmt.Errorf("failed to get log content: %w", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(body), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		v := opts.View()
		v.Info("Saved log to %s", outputFile)
		return nil
	}

	fmt.Fprintln(opts.Stdout, body)
	return nil
}
