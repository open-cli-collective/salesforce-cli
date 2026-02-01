package bulkcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newJobCommand(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage bulk jobs",
		Long: `Manage Salesforce Bulk API 2.0 jobs.

Examples:
  sfdc bulk job list
  sfdc bulk job status 750xx000000001
  sfdc bulk job results 750xx000000001
  sfdc bulk job errors 750xx000000001
  sfdc bulk job abort 750xx000000001`,
	}

	cmd.AddCommand(newJobListCommand(opts))
	cmd.AddCommand(newJobStatusCommand(opts))
	cmd.AddCommand(newJobResultsCommand(opts))
	cmd.AddCommand(newJobErrorsCommand(opts))
	cmd.AddCommand(newJobAbortCommand(opts))

	return cmd
}

func newJobListCommand(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List recent bulk jobs",
		Long: `List recent bulk ingest jobs.

Examples:
  sfdc bulk job list
  sfdc bulk job list -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobList(cmd.Context(), opts)
		},
	}
}

func runJobList(ctx context.Context, opts *root.Options) error {
	client, err := opts.BulkClient()
	if err != nil {
		return fmt.Errorf("failed to create bulk client: %w", err)
	}

	resp, err := client.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	v := opts.View()

	if len(resp.Records) == 0 {
		v.Info("No bulk jobs found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(resp)
	}

	headers := []string{"ID", "Object", "Operation", "State", "Processed", "Failed"}
	rows := make([][]string, 0, len(resp.Records))
	for _, job := range resp.Records {
		rows = append(rows, []string{
			job.ID,
			job.Object,
			string(job.Operation),
			string(job.State),
			fmt.Sprintf("%d", job.NumberRecordsProcessed),
			fmt.Sprintf("%d", job.NumberRecordsFailed),
		})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}

	v.Info("\n%d job(s)", len(resp.Records))
	return nil
}

func newJobStatusCommand(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status <job-id>",
		Short: "Get bulk job status",
		Long: `Get the status of a bulk ingest job.

Examples:
  sfdc bulk job status 750xx000000001
  sfdc bulk job status 750xx000000001 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobStatus(cmd.Context(), opts, args[0])
		},
	}
}

func runJobStatus(ctx context.Context, opts *root.Options, jobID string) error {
	client, err := opts.BulkClient()
	if err != nil {
		return fmt.Errorf("failed to create bulk client: %w", err)
	}

	job, err := client.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	return renderJobResult(opts, job)
}

func newJobResultsCommand(opts *root.Options) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "results <job-id>",
		Short: "Get successful results from a bulk job",
		Long: `Get the successful results from a completed bulk ingest job.

Examples:
  sfdc bulk job results 750xx000000001
  sfdc bulk job results 750xx000000001 --output results.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobResults(cmd.Context(), opts, args[0], output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")

	return cmd
}

func runJobResults(ctx context.Context, opts *root.Options, jobID, output string) error {
	client, err := opts.BulkClient()
	if err != nil {
		return fmt.Errorf("failed to create bulk client: %w", err)
	}

	data, err := client.GetSuccessfulResults(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get results: %w", err)
	}

	if output != "" {
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		opts.View().Info("Results written to %s", output)
		return nil
	}

	fmt.Fprintln(opts.Stdout, string(data))
	return nil
}

func newJobErrorsCommand(opts *root.Options) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "errors <job-id>",
		Short: "Get failed records from a bulk job",
		Long: `Get the failed records from a completed bulk ingest job.

Examples:
  sfdc bulk job errors 750xx000000001
  sfdc bulk job errors 750xx000000001 --output errors.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobErrors(cmd.Context(), opts, args[0], output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")

	return cmd
}

func runJobErrors(ctx context.Context, opts *root.Options, jobID, output string) error {
	client, err := opts.BulkClient()
	if err != nil {
		return fmt.Errorf("failed to create bulk client: %w", err)
	}

	data, err := client.GetFailedResults(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get failed results: %w", err)
	}

	if output != "" {
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		opts.View().Info("Errors written to %s", output)
		return nil
	}

	fmt.Fprintln(opts.Stdout, string(data))
	return nil
}

func newJobAbortCommand(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "abort <job-id>",
		Short: "Abort a bulk job",
		Long: `Abort a running bulk ingest job.

Examples:
  sfdc bulk job abort 750xx000000001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobAbort(cmd.Context(), opts, args[0])
		},
	}
}

func runJobAbort(ctx context.Context, opts *root.Options, jobID string) error {
	client, err := opts.BulkClient()
	if err != nil {
		return fmt.Errorf("failed to create bulk client: %w", err)
	}

	job, err := client.AbortJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to abort job: %w", err)
	}

	v := opts.View()
	v.Info("Job %s aborted", job.ID)
	return nil
}
