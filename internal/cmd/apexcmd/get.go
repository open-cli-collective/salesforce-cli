package apexcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newGetCommand(opts *root.Options) *cobra.Command {
	var (
		outputFile string
		trigger    bool
	)

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get Apex class or trigger source code",
		Long: `Get the source code of an Apex class or trigger.

Examples:
  sfdc apex get MyController                    # Display class source
  sfdc apex get MyController --output My.cls    # Save to file
  sfdc apex get MyTrigger --trigger             # Get trigger source`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd.Context(), opts, args[0], outputFile, trigger)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "f", "", "Output file path")
	cmd.Flags().BoolVar(&trigger, "trigger", false, "Get trigger instead of class")

	return cmd
}

func runGet(ctx context.Context, opts *root.Options, name, outputFile string, trigger bool) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	var body string
	var typeName string

	if trigger {
		typeName = "trigger"
		t, err := client.GetApexTrigger(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to get apex trigger: %w", err)
		}
		body = t.Body
	} else {
		typeName = "class"
		c, err := client.GetApexClass(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to get apex class: %w", err)
		}
		body = c.Body
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(body), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		v := opts.View()
		v.Info("Saved %s %s to %s", typeName, name, outputFile)
		return nil
	}

	// Output directly to stdout
	fmt.Fprintln(opts.Stdout, body)
	return nil
}
