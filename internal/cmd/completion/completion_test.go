package completion

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/open-cli-collective/salesforce-cli/internal/cmd/root"
)

func TestRegister(t *testing.T) {
	rootCmd, opts := root.NewCmd()
	Register(rootCmd, opts)

	// Check that completion command was added
	completionCmd, _, err := rootCmd.Find([]string{"completion"})
	assert.NoError(t, err)
	assert.NotNil(t, completionCmd)
	assert.Equal(t, "completion [bash|zsh|fish|powershell]", completionCmd.Use)

	// Check valid args
	assert.Contains(t, completionCmd.ValidArgs, "bash")
	assert.Contains(t, completionCmd.ValidArgs, "zsh")
	assert.Contains(t, completionCmd.ValidArgs, "fish")
	assert.Contains(t, completionCmd.ValidArgs, "powershell")
}

func TestCompletion_RequiresArg(t *testing.T) {
	rootCmd, opts := root.NewCmd()
	Register(rootCmd, opts)

	// Running without args should fail
	rootCmd.SetArgs([]string{"completion"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestCompletion_InvalidArg(t *testing.T) {
	rootCmd, opts := root.NewCmd()
	Register(rootCmd, opts)

	// Running with invalid arg should fail
	rootCmd.SetArgs([]string{"completion", "invalid"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestCompletion_Bash(t *testing.T) {
	rootCmd := &cobra.Command{Use: "sfdc"}
	opts := &root.Options{}
	Register(rootCmd, opts)

	rootCmd.SetArgs([]string{"completion", "bash"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestCompletion_Zsh(t *testing.T) {
	rootCmd := &cobra.Command{Use: "sfdc"}
	opts := &root.Options{}
	Register(rootCmd, opts)

	rootCmd.SetArgs([]string{"completion", "zsh"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestCompletion_Fish(t *testing.T) {
	rootCmd := &cobra.Command{Use: "sfdc"}
	opts := &root.Options{}
	Register(rootCmd, opts)

	rootCmd.SetArgs([]string{"completion", "fish"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestCompletion_PowerShell(t *testing.T) {
	rootCmd := &cobra.Command{Use: "sfdc"}
	opts := &root.Options{}
	Register(rootCmd, opts)

	rootCmd.SetArgs([]string{"completion", "powershell"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}
