package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandLineParser(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--version", "-l", "en-US", "-v", "file.txt"})
	require.NoError(t, err)
	require.True(t, opts.PrintVersion)
	require.Equal(t, "en-US", opts.Language)
	require.True(t, opts.Verbose)
	require.Equal(t, "file.txt", opts.Filename)

	opts, err = p.ParseOptions([]string{"-d", "A,B", "-e", "C"})
	require.NoError(t, err)
	require.Equal(t, []string{"A", "B"}, opts.DisabledRules)
	require.Equal(t, []string{"C"}, opts.EnabledRules)

	_, err = p.ParseOptions([]string{"--enabledonly", "-d", "X"})
	require.Error(t, err)

	_, err = p.ParseOptions([]string{"--unknown-flag"})
	require.Error(t, err)

	opts, err = p.ParseOptions(nil)
	require.NoError(t, err)
	require.True(t, opts.PrintUsage)
}
