package dumpcheck

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkipSentence(t *testing.T) {
	require.True(t, SkipSentence(""))
	require.True(t, SkipSentence("  lower case start."))
	require.False(t, SkipSentence("Upper case start."))
}

func TestExtractWikipediaSentences(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	path := filepath.Join(filepath.Dir(thisFile), "testdata", "wikipedia-en.xml")
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	src := NewWikipediaSentenceSource(f, "en")
	var buf strings.Builder
	n, err := ExtractWikipediaSentences(src, &buf)
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 4)
	out := buf.String()
	require.Contains(t, out, "This is the first document.")
	require.Contains(t, out, "This is the second document.")
	// none of the exported lines should start lower
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		require.False(t, SkipSentence(line), line)
	}
}
