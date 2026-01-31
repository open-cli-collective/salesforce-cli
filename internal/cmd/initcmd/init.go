// Package initcmd provides the init command for OAuth setup.
package initcmd

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/salesforce-cli/internal/auth"
	"github.com/open-cli-collective/salesforce-cli/internal/config"
	"github.com/open-cli-collective/salesforce-cli/internal/keychain"
)

var (
	instanceURL string
	clientID    string
	noVerify    bool
	noBrowser   bool
)

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
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't auto-open browser, just print URL")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: Check existing configuration
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

			// Token exists but verification failed
			fmt.Println()
			fmt.Println("Your OAuth token appears to be expired or revoked.")
			fmt.Print("Would you like to re-authenticate? [Y/n]: ")

			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "" && input != "y" && input != "yes" {
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

	// Step 2: Get instance URL
	if instanceURL == "" {
		instanceURL = cfg.InstanceURL
	}
	if instanceURL == "" {
		fmt.Println("Salesforce Instance URL")
		fmt.Println("Enter your Salesforce login URL, or leave blank for production.")
		fmt.Println()
		fmt.Println("  Production: login.salesforce.com (default)")
		fmt.Println("  Sandbox:    test.salesforce.com")
		fmt.Println("  Custom:     mycompany.my.salesforce.com")
		fmt.Println()
		fmt.Print("Instance URL [login.salesforce.com]: ")

		input, _ := reader.ReadString('\n')
		instanceURL = strings.TrimSpace(input)
		if instanceURL == "" {
			instanceURL = "login.salesforce.com"
		}
	}

	// Step 3: Get client ID
	if clientID == "" {
		clientID = cfg.ClientID
	}
	if clientID == "" {
		fmt.Println()
		fmt.Println("Connected App Client ID")
		fmt.Println("Create at: Setup → App Manager → New Connected App")
		fmt.Println("Required scopes: api, refresh_token, offline_access")
		fmt.Println("Callback URL: http://localhost:8080/callback")
		fmt.Println()
		fmt.Print("Client ID: ")

		input, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(input)
		if clientID == "" {
			return fmt.Errorf("client ID is required")
		}
	}

	// Step 4: Save configuration
	cfg.InstanceURL = instanceURL
	cfg.ClientID = clientID
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Step 5: Start OAuth flow
	oauthConfig := auth.GetOAuthConfig(instanceURL, clientID)
	authURL := auth.GetAuthURL(oauthConfig)

	fmt.Println()
	if noBrowser {
		fmt.Println("Open this URL in your browser:")
	} else {
		fmt.Println("Opening browser for Salesforce login...")
		fmt.Println()
		fmt.Println("If browser doesn't open, visit:")
	}
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()

	// Start callback server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resultChan, err := auth.StartCallbackServer(ctx, auth.DefaultCallbackPort)
	if err != nil {
		fmt.Printf("Warning: Could not start callback server: %v\n", err)
		fmt.Println("You'll need to manually copy the authorization code.")
	}

	// Open browser (unless --no-browser)
	if !noBrowser {
		if err := openBrowser(authURL); err != nil {
			fmt.Printf("Could not open browser: %v\n", err)
		}
	}

	// Wait for callback or manual input
	var code string
	if resultChan != nil {
		fmt.Println("Waiting for authorization...")
		fmt.Println("(Or paste the authorization code or full redirect URL below)")
		fmt.Println()

		// Read from callback or stdin
		inputChan := make(chan string, 1)
		go func() {
			fmt.Print("> ")
			input, _ := reader.ReadString('\n')
			inputChan <- strings.TrimSpace(input)
		}()

		select {
		case result := <-resultChan:
			if result.Error != "" {
				return fmt.Errorf("authorization failed: %s", result.Error)
			}
			code = result.Code
			fmt.Println("Authorization received from callback.")
		case input := <-inputChan:
			code = extractAuthCode(input)
		case <-ctx.Done():
			return fmt.Errorf("authorization timed out")
		}
	} else {
		fmt.Println("After authorizing, paste the authorization code or full redirect URL:")
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		code = extractAuthCode(strings.TrimSpace(input))
	}

	if code == "" {
		return fmt.Errorf("no authorization code received")
	}

	// Step 6: Exchange code for token
	fmt.Println()
	fmt.Println("Exchanging authorization code for tokens...")

	token, err := auth.ExchangeAuthCode(ctx, oauthConfig, code)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Step 7: Save token
	if err := keychain.SetToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	fmt.Printf("Token saved to: %s\n", keychain.GetStorageBackend())

	// Step 8: Verify connectivity
	if !noVerify {
		fmt.Println()
		if err := verifyConnectivity(instanceURL); err != nil {
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

	// If it looks like a URL, try to extract the code parameter
	if strings.HasPrefix(input, "http://localhost") || strings.HasPrefix(input, "https://localhost") {
		if u, err := url.Parse(input); err == nil {
			return u.Query().Get("code")
		}
		return ""
	}

	// Otherwise treat as raw code
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

	// Test API access by fetching API versions
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

// openBrowser opens the default browser to the given URL.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}
