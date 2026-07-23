package fr

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenFrenchMultiWordChunker_SettingsAndRemovePreviousTags(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenFrenchMultiWordChunker applies Java constructor settings).
	// Official fr/multiwords.txt uses #separatorRegExp=[\t;] and phrase;tag lines.
	r := strings.NewReader("#separatorRegExp=[\t;]\nFoo Bar;N m s\n")
	c, err := OpenFrenchMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.RemovePreviousTags, "Java FrenchHybridDisambiguator.setRemovePreviousTags(true)")
	require.False(t, c.AddIgnoreSpelling, "French multiwords chunker does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "Foo Bar;N m s")
}

func TestFrenchMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverFrenchMultiwords() == "" {
		t.Skip("official fr/multiwords.txt not discoverable")
	}
	a := FrenchMultiWordChunker()
	b := FrenchMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "home page;N f s")
	require.Contains(t, a.Lines, "Intel Core; Z m sp")
	require.Contains(t, a.Lines, "capture d'écran; N f s")
	require.Contains(t, a.Lines, "point presse; N m s")
	require.True(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewFrenchHybridDisambiguator
	d := NewFrenchHybridDisambiguator()
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

func TestFrenchMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverFrenchMultiwords() == "" {
		t.Skip("official fr/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage (do not re-claim GlobalChunker / Rules).
	c := FrenchMultiWordChunker()
	require.NotNil(t, c)
	d := &FrenchHybridDisambiguator{Chunker: c}

	// POS after setRemovePreviousTags(true) and AnalyzedToken.trim of the tag column:
	// - tags starting with "N " (after trim of leading whitespace from "; TAG"): first keeps "N …";
	//   subsequent use "J …" (Java MultiWordChunker.getNextPosTag French branch).
	// - other tags (e.g. "Z m sp"): plain TAG on every content token.
	// NC* Romance AQ0 path does NOT apply (French multiwords use "N …" / "Z …", not NC*).
	type phraseCase struct {
		parts    []string
		wantTags []string // one expected tag per content token, in order
		label    string
	}
	positives := []phraseCase{
		// home page;N f s
		{[]string{"home", "page"}, []string{"N f s", "J f s"}, "home page"},
		// Intel Core; Z m sp  (leading space after ';' trimmed by AnalyzedToken)
		{[]string{"Intel", "Core"}, []string{"Z m sp", "Z m sp"}, "Intel Core"},
		// capture d'écran; N f s  (leading space trimmed → "N f s" → J on subsequent)
		{[]string{"capture", "d'écran"}, []string{"N f s", "J f s"}, "capture d'écran"},
		// point presse; N m s
		{[]string{"point", "presse"}, []string{"N m s", "J m s"}, "point presse"},
		// allowFirstCapitalized=true: first-cap of lowercase official entry
		{[]string{"Home", "page"}, []string{"N f s", "J f s"}, "Home page first-cap"},
		{[]string{"Capture", "d'écran"}, []string{"N f s", "J f s"}, "Capture d'écran first-cap"},
		{[]string{"Point", "presse"}, []string{"N m s", "J m s"}, "Point presse first-cap"},
		// allowAllUppercase=true
		{[]string{"HOME", "PAGE"}, []string{"N f s", "J f s"}, "HOME PAGE all-upper"},
		{[]string{"INTEL", "CORE"}, []string{"Z m sp", "Z m sp"}, "INTEL CORE all-upper"},
		{[]string{"CAPTURE", "D'ÉCRAN"}, []string{"N f s", "J f s"}, "CAPTURE D'ÉCRAN all-upper"},
		{[]string{"POINT", "PRESSE"}, []string{"N m s", "J m s"}, "POINT PRESSE all-upper"},
	}
	for _, tc := range positives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		require.Len(t, got, len(tc.wantTags), "%s content token count", tc.label)
		for i, want := range tc.wantTags {
			require.True(t, hasExactPOS(got[i], want),
				"%s token[%d] want %q in %v", tc.label, i, want, got[i])
			require.False(t, hasAnyAnglePOS(got[i]),
				"%s token[%d] angle-bracket chunk tags should be flattened by removePreviousTags: %v",
				tc.label, i, got[i])
		}
	}

	// Negatives: non-listed sequences must not receive multiword POS
	negatives := []struct {
		parts []string
		label string
	}{
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed"},
		// allowTitlecase=false: titlecase of all-lower official entry denied
		{[]string{"Home", "Page"}, "Home Page titlecase denied"},
		// all-lower of title-cased official proper name is not generated as a variant
		{[]string{"intel", "core"}, "intel core all-lower denied"},
		// wrong middle casing of proper name when not listed
		{[]string{"Intel", "core"}, "Intel core mixed denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasExactPOS(tags, "N f s") || hasExactPOS(tags, "J f s") ||
				hasExactPOS(tags, "Z m sp") || hasExactPOS(tags, "N m s") ||
				hasExactPOS(tags, "J m s") || hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
	}

	// NewFrenchHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewFrenchHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("home", "page")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "N f s"), "wired hybrid home: %v", got[0])
	require.True(t, hasExactPOS(got[1], "J f s"), "wired hybrid page: %v", got[1])
}
