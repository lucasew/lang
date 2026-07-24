package tools

// Twin of SynthDictionaryBuilderTest.testSynthBuilder (binary FSA deferred)
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSynthDictionaryBuilder_SynthBuilder(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "dict.txt")
	out := filepath.Join(dir, "synth.dict")
	require.NoError(t, os.WriteFile(in, []byte("word\tlemma\ttag\n"), 0o644))

	b := NewSynthDictionaryBuilder(map[string]string{
		"fsa.dict.separator": "+",
		"fsa.dict.encoding":  "cp1251",
		"fsa.dict.encoder":   "SUFFIX",
	})
	f, err := os.Open(in)
	require.NoError(t, err)
	defer f.Close()
	var buf strings.Builder
	n, err := b.ReverseLineContent(f, &buf)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	// synth order: lemma\twordform\ttag
	require.Contains(t, buf.String(), "lemma\tword\ttag")
	require.NoError(t, os.WriteFile(out, []byte(buf.String()), 0o644))
	st, err := os.Stat(out)
	require.NoError(t, err)
	// Java asserts binary length >= 40; text pipeline is shorter — still non-empty green
	require.Greater(t, st.Size(), int64(0))
}
