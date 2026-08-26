package tagging

// Twin of ManualTaggerTest
import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func deAddedPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(file)
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/added.txt")
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		dir = filepath.Dir(dir)
	}
	return "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/added.txt"
}

func TestManualTagger_Tag(t *testing.T) {
	f, err := os.Open(deAddedPath(t))
	require.NoError(t, err)
	defer f.Close()
	tagger, err := NewManualTagger(f)
	require.NoError(t, err)
	require.Equal(t, 0, len(tagger.Tag("")))
	require.Equal(t, 0, len(tagger.Tag("gibtsnicht")))
	got := tagger.Tag("drunter")
	require.NotEmpty(t, got)
	var parts []string
	for _, tw := range got {
		parts = append(parts, tw.String())
	}
	require.Equal(t, "[drunter/ADV:LOK+PRO]", "["+strings.Join(parts, ", ")+"]")
	require.Equal(t, 0, len(tagger.Tag("Drunter")))
}

func TestManualTagger_JavaTrimKeepsNBSPDetection(t *testing.T) {
	// Java rejects lines containing NBSP after line.trim() (trim does not remove NBSP).
	_, err := NewManualTagger(strings.NewReader("\u00a0foo\tbar\tNN\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "Non-breaking space")
}

func TestManualTagger_Inline(t *testing.T) {
	tagger, err := NewManualTagger(strings.NewReader("foo\tbar\tNN\n"))
	require.NoError(t, err)
	require.Equal(t, "bar", tagger.Tag("foo")[0].GetLemma())
	require.Equal(t, "NN", tagger.Tag("foo")[0].GetPosTag())
}
