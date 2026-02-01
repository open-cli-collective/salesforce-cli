// Package searchcmd provides the search command for SOSL searches.
package searchcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Register registers the search command with the root command.
func Register(parent *cobra.Command, opts *root.Options) {
	parent.AddCommand(NewCommand(opts))
}

// NewCommand creates the search command.
func NewCommand(opts *root.Options) *cobra.Command {
	var (
		inObjects string
		returning string
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for records using SOSL",
		Long: `Search for records across multiple objects using Salesforce Object Search Language (SOSL).

Examples:
  sfdc search "Acme"
  sfdc search "John Smith" --in Account,Contact
  sfdc search "test" --returning "Account(Id,Name),Contact(Id,FirstName,LastName)"
  sfdc search "FIND {Acme} IN ALL FIELDS RETURNING Account(Id,Name)"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd.Context(), opts, args[0], inObjects, returning)
		},
	}

	cmd.Flags().StringVar(&inObjects, "in", "", "Limit search to specific objects (comma-separated)")
	cmd.Flags().StringVar(&returning, "returning", "", "Specify return fields per object (e.g., Account(Id,Name),Contact(Id,Email))")

	return cmd
}

func runSearch(ctx context.Context, opts *root.Options, query, inObjects, returning string) error {
	client, err := opts.APIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Build SOSL query
	sosl := buildSOSL(query, inObjects, returning)

	result, err := client.Search(ctx, sosl)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	return renderSearchResult(opts, result)
}

// buildSOSL constructs a SOSL query from the input parameters
func buildSOSL(query, inObjects, returning string) string {
	// If query already starts with FIND, use it as-is
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "FIND") {
		return query
	}

	// Build SOSL from simplified parameters
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("FIND {%s}", query))

	// Add IN clause if specified
	if inObjects != "" {
		sb.WriteString(" IN ALL FIELDS")
	}

	// Add RETURNING clause
	if returning != "" {
		sb.WriteString(fmt.Sprintf(" RETURNING %s", returning))
	} else if inObjects != "" {
		// Build RETURNING from --in objects
		objects := strings.Split(inObjects, ",")
		for i := range objects {
			objects[i] = strings.TrimSpace(objects[i])
		}
		sb.WriteString(fmt.Sprintf(" RETURNING %s", strings.Join(objects, ",")))
	}

	return sb.String()
}

func renderSearchResult(opts *root.Options, result *api.SearchResult) error {
	v := opts.View()

	if len(result.SearchRecords) == 0 {
		v.Info("No records found")
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(result)
	}

	// Group results by object type
	byType := make(map[string][]api.SearchRecord)
	for _, rec := range result.SearchRecords {
		objType := rec.Attributes.Type
		byType[objType] = append(byType[objType], rec)
	}

	// Display results grouped by type
	for objType, records := range byType {
		v.Info("%s (%d):", objType, len(records))

		for _, rec := range records {
			// Build display string from fields
			fields := make([]string, 0)
			for name, value := range rec.Fields {
				if value != nil {
					fields = append(fields, fmt.Sprintf("%s=%v", name, value))
				}
			}

			if len(fields) > 0 {
				v.Info("  %s: %s", rec.ID, strings.Join(fields, ", "))
			} else {
				v.Info("  %s", rec.ID)
			}
		}
		v.Info("")
	}

	v.Info("%d record(s) found", len(result.SearchRecords))
	return nil
}
