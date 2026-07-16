package tools

// Twin of POSDictionaryBuilderTest.testPOSBuilder (binary FSA deferred)
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPOSDictionaryBuilder_POSBuilder(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "dict.txt")
	info := filepath.Join(dir, "dict.info")
	out := filepath.Join(dir, "dict.dict")
	require.NoError(t, os.WriteFile(in, []byte("word\tlemma\ttag\n"), 0o644))
	require.NoError(t, os.WriteFile(info, []byte(
		"fsa.dict.separator=+\n"+
			"fsa.dict.encoding=cp1251\n"+
			"fsa.dict.encoder=SUFFIX\n"), 0o644))

	// Text pipeline: normalize input into output path (compile deferred)
	props := map[string]string{
		"fsa.dict.separator": "+",
		"fsa.dict.encoding":  "cp1251",
		"fsa.dict.encoder":   "SUFFIX",
	}
	b := NewPOSDictionaryBuilder(props)
	require.Equal(t, "cp1251", b.Encoding())
	require.Equal(t, "+", b.Separator())
	f, err := os.Open(in)
	require.NoError(t, err)
	defer f.Close()
	var buf strings.Builder
	n, err := b.NormalizeTaggerInput(f, &buf)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Contains(t, buf.String(), "word\tlemma\ttag")
	require.NoError(t, os.WriteFile(out, []byte(buf.String()), 0o644))
	st, err := os.Stat(out)
	require.NoError(t, err)
	require.True(t, st.Size() > 0)
}
