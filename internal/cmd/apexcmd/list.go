package apexcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api/tooling"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newListCommand(opts *root.Options) *cobra.Command {
	var triggers bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Apex classes or triggers",
		Long: `List all Apex classes or triggers in the org.

Examples:
  sfdc apex list                 # List all Apex classes
  sfdc apex list --triggers      # List all Apex triggers
  sfdc apex list -o json         # Output as JSON`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), opts, triggers)
		},
	}

	cmd.Flags().BoolVar(&triggers, "triggers", false, "List triggers instead of classes")

	return cmd
}

func runList(ctx context.Context, opts *root.Options, triggers bool) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	v := opts.View()

	if triggers {
		return listTriggers(ctx, client, v, opts)
	}
	return listClasses(ctx, client, v, opts)
}

func listClasses(ctx context.Context, client *tooling.Client, v interface {
	Table([]string, [][]string) error
	JSON(interface{}) error
	Info(string, ...interface{})
}, opts *root.Options) error {
	classes, err := client.ListApexClasses(ctx)
	if err != nil {
		return fmt.Errorf("failed to list apex classes: %w", err)
	}

	if len(classes) == 0 {
		v.Info("No Apex classes found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(classes)
	}

	headers := []string{"ID", "Name", "Status", "Valid", "API Version", "Lines"}
	rows := make([][]string, 0, len(classes))
	for _, c := range classes {
		valid := "No"
		if c.IsValid {
			valid = "Yes"
		}
		rows = append(rows, []string{
			c.ID,
			c.Name,
			c.Status,
			valid,
			fmt.Sprintf("%.0f", c.APIVersion),
			fmt.Sprintf("%d", c.LengthWithoutComments),
		})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}
	v.Info("\n%d class(es)", len(classes))
	return nil
}

func listTriggers(ctx context.Context, client *tooling.Client, v interface {
	Table([]string, [][]string) error
	JSON(interface{}) error
	Info(string, ...interface{})
}, opts *root.Options) error {
	triggers, err := client.ListApexTriggers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list apex triggers: %w", err)
	}

	if len(triggers) == 0 {
		v.Info("No Apex triggers found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(triggers)
	}

	headers := []string{"ID", "Name", "Object", "Status", "Valid", "API Version"}
	rows := make([][]string, 0, len(triggers))
	for _, t := range triggers {
		valid := "No"
		if t.IsValid {
			valid = "Yes"
		}
		rows = append(rows, []string{
			t.ID,
			t.Name,
			t.TableEnumOrID,
			t.Status,
			valid,
			fmt.Sprintf("%.0f", t.APIVersion),
		})
	}

	if err := v.Table(headers, rows); err != nil {
		return err
	}
	v.Info("\n%d trigger(s)", len(triggers))
	return nil
}
