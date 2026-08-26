package sr

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenSerbianMultiWordChunker_SettingsDefaults(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenSerbianMultiWordChunker applies Java constructor defaults).
	// Official sr/multiwords.txt is empty today; format is still phrase\ttag.
	r := strings.NewReader("foo bar\tX\n")
	c, err := OpenSerbianMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.False(t, c.RemovePreviousTags, "Serbian multiwords does NOT setRemovePreviousTags")
	require.False(t, c.AddIgnoreSpelling, "Serbian multiwords does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "foo bar\tX")

	// F,F,F: allowFirstCapitalized/allowAllUppercase/allowTitlecase false — first-cap denied.
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Foo", "bar")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	for i, tags := range got {
		require.False(t, hasAnyAnglePOS(tags), "first-cap denied token[%d]: %v", i, tags)
	}
	// Exact case matches (settings apply to real maps when phrases exist).
	out = c.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("foo", "bar")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<X>"), "exact-case open: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</X>"), "exact-case close: %v", got[1])
	// All-caps denied.
	out = c.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("FOO", "BAR")))
	got = contentPOSTags(out)
	for i, tags := range got {
		require.False(t, hasAnyAnglePOS(tags), "all-caps denied token[%d]: %v", i, tags)
	}
	// Titlecase denied.
	out = c.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Foo", "Bar")))
	got = contentPOSTags(out)
	for i, tags := range got {
		require.False(t, hasAnyAnglePOS(tags), "titlecase denied token[%d]: %v", i, tags)
	}
}

func TestSerbianMultiWordChunker_ProcessCachedOfficialEmpty(t *testing.T) {
	if DiscoverSerbianMultiwords() == "" {
		t.Skip("official sr/multiwords.txt not discoverable")
	}
	a := SerbianMultiWordChunker()
	b := SerbianMultiWordChunker()
	require.NotNil(t, a, "Java constructs MultiWordChunker even when file is empty")
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords.txt is empty (0 lines) → no phrase entries after loadWords.
	// Do not invent multiword entries.
	require.Empty(t, a.Lines, "official sr/multiwords.txt is empty; Lines must stay empty")
	require.False(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewSerbianHybridDisambiguator
	d := NewSerbianHybridDisambiguator()
	require.NotNil(t, d.Chunker, "Java final field chunker = MultiWordChunker(/sr/multiwords.txt)")
	require.Same(t, a, d.Chunker)
}

func TestSerbianMultiWordChunker_DisambiguateOfficialIsNoOp(t *testing.T) {
	if DiscoverSerbianMultiwords() == "" {
		t.Skip("official sr/multiwords.txt not discoverable")
	}
	c := SerbianMultiWordChunker()
	require.NotNil(t, c)

	// Sample surfaces (incl. SR XML isolation tokens): empty multiwords → no invent open/close.
	for _, parts := range [][]string{
		{"XX", "vek"},
		{"I", "II"},
		{"foo", "bar"},
		{"random", "phrase"},
		{"Zdravo", "svete"},
	} {
		out := c.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(parts...)))
		for i, tags := range contentPOSTags(out) {
			require.False(t, hasAnyAnglePOS(tags),
				"%v multiword no invent token[%d]: %v", parts, i, tags)
		}
		// No setIgnoreSpelling on multiwords.
		for i, tr := range out.GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%v multiword must not ignore spelling on %q", parts, tr.GetToken())
		}
	}
}
