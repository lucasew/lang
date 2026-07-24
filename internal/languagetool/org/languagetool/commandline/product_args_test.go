package commandline

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeProductArgs(t *testing.T) {
	require.Equal(t, []string{"--lint", "-l", "en", "-"}, NormalizeProductArgs([]string{"lint", "-l", "en", "-"}))
	require.Equal(t, []string{"--list"}, NormalizeProductArgs([]string{"languages"}))
	require.Equal(t, []string{"--list"}, NormalizeProductArgs([]string{"list"}))
	require.Equal(t, []string{"--version"}, NormalizeProductArgs([]string{"version"}))
	require.Equal(t, []string{"--help"}, NormalizeProductArgs([]string{"help"}))
	require.Equal(t, []string{"-l", "en"}, NormalizeProductArgs([]string{"-l", "en"}))
}

func TestRunWithIO_ProductLintSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"lint", "-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 1, code)
	require.Contains(t, out.String(), "EN_A_VS_AN")
	require.Contains(t, out.String(), "location") // lint header
	_ = io.Discard
}

func TestRunWithIO_ProductLanguages(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"languages"}, DefaultCoreHooks(), &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "en-US")
}
