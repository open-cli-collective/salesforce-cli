package logcmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newTailCommand(opts *root.Options) *cobra.Command {
	var (
		userID   string
		interval int
	)

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Stream new debug logs",
		Long: `Continuously poll for new debug logs and display them.

Press Ctrl+C to stop.

Examples:
  sfdc log tail                     # Stream all new logs
  sfdc log tail --user 005xxx       # Filter by user ID
  sfdc log tail --interval 5        # Poll every 5 seconds`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogTail(cmd.Context(), opts, userID, interval)
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "Filter by user ID")
	cmd.Flags().IntVar(&interval, "interval", 3, "Polling interval in seconds")

	return cmd
}

func runLogTail(ctx context.Context, opts *root.Options, userID string, interval int) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	v := opts.View()
	v.Info("Tailing debug logs... (Ctrl+C to stop)")

	seenLogs := make(map[string]bool)

	// Get initial logs to establish baseline
	initialLogs, err := client.ListApexLogs(ctx, userID, 10)
	if err != nil {
		return fmt.Errorf("failed to get initial logs: %w", err)
	}
	for _, log := range initialLogs {
		seenLogs[log.ID] = true
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			v.Info("\nStopped")
			return nil
		case <-ticker.C:
			logs, err := client.ListApexLogs(ctx, userID, 10)
			if err != nil {
				v.Error("Failed to poll logs: %v", err)
				continue
			}

			// Process new logs (in reverse order to show oldest first)
			for i := len(logs) - 1; i >= 0; i-- {
				log := logs[i]
				if seenLogs[log.ID] {
					continue
				}
				seenLogs[log.ID] = true

				// Print log summary
				fmt.Fprintf(opts.Stdout, "\n[%s] %s (%s, %s)\n",
					log.StartTime.Format("15:04:05"),
					log.Operation,
					log.Status,
					formatSize(log.LogLength),
				)

				// Fetch and print log body
				body, err := client.GetApexLogBody(ctx, log.ID)
				if err != nil {
					v.Error("Failed to get log body: %v", err)
					continue
				}
				fmt.Fprintln(opts.Stdout, body)
				fmt.Fprintln(opts.Stdout, "---")
			}
		}
	}
}
