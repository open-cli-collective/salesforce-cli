package metadatacmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api/metadata"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newDeployCommand(opts *root.Options) *cobra.Command {
	var (
		sourceDir string
		checkOnly bool
		testLevel string
		wait      bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy metadata to the org",
		Long: `Deploy metadata from a local directory to the org.

The source directory should be in the standard Salesforce metadata format
(e.g., containing package.xml and subdirectories for each metadata type).

For complex deployments, use the official Salesforce CLI (sf).

Examples:
  sfdc metadata deploy --source ./src
  sfdc metadata deploy --source ./src --check-only
  sfdc metadata deploy --source ./src --test-level RunLocalTests
  sfdc metadata deploy --source ./src --wait`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sourceDir == "" {
				return fmt.Errorf("--source is required")
			}
			return runDeploy(cmd.Context(), opts, sourceDir, checkOnly, testLevel, wait)
		},
	}

	cmd.Flags().StringVar(&sourceDir, "source", "", "Source directory (required)")
	cmd.Flags().BoolVar(&checkOnly, "check-only", false, "Validate without deploying")
	cmd.Flags().StringVar(&testLevel, "test-level", "", "Test level: NoTestRun, RunLocalTests, RunAllTestsInOrg")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for deployment to complete")

	return cmd
}

func runDeploy(ctx context.Context, opts *root.Options, sourceDir string, checkOnly bool, testLevel string, wait bool) error {
	client, err := opts.MetadataClient()
	if err != nil {
		return fmt.Errorf("failed to create metadata client: %w", err)
	}

	v := opts.View()

	// Create zip from source directory
	v.Info("Creating deployment package from %s...", sourceDir)
	zipData, err := metadata.CreateZipFromDirectory(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to create deployment package: %w", err)
	}

	// Configure deployment options
	deployOpts := metadata.DeployOptions{
		CheckOnly:       checkOnly,
		RollbackOnError: true,
		SinglePackage:   true,
	}
	if testLevel != "" {
		deployOpts.TestLevel = testLevel
	}

	// Start deployment
	action := "Deploying"
	if checkOnly {
		action = "Validating"
	}
	v.Info("%s to org...", action)

	result, err := client.Deploy(ctx, zipData, deployOpts)
	if err != nil {
		return fmt.Errorf("failed to start deployment: %w", err)
	}

	v.Info("Deployment ID: %s", result.ID)

	if !wait {
		v.Info("Deployment started. Use 'sfdc metadata deploy-status %s' to check status.", result.ID)
		return nil
	}

	// Poll for completion
	v.Info("Waiting for deployment to complete...")

	for {
		status, err := client.GetDeployStatus(ctx, result.ID, true)
		if err != nil {
			return fmt.Errorf("failed to get deployment status: %w", err)
		}

		if status.Done {
			return displayDeployResult(opts, status)
		}

		v.Info("Status: %s (%d/%d components)...",
			status.Status,
			status.NumberComponentsDeployed,
			status.NumberComponentsTotal)

		time.Sleep(3 * time.Second)
	}
}

func displayDeployResult(opts *root.Options, result *metadata.DeployResult) error {
	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(result)
	}

	// Summary
	if result.Success {
		v.Success("Deployment succeeded!")
	} else {
		v.Error("Deployment failed!")
	}

	v.Info("\nComponents: %d deployed, %d errors",
		result.NumberComponentsDeployed,
		result.NumberComponentErrors)

	if result.NumberTestsTotal > 0 {
		v.Info("Tests: %d completed, %d errors",
			result.NumberTestsCompleted,
			result.NumberTestErrors)
	}

	// Show component failures
	if result.DeployDetails != nil && len(result.DeployDetails.ComponentFailures) > 0 {
		v.Error("\nComponent failures:")
		for _, failure := range result.DeployDetails.ComponentFailures {
			fmt.Fprintf(opts.Stderr, "  %s.%s: %s\n",
				failure.ComponentType,
				failure.FullName,
				failure.Problem)
			if failure.LineNumber > 0 {
				fmt.Fprintf(opts.Stderr, "    at line %d, column %d\n",
					failure.LineNumber,
					failure.ColumnNumber)
			}
		}
	}

	// Show error message
	if result.ErrorMessage != "" {
		v.Error("\nError: %s", result.ErrorMessage)
	}

	if !result.Success {
		var parts []string
		if result.NumberComponentErrors > 0 {
			parts = append(parts, fmt.Sprintf("%d component error(s)", result.NumberComponentErrors))
		}
		if result.NumberTestErrors > 0 {
			parts = append(parts, fmt.Sprintf("%d test error(s)", result.NumberTestErrors))
		}
		return fmt.Errorf("deployment failed: %s", strings.Join(parts, ", "))
	}

	return nil
}
