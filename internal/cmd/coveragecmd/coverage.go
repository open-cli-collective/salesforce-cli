// Package coveragecmd provides commands for code coverage operations.
package coveragecmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the coverage command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the coverage command.
func NewCommand(opts *root.Options) *cobra.Command {
	var (
		className string
		minCover  int
	)

	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "Show code coverage",
		Long: `Show Apex code coverage for the org.

Examples:
  sfdc coverage                       # Show all coverage
  sfdc coverage --class MyController  # Show coverage for specific class
  sfdc coverage --min 75              # Fail if overall coverage < 75%
  sfdc coverage -o json               # Output as JSON`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCoverage(cmd.Context(), opts, className, minCover)
		},
	}

	cmd.Flags().StringVar(&className, "class", "", "Show coverage for specific class")
	cmd.Flags().IntVar(&minCover, "min", 0, "Minimum coverage percentage (exit 1 if below)")

	return cmd
}

func runCoverage(ctx context.Context, opts *root.Options, className string, minCover int) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	v := opts.View()

	// If specific class requested
	if className != "" {
		cov, err := client.GetCodeCoverageForClass(ctx, className)
		if err != nil {
			return fmt.Errorf("failed to get coverage: %w", err)
		}

		if opts.Output == "json" {
			return v.JSON(cov)
		}

		total := cov.NumLinesCovered + cov.NumLinesUncovered
		pct := 0.0
		if total > 0 {
			pct = float64(cov.NumLinesCovered) / float64(total) * 100
		}

		headers := []string{"Class", "Lines Covered", "Lines Uncovered", "Coverage %"}
		rows := [][]string{
			{
				cov.ApexClassOrTrigger.Name,
				fmt.Sprintf("%d", cov.NumLinesCovered),
				fmt.Sprintf("%d", cov.NumLinesUncovered),
				fmt.Sprintf("%.1f%%", pct),
			},
		}

		if err := v.Table(headers, rows); err != nil {
			return err
		}

		if minCover > 0 && int(pct) < minCover {
			return fmt.Errorf("coverage %.1f%% is below minimum %d%%", pct, minCover)
		}

		return nil
	}

	// Get all coverage
	coverage, err := client.GetCodeCoverage(ctx)
	if err != nil {
		return fmt.Errorf("failed to get coverage: %w", err)
	}

	if len(coverage) == 0 {
		v.Info("No code coverage data found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(coverage)
	}

	headers := []string{"Class/Trigger", "Lines Covered", "Lines Uncovered", "Coverage %"}
	rows := make([][]string, 0, len(coverage))

	totalCovered := 0
	totalUncovered := 0

	for _, cov := range coverage {
		total := cov.NumLinesCovered + cov.NumLinesUncovered
		pct := 0.0
		if total > 0 {
			pct = float64(cov.NumLinesCovered) / float64(total) * 100
		}

		totalCovered += cov.NumLinesCovered
		totalUncovered += cov.NumLinesUncovered

		rows = append(rows, []string{
			cov.ApexClassOrTrigger.Name,
			fmt.Sprintf("%d", cov.NumLinesCovered),
			fmt.Sprintf("%d", cov.NumLinesUncovered),
			fmt.Sprintf("%.1f%%", pct),
		})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}

	// Calculate and display overall coverage
	overallTotal := totalCovered + totalUncovered
	overallPct := 0.0
	if overallTotal > 0 {
		overallPct = float64(totalCovered) / float64(overallTotal) * 100
	}

	v.Info("\nOverall: %d/%d lines covered (%.1f%%)", totalCovered, overallTotal, overallPct)

	if minCover > 0 && int(overallPct) < minCover {
		return fmt.Errorf("overall coverage %.1f%% is below minimum %d%%", overallPct, minCover)
	}

	return nil
}
