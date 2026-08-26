package ru

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenRussianMultiWordChunker_SettingsDefaults(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenRussianMultiWordChunker applies Java constructor defaults).
	// Official ru/multiwords.txt is phrase\ttag (default separator).
	r := strings.NewReader("до мажор\tNN:Masc\n")
	c, err := OpenRussianMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.False(t, c.RemovePreviousTags, "Russian multiwords does NOT setRemovePreviousTags")
	require.False(t, c.AddIgnoreSpelling, "Russian multiwords does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "до мажор\tNN:Masc")
}

func TestRussianMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverRussianMultiwords() == "" {
		t.Skip("official ru/multiwords.txt not discoverable")
	}
	a := RussianMultiWordChunker()
	b := RussianMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "до мажор\tNN:Masc")
	require.Contains(t, a.Lines, "до минор\tNN:Masc")
	require.Contains(t, a.Lines, "откуда ни возьмись\tFR")
	require.Contains(t, a.Lines, "пиши пропал\tFR")
	require.Contains(t, a.Lines, "черт возьми\tCONJ")
	require.Contains(t, a.Lines, "будь здоров\tADV")
	require.Contains(t, a.Lines, "в будущем\tADV")
	require.Contains(t, a.Lines, "до свидания\tADV")
	require.Contains(t, a.Lines, "в целом\tADV")
	require.False(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewRussianHybridDisambiguator
	d := NewRussianHybridDisambiguator()
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

func TestRussianMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverRussianMultiwords() == "" {
		t.Skip("official ru/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage.
	c := RussianMultiWordChunker()
	require.NotNil(t, c)
	d := &RussianHybridDisambiguator{Chunker: c}

	// Defaults: no setRemovePreviousTags → open/close angle tags on first/last content tokens.
	type phraseCase struct {
		parts    []string
		wantOpen string
		wantEnd  string
		label    string
	}
	positives := []phraseCase{
		{[]string{"до", "мажор"}, "<NN:Masc>", "</NN:Masc>", "до мажор"},
		{[]string{"до", "минор"}, "<NN:Masc>", "</NN:Masc>", "до минор"},
		{[]string{"пиши", "пропал"}, "<FR>", "</FR>", "пиши пропал"},
		{[]string{"черт", "возьми"}, "<CONJ>", "</CONJ>", "черт возьми"},
		{[]string{"будь", "здоров"}, "<ADV>", "</ADV>", "будь здоров"},
		{[]string{"в", "будущем"}, "<ADV>", "</ADV>", "в будущем"},
		{[]string{"до", "свидания"}, "<ADV>", "</ADV>", "до свидания"},
		{[]string{"в", "целом"}, "<ADV>", "</ADV>", "в целом"},
		{[]string{"откуда", "ни", "возьмись"}, "<FR>", "</FR>", "откуда ни возьмись"},
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

	// Official listed first-capitalized forms (explicit multiwords lines, not flag invent)
	{
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("До", "мажор")))
		got := contentPOSTags(out)
		require.True(t, hasExactPOS(got[0], "<NN:Masc>"), "До мажор listed open: %v", got[0])
		require.True(t, hasExactPOS(got[1], "</NN:Masc>"), "До мажор listed close: %v", got[1])
	}

	// Negatives: non-listed sequences and casing denied by defaults F/F/F
	// (будь здоров has no listed capital form; ДО МАЖОР not listed)
	negatives := []struct {
		parts []string
		label string
	}{
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed"},
		// allowFirstCapitalized=false and not listed
		{[]string{"Будь", "здоров"}, "Будь здоров first-cap denied"},
		// allowAllUppercase=false
		{[]string{"ДО", "МАЖОР"}, "ДО МАЖОР all-upper denied"},
		// allowTitlecase=false
		{[]string{"До", "Мажор"}, "До Мажор titlecase denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
	}

	// NewRussianHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewRussianHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("до", "мажор")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<NN:Masc>"), "wired hybrid до: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</NN:Masc>"), "wired hybrid мажор: %v", got[1])
}
