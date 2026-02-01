package apexcmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func newExecuteCommand(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "execute [code]",
		Short: "Execute anonymous Apex code",
		Long: `Execute anonymous Apex code.

The code can be provided as an argument, from a file, or via stdin.

Examples:
  sfdc apex execute "System.debug('Hello');"
  sfdc apex execute --file script.apex
  echo "System.debug(UserInfo.getUserName());" | sfdc apex execute -
  sfdc apex execute -                           # Read from stdin`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var code string

			if file != "" {
				// Read from file
				data, readErr := os.ReadFile(file)
				if readErr != nil {
					return fmt.Errorf("failed to read file: %w", readErr)
				}
				code = string(data)
			} else if len(args) == 1 {
				if args[0] == "-" {
					// Read from stdin
					data, err := io.ReadAll(opts.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read stdin: %w", err)
					}
					code = string(data)
				} else {
					code = args[0]
				}
			} else {
				return fmt.Errorf("code required: provide as argument, --file, or pipe to stdin")
			}

			code = strings.TrimSpace(code)
			if code == "" {
				return fmt.Errorf("empty code provided")
			}

			return runExecute(cmd.Context(), opts, code)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "File containing Apex code")

	return cmd
}

func runExecute(ctx context.Context, opts *root.Options, code string) error {
	client, err := opts.ToolingClient()
	if err != nil {
		return fmt.Errorf("failed to create tooling client: %w", err)
	}

	result, err := client.ExecuteAnonymous(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to execute anonymous apex: %w", err)
	}

	v := opts.View()

	if opts.Output == "json" {
		return v.JSON(result)
	}

	if !result.Compiled {
		v.Error("Compile error at line %d, column %d:", result.Line, result.Column)
		fmt.Fprintln(opts.Stderr, result.CompileProblem)
		return fmt.Errorf("compilation failed")
	}

	if !result.Success {
		v.Error("Runtime error:")
		fmt.Fprintln(opts.Stderr, result.ExceptionMessage)
		if result.ExceptionStackTrace != "" {
			fmt.Fprintln(opts.Stderr, result.ExceptionStackTrace)
		}
		return fmt.Errorf("execution failed")
	}

	v.Success("Executed successfully")
	return nil
}
