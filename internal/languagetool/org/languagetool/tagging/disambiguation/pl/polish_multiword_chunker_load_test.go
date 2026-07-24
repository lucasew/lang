package pl

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenPolishMultiWordChunker_SettingsDefaults(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenPolishMultiWordChunker applies Java constructor defaults).
	// Official pl/multiwords.txt is phrase\ttag (default separator).
	r := strings.NewReader("to znaczy\tTO_ZNACZY\n")
	c, err := OpenPolishMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.False(t, c.RemovePreviousTags, "Polish multiwords does NOT setRemovePreviousTags")
	require.False(t, c.AddIgnoreSpelling, "Polish multiwords does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "to znaczy\tTO_ZNACZY")
}

func TestPolishMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverPolishMultiwords() == "" {
		t.Skip("official pl/multiwords.txt not discoverable")
	}
	a := PolishMultiWordChunker()
	b := PolishMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "...\tELLIPSIS")
	require.Contains(t, a.Lines, "to znaczy\tTO_ZNACZY")
	require.Contains(t, a.Lines, "to jest\tTO_JEST")
	require.Contains(t, a.Lines, "z uwagi na\tPREP:ACC")
	require.Contains(t, a.Lines, "co do\tPREP:GEN")
	require.Contains(t, a.Lines, "bez mała\tADV")
	require.False(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewPolishHybridDisambiguator
	d := NewPolishHybridDisambiguator()
	require.NotNil(t, d.Chunker)
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

// multiwordTokensNoSpace builds SENT_START + adjacent content tokens (no whitespace)
// for no-space multiwords such as official "...\tELLIPSIS".
func multiwordTokensNoSpace(parts ...string) []*languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for _, p := range parts {
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

func TestPolishMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverPolishMultiwords() == "" {
		t.Skip("official pl/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage.
	c := PolishMultiWordChunker()
	require.NotNil(t, c)
	d := &PolishHybridDisambiguator{Chunker: c}

	// Defaults: no setRemovePreviousTags → open/close angle tags on first/last content tokens.
	type phraseCase struct {
		parts    []string
		wantOpen string
		wantEnd  string
		label    string
	}
	positives := []phraseCase{
		{[]string{"to", "znaczy"}, "<TO_ZNACZY>", "</TO_ZNACZY>", "to znaczy"},
		{[]string{"to", "jest"}, "<TO_JEST>", "</TO_JEST>", "to jest"},
		{[]string{"co", "do"}, "<PREP:GEN>", "</PREP:GEN>", "co do"},
		{[]string{"z", "uwagi", "na"}, "<PREP:ACC>", "</PREP:ACC>", "z uwagi na"},
		{[]string{"bez", "mała"}, "<ADV>", "</ADV>", "bez mała"},
		{[]string{"w", "związku", "z"}, "<PREP:INST>", "</PREP:INST>", "w związku z"},
	}
	for _, tc := range positives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		require.Len(t, got, len(tc.parts), "%s content token count", tc.label)
		require.True(t, hasExactPOS(got[0], tc.wantOpen),
			"%s first token want %q in %v", tc.label, tc.wantOpen, got[0])
		require.True(t, hasExactPOS(got[len(got)-1], tc.wantEnd),
			"%s last token want %q in %v", tc.label, tc.wantEnd, got[len(got)-1])
		for i := 1; i < len(got)-1; i++ {
			require.False(t, hasAnyAnglePOS(got[i]),
				"%s interior token[%d] should have no angle multiword POS, got %v", tc.label, i, got[i])
		}
		// No setIgnoreSpelling
		for i, tr := range out.GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s token %q must not ignore spelling", tc.label, tr.GetToken())
		}
	}

	// No-space multiword: official "...\tELLIPSIS"
	{
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", ".")))
		got := contentPOSTags(out)
		require.Len(t, got, 3, "...")
		require.True(t, hasExactPOS(got[0], "<ELLIPSIS>"), "first . open: %v", got[0])
		require.True(t, hasExactPOS(got[2], "</ELLIPSIS>"), "last . close: %v", got[2])
		require.False(t, hasAnyAnglePOS(got[1]), "middle . no angle: %v", got[1])
	}

	// Negatives: non-listed sequences and casing denied by defaults F/F/F
	negatives := []struct {
		parts []string
		label string
	}{
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed"},
		// allowFirstCapitalized=false
		{[]string{"To", "znaczy"}, "To znaczy first-cap denied"},
		{[]string{"Co", "do"}, "Co do first-cap denied"},
		// allowAllUppercase=false
		{[]string{"TO", "ZNACZY"}, "TO ZNACZY all-upper denied"},
		// allowTitlecase=false
		{[]string{"To", "Znaczy"}, "To Znaczy titlecase denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
	}

	// NewPolishHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewPolishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("to", "znaczy")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<TO_ZNACZY>"), "wired hybrid to: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</TO_ZNACZY>"), "wired hybrid znaczy: %v", got[1])
}
