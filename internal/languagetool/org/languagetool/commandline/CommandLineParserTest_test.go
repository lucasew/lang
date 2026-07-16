package commandline

// Twin of languagetool-commandline/src/test/java/org/languagetool/commandline/CommandLineParserTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandLineParser_Usage(t *testing.T) {
	p := &CommandLineParser{}
	// Go port: empty args print usage instead of WrongParameterNumberException
	opts, err := p.ParseOptions(nil)
	require.NoError(t, err)
	require.True(t, opts.PrintUsage)

	opts, err = p.ParseOptions([]string{"--help"})
	require.NoError(t, err)
	require.True(t, opts.PrintUsage)
}

func TestCommandLineParser_Errors(t *testing.T) {
	p := &CommandLineParser{}
	_, err := p.ParseOptions([]string{"--apply", "--taggeronly"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "apply")
}

func TestCommandLineParser_Simple(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"filename.txt"})
	require.NoError(t, err)
	require.Equal(t, "filename.txt", opts.Filename)
	require.False(t, opts.Verbose)

	opts, err = p.ParseOptions([]string{"-v", "-l", "xx", "filename.txt"})
	require.NoError(t, err)
	require.True(t, opts.Verbose)
	require.Equal(t, "filename.txt", opts.Filename)

	opts, err = p.ParseOptions([]string{"--version"})
	require.NoError(t, err)
	require.True(t, opts.PrintVersion)

	opts, err = p.ParseOptions([]string{"--list"})
	require.NoError(t, err)
	require.True(t, opts.PrintLanguages)
}
