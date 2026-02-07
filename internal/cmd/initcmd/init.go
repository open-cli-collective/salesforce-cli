// Package initcmd provides the init command for OAuth setup.
package initcmd

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/auth"
	"github.com/open-cli-collective/salesforce-cli/internal/config"
	"github.com/open-cli-collective/salesforce-cli/internal/keychain"
)

var (
	instanceURL string
	clientID    string
	noVerify    bool
)

// Register registers the init command with the parent command.
// The opts parameter is accepted for consistency with other commands but not used.
func Register(parent *cobra.Command, _ interface{}) {
	parent.AddCommand(NewCommand())
}

// NewCommand returns the init command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up Salesforce authentication",
		Long: `Guided setup for Salesforce OAuth 2.0 authentication.

This command walks you through the OAuth flow with clear instructions.
After setup, you can use commands like 'sfdc query', etc.

Prerequisites:
  1. Create a Connected App in Salesforce Setup
  2. Enable OAuth Settings
  3. Set Callback URL: http://localhost:8080/callback
  4. Select scopes: api, refresh_token, offline_access
  5. Note the Consumer Key (Client ID)`,
		Args: cobra.NoArgs,
		RunE: runInit,
	}

	cmd.Flags().StringVar(&instanceURL, "instance-url", "", "Salesforce instance URL (e.g., login.salesforce.com)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Connected App Consumer Key")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "Skip connectivity verification after setup")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Checking existing configuration...")
	cfg, _ := config.Load()

	if keychain.HasStoredToken() {
		fmt.Printf("Instance URL: %s\n", cfg.InstanceURL)
		fmt.Printf("Token:        Found (stored in %s)\n", keychain.GetStorageBackend())

		if !noVerify {
			if err := verifyConnectivity(cfg.InstanceURL); err == nil {
				fmt.Println()
				fmt.Println("Already configured and working.")
				fmt.Println("Use 'sfdc config clear' to reset.")
				return nil
			}

			fmt.Println()
			fmt.Println("Your OAuth token appears to be expired or revoked.")

			var reauth bool
			err := huh.NewConfirm().
				Title("Would you like to re-authenticate?").
				Value(&reauth).
				Run()
			if err != nil {
				return err
			}
			if !reauth {
				fmt.Println("You can manually clear the token with: sfdc config clear")
				return nil
			}

			fmt.Println("Clearing old token...")
			if err := keychain.DeleteToken(); err != nil {
				return fmt.Errorf("failed to clear token: %w", err)
			}
		}
	} else {
		fmt.Println("Instance URL: Not configured")
		fmt.Println("Token:        Not found")
	}
	fmt.Println()

	// Pre-fill from existing config, then override with CLI flags
	// Priority: CLI flag > existing config value
	formInstanceURL := ""
	formClientID := ""

	if instanceURL != "" {
		formInstanceURL = instanceURL
	} else if cfg.InstanceURL != "" {
		formInstanceURL = cfg.InstanceURL
	}

	if clientID != "" {
		formClientID = clientID
	} else if cfg.ClientID != "" {
		formClientID = cfg.ClientID
	}

	// Build the form for configuration inputs
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Instance URL").
				Description("Production: login.salesforce.com | Sandbox: test.salesforce.com").
				Placeholder("login.salesforce.com").
				Value(&formInstanceURL),

			huh.NewInput().
				Title("Client ID").
				Description("Connected App Consumer Key from Setup → App Manager").
				Value(&formClientID).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("client ID is required")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Apply defaults
	if formInstanceURL == "" {
		formInstanceURL = "login.salesforce.com"
	}

	cfg.InstanceURL = formInstanceURL
	cfg.ClientID = formClientID
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	oauthConfig := auth.GetOAuthConfig(formInstanceURL, formClientID)
	authURL := auth.GetAuthURL(oauthConfig)

	fmt.Println()
	fmt.Println("Open this URL in your browser:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("After clicking 'Allow', your browser will redirect to a localhost URL.")
	fmt.Println("This will show an error - that's expected!")
	fmt.Println()
	fmt.Println("Copy the ENTIRE URL from your browser's address bar and paste it here,")
	fmt.Println("or just paste the 'code' parameter value:")
	fmt.Println()
	fmt.Print("> ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	code := extractAuthCode(strings.TrimSpace(input))
	if code == "" {
		return fmt.Errorf("no authorization code received")
	}

	fmt.Println()
	fmt.Println("Exchanging authorization code for tokens...")

	ctx := context.Background()
	token, err := auth.ExchangeAuthCode(ctx, oauthConfig, code)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	if err := keychain.SetToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	fmt.Printf("Token saved to: %s\n", keychain.GetStorageBackend())

	if !noVerify {
		fmt.Println()
		if err := verifyConnectivity(formInstanceURL); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Println("Setup complete! Try: sfdc query \"SELECT Id, Name FROM Account LIMIT 5\"")
	return nil
}

// extractAuthCode extracts the authorization code from user input.
// It accepts either a full redirect URL or just the code value.
func extractAuthCode(input string) string {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "http://localhost") || strings.HasPrefix(input, "https://localhost") {
		if u, err := url.Parse(input); err == nil {
			return u.Query().Get("code")
		}
		return ""
	}

	return input
}

// verifyConnectivity tests the Salesforce API connection.
func verifyConnectivity(instanceURL string) error {
	fmt.Println("Verifying Salesforce API connection...")

	ctx := context.Background()
	client, err := auth.GetHTTPClient(ctx)
	if err != nil {
		fmt.Println("  OAuth token: FAILED")
		return fmt.Errorf("failed to create client: %w", err)
	}
	fmt.Println("  OAuth token: OK")

	normalizedURL := "https://" + strings.TrimPrefix(strings.TrimPrefix(instanceURL, "https://"), "http://")
	resp, err := client.Get(normalizedURL + "/services/data/")
	if err != nil {
		fmt.Println("  API access:  FAILED")
		return fmt.Errorf("failed to access Salesforce API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("  API access:  FAILED")
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	fmt.Println("  API access:  OK")

	return nil
}
