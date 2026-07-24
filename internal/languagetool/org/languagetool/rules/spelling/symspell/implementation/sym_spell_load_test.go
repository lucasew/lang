package implementation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of SymSpell.loadDictionary(BufferedReader, termIndex, countIndex).
func TestLoadDictionary(t *testing.T) {
	s := DefaultSymSpell()
	// format: word count
	corpus := "hello 100\nworld 50\n# skip? no fields\nbadline\nok 3\n"
	require.True(t, s.LoadDictionary(strings.NewReader(corpus), 0, 1))
	require.Equal(t, 3, s.WordCount())
	got := s.Lookup("helo", VerbosityClosest)
	require.NotEmpty(t, got)
	require.Equal(t, "hello", got[0].Term)
}

// Twin: Java line.split("\\s") keeps empty mid-fields; NBSP is not a delimiter.
func TestJavaSplitASCIIWhitespaceSingle(t *testing.T) {
	// "a  b" → ["a", "", "b"] (not Fields collapse)
	require.Equal(t, []string{"a", "", "b"}, javaSplitASCIIWhitespaceSingle("a  b"))
	// tab delimiter
	require.Equal(t, []string{"word", "42"}, javaSplitASCIIWhitespaceSingle("word\t42"))
	// trailing empty dropped (Java limit 0)
	require.Equal(t, []string{"a", "b"}, javaSplitASCIIWhitespaceSingle("a b "))
	// NBSP not \s without UNICODE_CHARACTER_CLASS
	require.Equal(t, []string{"a\u00a0b", "1"}, javaSplitASCIIWhitespaceSingle("a\u00a0b 1"))
}

func TestLoadDictionary_DoubleSpaceEmptyMidField(t *testing.T) {
	s := DefaultSymSpell()
	// termIndex 0, countIndex 2: "word  99" → ["word", "", "99"]
	require.True(t, s.LoadDictionary(strings.NewReader("word  99\n"), 0, 2))
	require.Equal(t, 1, s.WordCount())
}

// Twin: termIndex/countIndex column positions.
func TestLoadDictionary_ColumnIndex(t *testing.T) {
	s := DefaultSymSpell()
	// count word
	corpus := "42 alpha\n7 beta\n"
	require.True(t, s.LoadDictionary(strings.NewReader(corpus), 1, 0))
	require.Equal(t, 2, s.WordCount())
	got := s.Lookup("alpha", VerbosityTop)
	require.Len(t, got, 1)
	require.Equal(t, int64(42), got[0].Count)
}

// Twin of createDictionary from plain text.
func TestCreateDictionary(t *testing.T) {
	s := DefaultSymSpell()
	text := "Hello, world! café's test-case underscore_word"
	require.True(t, s.CreateDictionary(strings.NewReader(text)))
	// parseWords lowercases and extracts tokens
	require.GreaterOrEqual(t, s.WordCount(), 4)
	// "hello" from Hello
	got := s.Lookup("hello", VerbosityTop)
	require.NotEmpty(t, got)
	require.Equal(t, "hello", got[0].Term)
	// café with apostrophe form
	require.Contains(t, parseWords("café's"), "café's")
	require.Contains(t, parseWords("test-case"), "test-case")
	require.Contains(t, parseWords("underscore_word"), "underscore_word")
}

// Twin of parseWords.
func TestParseWords(t *testing.T) {
	require.Equal(t, []string{"hello", "world"}, parseWords("Hello, world!"))
	require.Equal(t, []string{"don't"}, parseWords("Don't"))
	require.Equal(t, []string{"a", "b"}, parseWords("A  B"))
	require.Empty(t, parseWords(""))
	require.Empty(t, parseWords("123 456")) // digits only — not \p{L}
}

// Twin of purgeBelowThresholdWords.
func TestPurgeBelowThresholdWords(t *testing.T) {
	s := NewSymSpell(16, 2, 7, 5)
	require.False(t, s.CreateDictionaryEntry("rare", 2, nil))
	require.Equal(t, 0, s.WordCount())
	require.NotEmpty(t, s.belowThresholdWords)
	s.PurgeBelowThresholdWords()
	require.Empty(t, s.belowThresholdWords)
	// can re-accumulate
	require.False(t, s.CreateDictionaryEntry("rare", 2, nil))
	require.NotEmpty(t, s.belowThresholdWords)
}

// Twin: LoadDictionaryFile missing path → false.
func TestLoadDictionaryFile_Missing(t *testing.T) {
	s := DefaultSymSpell()
	require.False(t, s.LoadDictionaryFile("/nonexistent/path/dict.txt", 0, 1))
}

// Twin: LoadDictionaryFile success from temp file.
func TestLoadDictionaryFile_OK(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "freq.txt")
	require.NoError(t, os.WriteFile(p, []byte("spell 10\ncheck 5\n"), 0o644))
	s := DefaultSymSpell()
	require.True(t, s.LoadDictionaryFile(p, 0, 1))
	require.Equal(t, 2, s.WordCount())
	got := s.Lookup("spel", VerbosityClosest)
	require.NotEmpty(t, got)
	require.Equal(t, "spell", got[0].Term)
}

// CommitStaged is alias of CommitStaging.
func TestCommitStagedAlias(t *testing.T) {
	s := DefaultSymSpell()
	st := NewSuggestionStage(8)
	require.True(t, s.CreateDictionaryEntry("alias", 1, st))
	s.CommitStaged(st)
	got := s.Lookup("alias", VerbosityTop)
	require.Len(t, got, 1)
}
