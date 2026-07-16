package errorcorpus

// Twin of PedlerCorpusTest.testCorpusAccess
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPedlerCorpus_CorpusAccess(t *testing.T) {
	dir := t.TempDir()
	// Three lines matching Java fixture expectations
	content := "But <ERR targ=foo>also</ERR> please <ERR targ=note>not</ERR> that grammar checkers aren't perfect.\n" +
		"But <ERR targ=bad suggestion>also also</ERR> please <ERR targ=note>note note</ERR> that grammar checkers aren't perfect.\n" +
		"Third sentence with no error markup.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "errors.txt"), []byte(content), 0o644))
	// non-txt ignored
	require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.md"), []byte("x"), 0o644))

	corpus, err := NewPedlerCorpus(dir)
	require.NoError(t, err)
	require.True(t, corpus.HasNext())
	s1 := corpus.Next()
	require.Equal(t, "But also please not that grammar checkers aren't perfect.", s1.PlainText)
	require.Equal(t, "But <ERR targ=foo>also</ERR> please <ERR targ=note>not</ERR> that grammar checkers aren't perfect.", s1.MarkupText)

	require.True(t, corpus.HasNext())
	s2 := corpus.Next()
	require.Equal(t, "But also also please note note that grammar checkers aren't perfect.", s2.PlainText)
	require.Equal(t, "But <ERR targ=bad suggestion>also also</ERR> please <ERR targ=note>note note</ERR> that grammar checkers aren't perfect.", s2.MarkupText)

	require.True(t, corpus.HasNext())
	_ = corpus.Next()
	require.False(t, corpus.HasNext())
}
