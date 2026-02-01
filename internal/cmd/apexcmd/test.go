package apexcmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api/tooling"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newTestCommand(opts *root.Options) *cobra.Command {
	var (
		className  string
		methodName string
		wait       bool
	)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run Apex tests",
		Long: `Run Apex tests asynchronously.

Examples:
  sfdc apex test --class MyControllerTest
  sfdc apex test --class MyControllerTest --method testCreate
  sfdc apex test --class MyTest --wait
  sfdc apex test --class MyTest -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if className == "" {
				return fmt.Errorf("--class is required")
			}
			return runTest(cmd.Context(), opts, className, methodName, wait)
		},
	}

	cmd.Flags().StringVar(&className, "class", "", "Test class name (required)")
	cmd.Flags().StringVar(&methodName, "method", "", "Specific test method to run")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for tests to complete")

	return cmd
}

func runTest(ctx context.Context, opts *root.Options, className, methodName string, wait bool) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	v := opts.View()

	// Get the class ID for the test class
	classID, err := client.GetApexClassID(ctx, className)
	if err != nil {
		return fmt.Errorf("failed to find test class: %w", err)
	}

	v.Info("Running tests for %s...", className)

	// Enqueue the test run
	jobID, err := client.RunTestsAsync(ctx, []string{classID})
	if err != nil {
		return fmt.Errorf("failed to enqueue tests: %w", err)
	}

	v.Info("Test job ID: %s", jobID)

	if !wait {
		v.Info("Tests enqueued. Use 'sfdc apex test-status %s' to check results.", jobID)
		return nil
	}

	// Poll for completion
	v.Info("Waiting for tests to complete...")

	for {
		job, err := client.GetAsyncJobStatus(ctx, jobID)
		if err != nil {
			return fmt.Errorf("failed to get job status: %w", err)
		}

		switch job.Status {
		case "Completed", "Aborted", "Failed":
			return displayTestResults(ctx, client, opts, jobID, methodName)
		case "Queued", "Processing", "Preparing", "Holding":
			time.Sleep(2 * time.Second)
		default:
			return fmt.Errorf("unexpected job status: %s", job.Status)
		}
	}
}

func displayTestResults(ctx context.Context, client *tooling.Client, opts *root.Options, jobID, filterMethod string) error {
	results, err := client.GetTestResults(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get test results: %w", err)
	}

	// Filter by method if specified
	if filterMethod != "" {
		filtered := make([]tooling.ApexTestResult, 0)
		for _, r := range results {
			if r.MethodName == filterMethod {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	v := opts.View()

	if len(results) == 0 {
		v.Info("No test results found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(results)
	}

	headers := []string{"Class", "Method", "Outcome", "Time (ms)", "Message"}
	rows := make([][]string, 0, len(results))

	passCount := 0
	failCount := 0
	totalTime := 0

	for _, r := range results {
		message := r.Message
		if len(message) > 50 {
			message = message[:50] + "..."
		}

		rows = append(rows, []string{
			r.ClassName,
			r.MethodName,
			r.Outcome,
			fmt.Sprintf("%d", r.RunTime),
			message,
		})

		totalTime += r.RunTime
		switch r.Outcome {
		case "Pass":
			passCount++
		case "Fail", "CompileFail":
			failCount++
		}
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}

	// Summary
	var summaryParts []string
	summaryParts = append(summaryParts, fmt.Sprintf("%d passed", passCount))
	if failCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d failed", failCount))
	}
	summaryParts = append(summaryParts, fmt.Sprintf("%dms total", totalTime))

	v.Info("\n%s", strings.Join(summaryParts, ", "))

	// Show failure details
	for _, r := range results {
		if r.Outcome == "Fail" || r.Outcome == "CompileFail" {
			fmt.Fprintf(opts.Stderr, "\n%s.%s:\n", r.ClassName, r.MethodName)
			fmt.Fprintf(opts.Stderr, "  %s\n", r.Message)
			if r.StackTrace != "" {
				fmt.Fprintf(opts.Stderr, "  Stack trace:\n")
				for _, line := range strings.Split(r.StackTrace, "\n") {
					fmt.Fprintf(opts.Stderr, "    %s\n", line)
				}
			}
		}
	}

	if failCount > 0 {
		return fmt.Errorf("%d test(s) failed", failCount)
	}

	return nil
}
