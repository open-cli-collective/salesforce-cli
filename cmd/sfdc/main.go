// Package main is the entry point for the sfdc CLI.
package main

import (
	"fmt"
	"os"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/completion"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/configcmd"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/initcmd"
	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

// Exit codes
const (
	exitOK    = 0
	exitError = 1
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitError)
	}
	os.Exit(exitOK)
}

func run() error {
	rootCmd, opts := root.NewCmd()

	// Register all commands
	initcmd.Register(rootCmd, opts)
	configcmd.Register(rootCmd, opts)
	completion.Register(rootCmd, opts)

	return rootCmd.Execute()
}
