package commandline

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOutputFormat(t *testing.T) {
	f, err := parseOutputFormat("text")
	require.NoError(t, err)
	require.Equal(t, OutputLint, f)
	f, err = parseOutputFormat("json")
	require.NoError(t, err)
	require.Equal(t, OutputJSON, f)
	f, err = parseOutputFormat("sarif")
	require.NoError(t, err)
	require.Equal(t, OutputSARIF, f)
	f, err = parseOutputFormat("plaintext")
	require.NoError(t, err)
	require.Equal(t, OutputPlaintext, f)
	_, err = parseOutputFormat("bogus")
	require.Error(t, err)
}

func TestParseOptions_Format(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--format", "text", "-l", "en", "-"})
	require.NoError(t, err)
	require.Equal(t, OutputLint, opts.OutputFormat)
}

func TestCoreCheckHook_FormatText(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--format", "text", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 1, code)
	require.Contains(t, out.String(), "EN_A_VS_AN")
	require.Contains(t, out.String(), "location")
}
