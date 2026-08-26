package commandline

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectUnknownWords(t *testing.T) {
	known := map[string]bool{"the": true, "cat": true, "sat": true}
	words := CollectUnknownWords("The cat sat on xyzzy mat.", func(tok string) bool {
		return known[tok] || known[stringsToLower(tok)]
	})
	require.Contains(t, words, "xyzzy")
	require.Contains(t, words, "mat")
	require.NotContains(t, words, "cat")
}

func stringsToLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func TestMain_ListUnknown(t *testing.T) {
	var out, errb bytes.Buffer
	known := map[string]bool{"hello": true, "world": true}
	code := RunWithIO([]string{"-l", "en", "-u", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "hello xyzzy world", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.IsListUnknown())
			unk := CollectUnknownWords(text, func(tok string) bool {
				return known[tok] || known[stringsToLower(tok)]
			})
			PrintUnknownWords(w, unk)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "Unknown words:")
	require.Contains(t, out.String(), "xyzzy")
}

func TestMain_NoListUnknown(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "ok", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.False(t, opts.IsListUnknown())
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.NotContains(t, out.String(), "Unknown words:")
}
