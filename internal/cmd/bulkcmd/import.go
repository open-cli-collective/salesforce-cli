package bulkcmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api/bulk"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newImportCommand(opts *root.Options) *cobra.Command {
	var (
		file       string
		operation  string
		externalID string
		wait       bool
	)

	cmd := &cobra.Command{
		Use:   "import <object>",
		Short: "Import data from a CSV file using Bulk API 2.0",
		Long: `Import data from a CSV file into Salesforce using Bulk API 2.0.

The CSV file must have a header row with field names matching the Salesforce object.

Operations:
  insert  - Create new records
  update  - Update existing records (requires Id column)
  upsert  - Insert or update based on external ID field
  delete  - Delete records (requires Id column)

Examples:
  sfdc bulk import Account --file accounts.csv --operation insert
  sfdc bulk import Contact --file contacts.csv --operation upsert --external-id Email
  sfdc bulk import Account --file accounts.csv --operation update --wait
  sfdc bulk import Account --file delete-ids.csv --operation delete`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(cmd.Context(), opts, args[0], file, operation, externalID, wait)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to CSV file (required)")
	cmd.Flags().StringVar(&operation, "operation", "insert", "Operation: insert, update, upsert, delete")
	cmd.Flags().StringVar(&externalID, "external-id", "", "External ID field for upsert operation")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for job to complete")

	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runImport(ctx context.Context, opts *root.Options, object, file, operation, externalID string, wait bool) error {
	// Validate operation
	op := bulk.Operation(strings.ToLower(operation))
	switch op {
	case bulk.OperationInsert, bulk.OperationUpdate, bulk.OperationUpsert, bulk.OperationDelete:
		// Valid
	default:
		return fmt.Errorf("invalid operation: %s (must be insert, update, upsert, or delete)", operation)
	}

	// Upsert requires external ID
	if op == bulk.OperationUpsert && externalID == "" {
		return fmt.Errorf("--external-id is required for upsert operation")
	}

	// Read CSV file
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create bulk client
	client, err := opts.BulkClient()
	if err != nil {
		return fmt.Errorf("failed to create bulk client: %w", err)
	}

	v := opts.View()

	// Create job
	v.Info("Creating bulk %s job for %s...", operation, object)
	job, err := client.CreateJob(ctx, bulk.JobConfig{
		Object:     object,
		Operation:  op,
		ExternalID: externalID,
	})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	v.Info("Job created: %s", job.ID)

	// Upload data
	v.Info("Uploading data...")
	if err := client.UploadJobData(ctx, job.ID, data); err != nil {
		return fmt.Errorf("failed to upload data: %w", err)
	}

	// Close job to start processing
	v.Info("Starting job processing...")
	job, err = client.CloseJob(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to close job: %w", err)
	}

	if !wait {
		v.Info("Job %s is processing. Use 'sfdc bulk job status %s' to check progress.", job.ID, job.ID)
		return nil
	}

	// Poll until complete
	v.Info("Waiting for job to complete...")
	job, err = client.PollJob(ctx, job.ID, bulk.DefaultPollConfig())
	if err != nil {
		return fmt.Errorf("failed waiting for job: %w", err)
	}

	// Show results
	return renderJobResult(opts, job)
}

func renderJobResult(opts *root.Options, job *bulk.JobInfo) error {
	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(job)
	}

	v.Info("Job completed:")
	v.Info("  ID:                %s", job.ID)
	v.Info("  State:             %s", job.State)
	v.Info("  Records Processed: %d", job.NumberRecordsProcessed)
	v.Info("  Records Failed:    %d", job.NumberRecordsFailed)

	if job.NumberRecordsFailed > 0 {
		v.Info("\nUse 'sfdc bulk job errors %s' to see failed records.", job.ID)
	}

	return nil
}
