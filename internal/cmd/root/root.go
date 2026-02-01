// Package root provides the root command and global options.
package root

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/api"
	"github.com/open-cli-collective/salesforce-cli/api/bulk"
	"github.com/open-cli-collective/salesforce-cli/api/metadata"
	"github.com/open-cli-collective/salesforce-cli/api/tooling"
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
	// testBulkClient is used for testing; if set, BulkClient() returns this instead
	testBulkClient *bulk.Client
	// testToolingClient is used for testing; if set, ToolingClient() returns this instead
	testToolingClient *tooling.Client
	// testMetadataClient is used for testing; if set, MetadataClient() returns this instead
	testMetadataClient *metadata.Client
}

// View returns a configured View instance
func (o *Options) View() *view.View {
	v := view.NewWithFormat(o.Output, o.NoColor)
	v.Out = o.Stdout
	v.Err = o.Stderr
	return v
}

// loadClientConfig loads common configuration needed for API clients.
func (o *Options) loadClientConfig() (instanceURL string, httpClient *http.Client, err error) {
	cfg, err := config.Load()
	if err != nil {
		return "", nil, err
	}

	httpClient, err = auth.GetHTTPClient(context.Background())
	if err != nil {
		return "", nil, err
	}

	return cfg.InstanceURL, httpClient, nil
}

// APIClient creates a new API client from config
func (o *Options) APIClient() (*api.Client, error) {
	if o.testClient != nil {
		return o.testClient, nil
	}

	instanceURL, httpClient, err := o.loadClientConfig()
	if err != nil {
		return nil, err
	}

	return api.New(api.ClientConfig{
		InstanceURL: instanceURL,
		HTTPClient:  httpClient,
		APIVersion:  o.APIVersion,
	})
}

// SetAPIClient sets a test client (for testing only)
func (o *Options) SetAPIClient(client *api.Client) {
	o.testClient = client
}

// BulkClient creates a new Bulk API client from config
func (o *Options) BulkClient() (*bulk.Client, error) {
	if o.testBulkClient != nil {
		return o.testBulkClient, nil
	}

	instanceURL, httpClient, err := o.loadClientConfig()
	if err != nil {
		return nil, err
	}

	return bulk.New(bulk.ClientConfig{
		InstanceURL: instanceURL,
		HTTPClient:  httpClient,
		APIVersion:  o.APIVersion,
	})
}

// SetBulkClient sets a test bulk client (for testing only)
func (o *Options) SetBulkClient(client *bulk.Client) {
	o.testBulkClient = client
}

// ToolingClient creates a new Tooling API client from config
func (o *Options) ToolingClient() (*tooling.Client, error) {
	if o.testToolingClient != nil {
		return o.testToolingClient, nil
	}

	instanceURL, httpClient, err := o.loadClientConfig()
	if err != nil {
		return nil, err
	}

	return tooling.New(tooling.ClientConfig{
		InstanceURL: instanceURL,
		HTTPClient:  httpClient,
		APIVersion:  o.APIVersion,
	})
}

// SetToolingClient sets a test tooling client (for testing only)
func (o *Options) SetToolingClient(client *tooling.Client) {
	o.testToolingClient = client
}

// MetadataClient creates a new Metadata API client from config
func (o *Options) MetadataClient() (*metadata.Client, error) {
	if o.testMetadataClient != nil {
		return o.testMetadataClient, nil
	}

	instanceURL, httpClient, err := o.loadClientConfig()
	if err != nil {
		return nil, err
	}

	return metadata.New(metadata.ClientConfig{
		InstanceURL: instanceURL,
		HTTPClient:  httpClient,
		APIVersion:  o.APIVersion,
	})
}

// SetMetadataClient sets a test metadata client (for testing only)
func (o *Options) SetMetadataClient(client *metadata.Client) {
	o.testMetadataClient = client
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
