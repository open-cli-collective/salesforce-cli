// Package root provides the root command and global options.
package root

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api"
	"github.com/open-cli-collective/salesforce-cli/internal/auth"
	"github.com/open-cli-collective/salesforce-cli/internal/config"
	"github.com/open-cli-collective/salesforce-cli/internal/version"
	"github.com/open-cli-collective/salesforce-cli/internal/view"
)

// Options contains global options for commands
type Options struct {
	Output     string
	NoColor    bool
	Verbose    bool
	APIVersion string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer

	// testClient is used for testing; if set, APIClient() returns this instead
	testClient *api.Client
}

// View returns a configured View instance
func (o *Options) View() *view.View {
	v := view.NewWithFormat(o.Output, o.NoColor)
	v.Out = o.Stdout
	v.Err = o.Stderr
	return v
}

// APIClient creates a new API client from config
func (o *Options) APIClient() (*api.Client, error) {
	if o.testClient != nil {
		return o.testClient, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	httpClient, err := auth.GetHTTPClient(context.Background())
	if err != nil {
		return nil, err
	}

	return api.New(api.ClientConfig{
		InstanceURL: cfg.InstanceURL,
		HTTPClient:  httpClient,
		APIVersion:  o.APIVersion,
	})
}

// SetAPIClient sets a test client (for testing only)
func (o *Options) SetAPIClient(client *api.Client) {
	o.testClient = client
}

// NewCmd creates the root command and returns the options struct
func NewCmd() (*cobra.Command, *Options) {
	opts := &Options{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	cmd := &cobra.Command{
		Use:   "sfdc",
		Short: "A CLI for Salesforce",
		Long: `sfdc is a command-line interface for Salesforce.

It provides tools for querying data, managing records, and more.
Run 'sfdc init' to set up authentication.`,
		Version:       version.Info(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags - bound to opts struct
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "table", "Output format: table, json, plain")
	cmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "Disable colored output")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().StringVar(&opts.APIVersion, "api-version", "", "Salesforce API version (default: v62.0)")

	return cmd, opts
}

// RegisterCommands registers subcommands with the root command
func RegisterCommands(root *cobra.Command, opts *Options, registrars ...func(*cobra.Command, *Options)) {
	for _, register := range registrars {
		register(root, opts)
	}
}
