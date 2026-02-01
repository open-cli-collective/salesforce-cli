// Package limitscmd provides the limits command for viewing org API limits.
package limitscmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the limits command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the limits command.
func NewCommand(opts *root.Options) *cobra.Command {
	var show string

	cmd := &cobra.Command{
		Use:   "limits",
		Short: "Display org API limits",
		Long: `Display the current Salesforce org's API limits and usage.

Examples:
  sfdc limits
  sfdc limits -o json
  sfdc limits --show DailyApiRequests`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLimits(cmd.Context(), opts, show)
		},
	}

	cmd.Flags().StringVar(&show, "show", "", "Show only a specific limit by name")

	return cmd
}

func runLimits(ctx context.Context, opts *root.Options, show string) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	limits, err := client.GetLimits(ctx)
	if err != nil {
		return fmt.Errorf("failed to get limits: %w", err)
	}

	// If showing a specific limit
	if show != "" {
		return renderSingleLimit(opts, limits, show)
	}

	return renderLimits(opts, limits)
}

func renderSingleLimit(opts *root.Options, limits api.Limits, name string) error {
	v := opts.View()

	limit, ok := limits[name]
	if !ok {
		return fmt.Errorf("limit %q not found", name)
	}

	if opts.Output == "json" {
		return v.JSON(map[string]interface{}{
			"name":      name,
			"max":       limit.Max,
			"remaining": limit.Remaining,
			"used":      limit.Max - limit.Remaining,
		})
	}

	used := limit.Max - limit.Remaining
	pct := float64(0)
	if limit.Max > 0 {
		pct = float64(used) / float64(limit.Max) * 100
	}

	v.Info("%s", name)
	v.Info("  Max:       %d", limit.Max)
	v.Info("  Remaining: %d", limit.Remaining)
	v.Info("  Used:      %d (%.1f%%)", used, pct)

	return nil
}

func renderLimits(opts *root.Options, limits api.Limits) error {
	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(limits)
	}

	// Sort limit names for consistent output
	names := make([]string, 0, len(limits))
	for name := range limits {
		names = append(names, name)
	}
	sort.Strings(names)

	headers := []string{"Limit", "Max", "Remaining", "Used", "Usage %"}
	rows := make([][]string, 0, len(names))

	for _, name := range names {
		limit := limits[name]
		used := limit.Max - limit.Remaining
		pct := float64(0)
		if limit.Max > 0 {
			pct = float64(used) / float64(limit.Max) * 100
		}

		rows = append(rows, []string{
			name,
			fmt.Sprintf("%d", limit.Max),
			fmt.Sprintf("%d", limit.Remaining),
			fmt.Sprintf("%d", used),
			fmt.Sprintf("%.1f%%", pct),
		})
	}

	return v.Table(headers, rows)
}
