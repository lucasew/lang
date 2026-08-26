package ar

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenArabicMultiWordChunker_SettingsDefaults(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenArabicMultiWordChunker applies Java constructor defaults).
	// Official ar/multiwords.txt is comment-only today; format is still phrase\ttag.
	r := strings.NewReader("foo bar\tX\n")
	c, err := OpenArabicMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.False(t, c.RemovePreviousTags, "Arabic multiwords does NOT setRemovePreviousTags")
	require.False(t, c.AddIgnoreSpelling, "Arabic multiwords does NOT setIgnoreSpelling")
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

func TestArabicMultiWordChunker_ProcessCachedOfficialEmpty(t *testing.T) {
	if DiscoverArabicMultiwords() == "" {
		t.Skip("official ar/multiwords.txt not discoverable")
	}
	a := ArabicMultiWordChunker()
	b := ArabicMultiWordChunker()
	require.NotNil(t, a, "Java constructs MultiWordChunker.getInstance even when file is comment-only")
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords.txt is comment-only → no phrase entries after loadWords.
	// Do not invent multiword entries.
	require.Empty(t, a.Lines, "official ar/multiwords.txt is comment-only; Lines must stay empty")
	require.False(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewArabicHybridDisambiguator
	d := NewArabicHybridDisambiguator()
	require.NotNil(t, d.Chunker, "Java final field chunker = MultiWordChunker.getInstance(...)")
	require.Same(t, a, d.Chunker)
}

// multiwordTokens builds SENT_START + alternating content/space tokens for MultiWordChunker.
func multiwordTokens(parts ...string) []*languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for i, p := range parts {
		if i > 0 {
			toks = append(toks, languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)))
		}
		toks = append(toks, languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(p, nil, nil)))
	}
	return toks
}

func contentPOSTags(out *languagetool.AnalyzedSentence) [][]string {
	var all [][]string
	for i, tr := range out.GetTokens() {
		if i == 0 || tr.IsWhitespace() {
			continue
		}
		var tags []string
		for _, r := range tr.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
		all = append(all, tags)
	}
	return all
}

func hasExactPOS(tags []string, want string) bool {
	for _, p := range tags {
		if p == want {
			return true
		}
	}
	return false
}

func hasAnyAnglePOS(tags []string) bool {
	for _, p := range tags {
		if strings.Contains(p, "<") {
			return true
		}
	}
	return false
}

func TestArabicMultiWordChunker_DisambiguateOfficialIsNoOp(t *testing.T) {
	if DiscoverArabicMultiwords() == "" {
		t.Skip("official ar/multiwords.txt not discoverable")
	}
	c := ArabicMultiWordChunker()
	require.NotNil(t, c)

	// Sample surfaces (incl. Arabic XML isolation tokens): empty multiwords → no invent open/close.
	for _, parts := range [][]string{
		{"قد", "عامل"},
		{"في", "عامل"},
		{"foo", "bar"},
		{"ثلاثة", "وثلاثون"},
		{"random", "phrase"},
	} {
		out := c.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"official empty multiwords must not invent angle POS on %v token[%d]: %v", parts, i, tags)
		}
		for i, tr := range out.GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"no setIgnoreSpelling on multiwords; token %q", tr.GetToken())
		}
	}

	// NewArabicHybridDisambiguator wires Chunker and still does not invent multiword POS.
	full := NewArabicHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("قد", "عامل")))
	got := contentPOSTags(out)
	for i, tags := range got {
		require.False(t, hasAnyAnglePOS(tags),
			"wired hybrid multiword no invent token[%d]: %v", i, tags)
	}
}
