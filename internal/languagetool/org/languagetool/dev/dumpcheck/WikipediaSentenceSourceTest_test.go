package dumpcheck

// Twin of languagetool-wikipedia/src/test/java/org/languagetool/dev/dumpcheck/WikipediaSentenceSourceTest.java
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of WikipediaSentenceSourceTest.testWikipediaSource
func TestWikipediaSentenceSource_WikipediaSource(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	path := filepath.Join(filepath.Dir(thisFile), "testdata", "wikipedia-en.xml")
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	src := NewWikipediaSentenceSource(f, "en")
	require.True(t, src.HasNext())
	s, err := src.Next()
	require.NoError(t, err)
	require.Equal(t, "This is the first document.", s.GetText())
	s, err = src.Next()
	require.NoError(t, err)
	require.Equal(t, "It has three sentences.", s.GetText())
	s, err = src.Next()
	require.NoError(t, err)
	require.Equal(t, "Here's the last sentence.", s.GetText())

	s, err = src.Next()
	require.NoError(t, err)
	require.Equal(t, "This is the second document.", s.GetText())
	s, err = src.Next()
	require.NoError(t, err)
	require.Equal(t, "It has two sentences.", s.GetText())
	require.False(t, src.HasNext())
}
