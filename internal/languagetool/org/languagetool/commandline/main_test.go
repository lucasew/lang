package commandline

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunHelpVersion(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"--help"}, RunHooks{}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "Usage:")

	out.Reset()
	code = RunWithIO([]string{"--version"}, RunHooks{}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "languagetool")
}

func TestRunCheck(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "hello", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "en", opts.Language)
			require.Equal(t, "hello", text)
			_, _ = io.WriteString(w, "ok\n")
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Equal(t, "ok\n", out.String())
}
