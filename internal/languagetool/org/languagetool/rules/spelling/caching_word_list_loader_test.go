package spelling

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/stretchr/testify/require"
)

func withFreshWordListLoader(t *testing.T) *CachingWordListLoader {
	t.Helper()
	prev := GetDataBroker()
	clearWordListCache()
	SetDataBroker(nil)
	t.Cleanup(func() {
		clearWordListCache()
		SetDataBroker(prev)
	})
	return NewCachingWordListLoader()
}

// Twin: CachingWordListLoader CacheLoader line rules — comments, trailing #, empty, Java trim.
func TestParseWordListLines(t *testing.T) {
	in := "# comment\nfoo #trail\nbar\n\n  baz  \n"
	got, err := ParseWordListLines(strings.NewReader(in))
	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar", "baz"}, got)
}

// Twin: CachingWordListLoader uses String.trim (≤U+0020), not Unicode TrimSpace.
// NBSP-only suffix must remain part of the entry until comment strip; NBSP is not trimmed.
func TestParseWordListLines_JavaTrimNotUnicode(t *testing.T) {
	// "foo" + NBSP after trim of ASCII spaces stays "foo\u00a0" then no # → kept with NBSP.
	in := "  foo\u00a0\nbar #x\n"
	got, err := ParseWordListLines(strings.NewReader(in))
	require.NoError(t, err)
	require.Equal(t, []string{"foo\u00a0", "bar"}, got)
}

// Twin: raw-line empty/# skip before trim; value may still be empty after trim/substringBefore.
// Java: "   " and "  #x" are not skipped by isEmpty/startsWith("#"), then add "".
func TestParseWordListLines_EmptyAfterTrimStillAdded(t *testing.T) {
	in := "   \n  #only-comment-after-spaces\nreal\n"
	got, err := ParseWordListLines(strings.NewReader(in))
	require.NoError(t, err)
	require.Equal(t, []string{"", "", "real"}, got)
}

// Twin: resourceExists false → empty list (not error), and result is cached.
func TestCachingWordListLoader_MissingResource(t *testing.T) {
	l := withFreshWordListLoader(t)
	m := broker.NewMapResourceDataBroker()
	SetDataBroker(m)

	words := l.LoadWords("missing/words.txt")
	require.Empty(t, words)
	require.NotNil(t, words) // empty list, not nil

	// cache hit: even if resource appears later, missing path stays empty (Guava cache semantics)
	m.Resource["missing/words.txt"] = "late\n"
	words2 := l.LoadWords("missing/words.txt")
	require.Empty(t, words2)
}

// Twin: resourceExists true → getFromResourceDirAsLines + parse.
func TestCachingWordListLoader_MapBroker(t *testing.T) {
	l := withFreshWordListLoader(t)
	m := broker.NewMapResourceDataBroker()
	m.Resource["en/spelling.txt"] = "#c\nfoo #trail\nbar\n\n"
	SetDataBroker(m)

	words := l.LoadWords("en/spelling.txt")
	require.Equal(t, []string{"foo", "bar"}, words)

	// same path returns same cached slice (Java cache.getUnchecked)
	words2 := l.LoadWords("en/spelling.txt")
	require.Equal(t, words, words2)
}

// Twin: second LoadWords same path does not re-read; mutating broker content is ignored.
func TestCachingWordListLoader_CacheNoReread(t *testing.T) {
	l := withFreshWordListLoader(t)
	m := broker.NewMapResourceDataBroker()
	m.Resource["en/cache.txt"] = "alpha\nbeta\n"
	SetDataBroker(m)

	first := l.LoadWords("en/cache.txt")
	require.Equal(t, []string{"alpha", "beta"}, first)

	// Mutate underlying resource (Guava would still serve cached list until expire).
	m.Resource["en/cache.txt"] = "gamma\n"
	second := l.LoadWords("en/cache.txt")
	require.Equal(t, []string{"alpha", "beta"}, second)
	require.Equal(t, first, second)

	// Distinct path still loads fresh content.
	m.Resource["en/other.txt"] = "delta\n"
	other := l.LoadWords("en/other.txt")
	require.Equal(t, []string{"delta"}, other)
}

// Twin: nil DataBroker behaves like resourceExists false for all paths.
func TestCachingWordListLoader_NilBroker(t *testing.T) {
	l := withFreshWordListLoader(t)
	require.Empty(t, l.LoadWords("any/path.txt"))
}

// Twin: instances share the process-level cache (Java static LoadingCache).
func TestCachingWordListLoader_SharedCacheAcrossInstances(t *testing.T) {
	_ = withFreshWordListLoader(t)
	m := broker.NewMapResourceDataBroker()
	m.Resource["shared.txt"] = "one\n"
	SetDataBroker(m)

	a := NewCachingWordListLoader()
	b := NewCachingWordListLoader()
	w1 := a.LoadWords("shared.txt")
	// Mutate after first instance loaded
	m.Resource["shared.txt"] = "two\n"
	w2 := b.LoadWords("shared.txt")
	require.Equal(t, []string{"one"}, w1)
	require.Equal(t, []string{"one"}, w2)
}
