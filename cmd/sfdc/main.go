// Package main is the entry point for the sfdc CLI.
package main

import (
	"fmt"
	"os"

	"github.com/open-cli-collective/salesforce-cli/internal/version"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// TODO: Initialize root command and execute
	fmt.Printf("sfdc %s\n", version.Info())
	return nil
}
