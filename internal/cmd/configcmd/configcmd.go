// Package configcmd provides the config command and subcommands.
package configcmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/auth"
	"github.com/open-cli-collective/salesforce-cli/internal/config"
	"github.com/open-cli-collective/salesforce-cli/internal/keychain"
)

// Register registers the config command with the parent command.
// The opts parameter is accepted for consistency with other commands but not used.
func Register(parent *cobra.Command, _ interface{}) {
	parent.AddCommand(NewCommand())
}

// NewCommand returns the config command with subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "View, test, and manage Salesforce CLI configuration.",
	}

	cmd.AddCommand(newShowCommand())
	cmd.AddCommand(newTestCommand())
	cmd.AddCommand(newClearCommand())

	return cmd
}

func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Long:  "Display the current Salesforce CLI configuration including instance URL and token storage.",
		Args:  cobra.NoArgs,
		RunE:  runShow,
	}
}

func newTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Verify authentication works",
		Long:  "Test the current OAuth token by making a request to the Salesforce API.",
		Args:  cobra.NoArgs,
		RunE:  runTest,
	}
}

func newClearCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove stored credentials",
		Long:  "Remove all stored credentials and configuration. This will require re-authentication.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClear(force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Println("Salesforce CLI Configuration")
	fmt.Println("============================")
	fmt.Println()

	if cfg.InstanceURL != "" {
		fmt.Printf("Instance URL:    %s\n", cfg.InstanceURL)
	} else {
		fmt.Println("Instance URL:    Not configured")
	}

	if cfg.ClientID != "" {
		masked := maskClientID(cfg.ClientID)
		fmt.Printf("Client ID:       %s\n", masked)
	} else {
		fmt.Println("Client ID:       Not configured")
	}

	fmt.Println()
	if keychain.HasStoredToken() {
		fmt.Printf("Token:           Found (stored in %s)\n", keychain.GetStorageBackend())
	} else {
		fmt.Println("Token:           Not found")
	}

	fmt.Println()
	configPath, err := config.GetConfigPath()
	if err != nil {
		configPath = "(unable to determine)"
	}
	fmt.Printf("Config file:     %s\n", configPath)

	return nil
}

func runTest(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.InstanceURL == "" || cfg.ClientID == "" {
		return fmt.Errorf("not configured - please run 'sfdc init' first")
	}

	fmt.Println("Testing Salesforce connection...")
	fmt.Println()

	if !keychain.HasStoredToken() {
		fmt.Println("  Token:       NOT FOUND")
		return fmt.Errorf("no OAuth token found - please run 'sfdc init' first")
	}
	fmt.Println("  Token:       Found")

	ctx := context.Background()
	client, err := auth.GetHTTPClient(ctx)
	if err != nil {
		fmt.Println("  OAuth:       FAILED")
		return fmt.Errorf("failed to create OAuth client: %w", err)
	}
	fmt.Println("  OAuth:       OK")

	normalizedURL := normalizeURL(cfg.InstanceURL)
	resp, err := client.Get(normalizedURL + "/services/data/")
	if err != nil {
		fmt.Println("  API:         FAILED")
		return fmt.Errorf("failed to access Salesforce API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("  API:         FAILED")
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	fmt.Println("  API:         OK")

	fmt.Println()
	fmt.Println("Connection successful!")
	return nil
}

func runClear(force bool) error {
	if !force {
		fmt.Print("This will remove all stored credentials. Continue? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	var hadToken, hadConfig bool
	var tokenErr, configErr error

	if keychain.HasStoredToken() {
		hadToken = true
		tokenErr = keychain.DeleteToken()
	}

	cfg, _ := config.Load()
	if cfg.InstanceURL != "" || cfg.ClientID != "" {
		hadConfig = true
		cfg.InstanceURL = ""
		cfg.ClientID = ""
		configErr = config.Save(cfg)
	}

	if tokenErr != nil {
		fmt.Printf("Warning: failed to remove token: %v\n", tokenErr)
	} else if hadToken {
		fmt.Println("Token removed.")
	}

	if configErr != nil {
		fmt.Printf("Warning: failed to clear config: %v\n", configErr)
	} else if hadConfig {
		fmt.Println("Configuration cleared.")
	}

	if !hadToken && !hadConfig {
		fmt.Println("Nothing to clear.")
	} else if tokenErr == nil && configErr == nil {
		fmt.Println()
		fmt.Println("All credentials cleared. Run 'sfdc init' to reconfigure.")
	}

	return nil
}

// maskClientID masks a client ID for display, showing only first and last 4 chars.
func maskClientID(clientID string) string {
	if clientID == "" {
		return ""
	}
	if len(clientID) <= 8 {
		return "********"
	}
	return clientID[:4] + "********" + clientID[len(clientID)-4:]
}

// normalizeURL ensures the URL has https:// prefix.
func normalizeURL(url string) string {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return strings.TrimSuffix(url, "/")
}
