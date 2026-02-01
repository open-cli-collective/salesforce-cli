package bulkcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api/bulk"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newExportCommand(opts *root.Options) *cobra.Command {
	var (
		output string
	)

	cmd := &cobra.Command{
		Use:   "export <soql>",
		Short: "Export data using a bulk query",
		Long: `Export data from Salesforce using Bulk API 2.0 query.

Use this for exporting large datasets. For smaller queries, use the query command.

Examples:
  sfdc bulk export "SELECT Id, Name, Industry FROM Account"
  sfdc bulk export "SELECT Id, Name FROM Account" --output accounts.csv
  sfdc bulk export "SELECT * FROM Contact" --output contacts.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(cmd.Context(), opts, args[0], output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (prints to stdout if not specified)")

	return cmd
}

func runExport(ctx context.Context, opts *root.Options, soql, output string) error {
	client, err := opts.BulkClient()
	if err != nil {
		return fmt.Errorf("failed to create bulk client: %w", err)
	}

	v := opts.View()

	// Create query job
	v.Info("Creating bulk query job...")
	job, err := client.CreateQueryJob(ctx, bulk.QueryConfig{
		Query: soql,
	})
	if err != nil {
		return fmt.Errorf("failed to create query job: %w", err)
	}

	v.Info("Job created: %s", job.ID)

	// Poll until complete
	v.Info("Waiting for query to complete...")
	job, err = client.PollQueryJob(ctx, job.ID, bulk.DefaultPollConfig())
	if err != nil {
		return fmt.Errorf("failed waiting for query job: %w", err)
	}

	if job.State != bulk.StateJobComplete {
		return fmt.Errorf("query job failed with state: %s", job.State)
	}

	v.Info("Query completed. Records: %d", job.NumberRecordsProcessed)

	// Get results
	v.Info("Downloading results...")
	data, err := client.GetQueryResults(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to get query results: %w", err)
	}

	// Write to file or stdout
	if output != "" {
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		v.Info("Results written to %s", output)
	} else {
		fmt.Fprintln(opts.Stdout, string(data))
	}

	return nil
}
