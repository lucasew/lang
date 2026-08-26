package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenCatalanMultiWordChunker_SettingsAndRemovePreviousTags(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenCatalanMultiWordChunker applies Java constructor settings).
	// Official ca/multiwords.txt uses #separatorRegExp=[\t;] and phrase;tag lines.
	r := strings.NewReader("#separatorRegExp=[\t;]\nFoo Bar;NPMNSP0\n")
	c, err := OpenCatalanMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.RemovePreviousTags, "Java CatalanHybridDisambiguator.setRemovePreviousTags(true)")
	require.False(t, c.AddIgnoreSpelling, "Catalan multiwords chunker does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "Foo Bar;NPMNSP0")
}

func TestCatalanMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverCatalanMultiwords() == "" {
		t.Skip("official ca/multiwords.txt not discoverable")
	}
	a := CatalanMultiWordChunker()
	b := CatalanMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "Agnes Callard;NPFSSP0")
	require.Contains(t, a.Lines, "uilleann pipes;NCFN000")
	require.Contains(t, a.Lines, "comme il faut;LOC_ADV")
	require.Contains(t, a.Lines, "El Correo Catalán;NPMSO00")
	require.True(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewCatalanHybridDisambiguator
	d := NewCatalanHybridDisambiguator()
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

func TestCatalanMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverCatalanMultiwords() == "" {
		t.Skip("official ca/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage (do not re-claim GlobalChunker / Multitoken / Rules).
	c := CatalanMultiWordChunker()
	require.NotNil(t, c)
	d := &CatalanHybridDisambiguator{Chunker: c}

	// POS after setRemovePreviousTags(true):
	// multi-token NP tags → plain TAG on every content token of the span.
	// multi-token NC* tags → first token keeps NC…; subsequent use AQ0+gender/number+0
	// (Java MultiWordChunker.getNextPosTag for Romance NC prefixes).
	// multi-token LOC_ADV → plain LOC_ADV on every content token.
	type phraseCase struct {
		parts    []string
		wantTags []string // one expected tag per content token, in order
		label    string
	}
	positives := []phraseCase{
		{[]string{"Agnes", "Callard"}, []string{"NPFSSP0", "NPFSSP0"}, "Agnes Callard"},
		{[]string{"uilleann", "pipes"}, []string{"NCFN000", "AQ0FN0"}, "uilleann pipes"},
		{[]string{"comme", "il", "faut"}, []string{"LOC_ADV", "LOC_ADV", "LOC_ADV"}, "comme il faut"},
		{[]string{"El", "Correo", "Catalán"}, []string{"NPMSO00", "NPMSO00", "NPMSO00"}, "El Correo Catalán"},
		// allowFirstCapitalized=true: first-cap of lowercase official entry
		{[]string{"Uilleann", "pipes"}, []string{"NCFN000", "AQ0FN0"}, "Uilleann pipes first-cap"},
		{[]string{"Comme", "il", "faut"}, []string{"LOC_ADV", "LOC_ADV", "LOC_ADV"}, "Comme il faut first-cap"},
		// allowAllUppercase=true
		{[]string{"AGNES", "CALLARD"}, []string{"NPFSSP0", "NPFSSP0"}, "AGNES CALLARD all-upper"},
		{[]string{"UILLEANN", "PIPES"}, []string{"NCFN000", "AQ0FN0"}, "UILLEANN PIPES all-upper"},
		{[]string{"COMME", "IL", "FAUT"}, []string{"LOC_ADV", "LOC_ADV", "LOC_ADV"}, "COMME IL FAUT all-upper"},
		{[]string{"EL", "CORREO", "CATALÁN"}, []string{"NPMSO00", "NPMSO00", "NPMSO00"}, "EL CORREO CATALÁN all-upper"},
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
		{[]string{"Uilleann", "Pipes"}, "Uilleann Pipes titlecase denied"},
		// all-lower of title-cased official proper name is not generated as a variant
		{[]string{"agnes", "callard"}, "agnes callard all-lower denied"},
		// wrong middle casing of proper name when not listed
		{[]string{"Agnes", "callard"}, "Agnes callard mixed denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasExactPOS(tags, "NPFSSP0") || hasExactPOS(tags, "NCFN000") ||
				hasExactPOS(tags, "AQ0FN0") || hasExactPOS(tags, "LOC_ADV") ||
				hasExactPOS(tags, "NPMSO00") || hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
	}

	// NewCatalanHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewCatalanHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Agnes", "Callard")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPFSSP0"), "wired hybrid Agnes: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPFSSP0"), "wired hybrid Callard: %v", got[1])
}
