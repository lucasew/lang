package es

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenSpanishMultiWordChunker_SettingsAndRemovePreviousTags(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenSpanishMultiWordChunker applies Java constructor settings).
	// Official es/multiwords.txt uses #separatorRegExp=[\t;] and phrase;tag lines.
	r := strings.NewReader("#separatorRegExp=[\t;]\nFoo Bar;NPMNSP0\n")
	c, err := OpenSpanishMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.RemovePreviousTags, "Java SpanishHybridDisambiguator.setRemovePreviousTags(true)")
	require.False(t, c.AddIgnoreSpelling, "Spanish multiwords chunker does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "Foo Bar;NPMNSP0")
}

func TestSpanishMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverSpanishMultiwords() == "" {
		t.Skip("official es/multiwords.txt not discoverable")
	}
	a := SpanishMultiWordChunker()
	b := SpanishMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "Peter Pan;NPMNSP0")
	require.Contains(t, a.Lines, "time lapse;NCMS000")
	require.Contains(t, a.Lines, "folia cerebelosa;NCFS000")
	require.Contains(t, a.Lines, "jet ski;NCMS000")
	require.Contains(t, a.Lines, "Jet Ski;NCMS000")
	require.Contains(t, a.Lines, "tomate cherry;NCMS000")
	require.True(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewSpanishHybridDisambiguator
	d := NewSpanishHybridDisambiguator()
	require.NotNil(t, d.Chunker)
	require.Same(t, a, d.Chunker)
	// This sector does not wire GlobalChunker / Rules
	require.Nil(t, d.GlobalChunker)
	require.Nil(t, d.Rules)
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

func TestSpanishMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverSpanishMultiwords() == "" {
		t.Skip("official es/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage (do not re-claim GlobalChunker / Rules).
	c := SpanishMultiWordChunker()
	require.NotNil(t, c)
	d := &SpanishHybridDisambiguator{Chunker: c}

	// POS after setRemovePreviousTags(true):
	// multi-token NP tags → plain TAG on every content token of the span.
	// multi-token NC* tags → first token keeps NC…; subsequent use AQ0+gender/number+0
	// (Java MultiWordChunker.getNextPosTag for Romance NC prefixes).
	type phraseCase struct {
		parts    []string
		wantTags []string // one expected tag per content token, in order
		label    string
	}
	positives := []phraseCase{
		{[]string{"Peter", "Pan"}, []string{"NPMNSP0", "NPMNSP0"}, "Peter Pan"},
		{[]string{"time", "lapse"}, []string{"NCMS000", "AQ0MS0"}, "time lapse"},
		{[]string{"folia", "cerebelosa"}, []string{"NCFS000", "AQ0FS0"}, "folia cerebelosa"},
		{[]string{"tomate", "cherry"}, []string{"NCMS000", "AQ0MS0"}, "tomate cherry"},
		{[]string{"jet", "ski"}, []string{"NCMS000", "AQ0MS0"}, "jet ski"},
		{[]string{"Jet", "Ski"}, []string{"NCMS000", "AQ0MS0"}, "Jet Ski exact listed"},
		// allowFirstCapitalized=true: first-cap of lowercase official entry
		{[]string{"Tomate", "cherry"}, []string{"NCMS000", "AQ0MS0"}, "Tomate cherry first-cap"},
		{[]string{"Time", "lapse"}, []string{"NCMS000", "AQ0MS0"}, "Time lapse first-cap"},
		{[]string{"Folia", "cerebelosa"}, []string{"NCFS000", "AQ0FS0"}, "Folia cerebelosa first-cap"},
		// allowAllUppercase=true
		{[]string{"PETER", "PAN"}, []string{"NPMNSP0", "NPMNSP0"}, "PETER PAN all-upper"},
		{[]string{"TIME", "LAPSE"}, []string{"NCMS000", "AQ0MS0"}, "TIME LAPSE all-upper"},
		{[]string{"JET", "SKI"}, []string{"NCMS000", "AQ0MS0"}, "JET SKI all-upper"},
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
		// allowTitlecase=false: middle tokens lower when official is title-cased "Peter Pan"
		{[]string{"peter", "pan"}, "peter pan all-lower denied"},
		// wrong middle casing of proper name when not listed
		{[]string{"Peter", "pan"}, "Peter pan mixed denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasExactPOS(tags, "NPMNSP0") || hasExactPOS(tags, "NCMS000") ||
				hasExactPOS(tags, "AQ0MS0") || hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
	}

	// NewSpanishHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewSpanishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPMNSP0"), "wired hybrid Peter: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMNSP0"), "wired hybrid Pan: %v", got[1])
}
