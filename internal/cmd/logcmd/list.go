package logcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newListCommand(opts *root.Options) *cobra.Command {
	var (
		userID string
		limit  int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List debug logs",
		Long: `List debug logs from the org.

Examples:
  sfdc log list                        # List recent logs
  sfdc log list --limit 20             # List last 20 logs
  sfdc log list --user 005xxx          # Filter by user ID
  sfdc log list -o json                # Output as JSON`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogList(cmd.Context(), opts, userID, limit)
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "Filter by user ID")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of logs to return")

	return cmd
}

func runLogList(ctx context.Context, opts *root.Options, userID string, limit int) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	logs, err := client.ListApexLogs(ctx, userID, limit)
	if err != nil {
		return fmt.Errorf("failed to list logs: %w", err)
	}

	v := opts.View()

	if len(logs) == 0 {
		v.Info("No debug logs found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(logs)
	}

	headers := []string{"ID", "Operation", "Status", "Size", "Duration", "Start Time"}
	rows := make([][]string, 0, len(logs))
	for _, log := range logs {
		rows = append(rows, []string{
			log.ID,
			truncate(log.Operation, 30),
			log.Status,
			formatSize(log.LogLength),
			fmt.Sprintf("%dms", log.DurationMS),
			log.StartTime.Format("2006-01-02 15:04:05"),
		})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}
	v.Info("\n%d log(s)", len(logs))
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
