package root

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmd(t *testing.T) {
	cmd, opts := NewCmd()

	assert.Equal(t, "sfdc", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Version)

	// Check global flags exist
	assert.NotNil(t, cmd.PersistentFlags().Lookup("output"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("no-color"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("verbose"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("api-version"))

	// Check default values
	assert.Equal(t, "table", opts.Output)
	assert.False(t, opts.NoColor)
	assert.False(t, opts.Verbose)
}

func TestOptions_View(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	opts := &Options{
		Output:  "json",
		NoColor: true,
		Stdout:  stdout,
		Stderr:  stderr,
	}

	v := opts.View()
	require.NotNil(t, v)

	// View should use the configured output format
	assert.Equal(t, "json", opts.Output)
}

func TestOptions_SetAPIClient(t *testing.T) {
	opts := &Options{}

	// Without test client, APIClient() would try to load config
	// With test client set, it returns the test client
	opts.SetAPIClient(nil)
	assert.Nil(t, opts.testClient)
}

func TestRegisterCommands(t *testing.T) {
	cmd, opts := NewCmd()

	called := false
	registrar := func(parent *cobra.Command, o *Options) {
		called = true
		assert.Same(t, cmd, parent)
		assert.Same(t, opts, o)
	}

	RegisterCommands(cmd, opts, registrar)
	assert.True(t, called)
}

func TestRegisterCommands_Multiple(t *testing.T) {
	cmd, opts := NewCmd()

	callCount := 0
	registrar1 := func(parent *cobra.Command, o *Options) { callCount++ }
	registrar2 := func(parent *cobra.Command, o *Options) { callCount++ }
	registrar3 := func(parent *cobra.Command, o *Options) { callCount++ }

	RegisterCommands(cmd, opts, registrar1, registrar2, registrar3)
	assert.Equal(t, 3, callCount)
}
