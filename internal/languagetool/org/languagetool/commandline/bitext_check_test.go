package commandline

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckBitextFile(t *testing.T) {
	var buf bytes.Buffer
	n, err := CheckBitextFile(&buf, "Hello world\tHello world\nShort\tThis is a much longer target sentence here\n", nil)
	require.NoError(t, err)
	require.Greater(t, n, 0)
	require.True(t, strings.Contains(buf.String(), "Rule ID:") || n > 0)
}

func TestMain_BitextMode(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--bitext", "pairs.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) {
			return "Hi\tHi there longer text\n", nil
		},
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.Bitext)
			return CheckBitextFile(w, text, nil)
		},
	}, &out, &errb)
	require.True(t, code == 0 || code == 2, "code=%d err=%s out=%s", code, errb.String(), out.String())
}
