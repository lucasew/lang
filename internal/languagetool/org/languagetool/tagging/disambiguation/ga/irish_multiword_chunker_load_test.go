package ga

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenIrishMultiWordChunker_SettingsDefaults(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenIrishMultiWordChunker applies Java constructor defaults).
	// Official ga/multiwords.txt is phrase\ttag (default separator).
	r := strings.NewReader("foo bar\tPrep:Cmpd\n")
	c, err := OpenIrishMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.False(t, c.RemovePreviousTags, "Irish multiwords does NOT setRemovePreviousTags")
	require.False(t, c.AddIgnoreSpelling, "Irish multiwords does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "foo bar\tPrep:Cmpd")
}

func TestIrishMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverIrishMultiwords() == "" {
		t.Skip("official ga/multiwords.txt not discoverable")
	}
	a := IrishMultiWordChunker()
	b := IrishMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "ar ais\tAdv:Dir")
	require.Contains(t, a.Lines, "a lán\tSubst:Noun:Sg")
	require.Contains(t, a.Lines, "ar feadh\tPrep:Cmpd")
	require.Contains(t, a.Lines, "chun go\tConj:Subord")
	require.Contains(t, a.Lines, "de bharr\tPrep:Cmpd")
	require.Contains(t, a.Lines, "mar a déarfá\tCmc")
	require.Contains(t, a.Lines, "ar chor ar bith\tAdv:Gn")
	require.False(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewIrishHybridDisambiguator
	d := NewIrishHybridDisambiguator()
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

func TestIrishMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverIrishMultiwords() == "" {
		t.Skip("official ga/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage (Rules not claimed this sector).
	c := IrishMultiWordChunker()
	require.NotNil(t, c)
	d := &IrishHybridDisambiguator{Chunker: c}

	// Defaults: no setRemovePreviousTags → open/close angle tags on first/last content tokens.
	// (Java MultiWordChunker without removePreviousTags.)
	type phraseCase struct {
		parts    []string
		wantOpen string // first content token
		wantEnd  string // last content token
		label    string
	}
	positives := []phraseCase{
		// ar ais → Adv:Dir
		{[]string{"ar", "ais"}, "<Adv:Dir>", "</Adv:Dir>", "ar ais"},
		// a lán → Subst:Noun:Sg
		{[]string{"a", "lán"}, "<Subst:Noun:Sg>", "</Subst:Noun:Sg>", "a lán"},
		// ar feadh → Prep:Cmpd
		{[]string{"ar", "feadh"}, "<Prep:Cmpd>", "</Prep:Cmpd>", "ar feadh"},
		// chun go → Conj:Subord
		{[]string{"chun", "go"}, "<Conj:Subord>", "</Conj:Subord>", "chun go"},
		// de bharr → Prep:Cmpd
		{[]string{"de", "bharr"}, "<Prep:Cmpd>", "</Prep:Cmpd>", "de bharr"},
		// mar a déarfá → Cmc (3 content tokens: open first, empty middle, close last)
		{[]string{"mar", "a", "déarfá"}, "<Cmc>", "</Cmc>", "mar a déarfá"},
		// ar chor ar bith → Adv:Gn (4 content tokens)
		{[]string{"ar", "chor", "ar", "bith"}, "<Adv:Gn>", "</Adv:Gn>", "ar chor ar bith"},
	}
	for _, tc := range positives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		require.Len(t, got, len(tc.parts), "%s content token count", tc.label)
		require.True(t, hasExactPOS(got[0], tc.wantOpen),
			"%s first token want %q in %v", tc.label, tc.wantOpen, got[0])
		require.True(t, hasExactPOS(got[len(got)-1], tc.wantEnd),
			"%s last token want %q in %v", tc.label, tc.wantEnd, got[len(got)-1])
		// Interior content tokens (if any) should not get open/close multiword tags
		for i := 1; i < len(got)-1; i++ {
			require.False(t, hasAnyAnglePOS(got[i]),
				"%s interior token[%d] should have no angle multiword POS, got %v", tc.label, i, got[i])
		}
	}

	// Negatives: non-listed sequences and casing denied by defaults F/F/F
	negatives := []struct {
		parts []string
		label string
	}{
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed"},
		// allowFirstCapitalized=false: first-cap of lowercase official entry denied
		{[]string{"Ar", "ais"}, "Ar ais first-cap denied"},
		{[]string{"A", "lán"}, "A lán first-cap denied"},
		{[]string{"Chun", "go"}, "Chun go first-cap denied"},
		// allowAllUppercase=false
		{[]string{"AR", "AIS"}, "AR AIS all-upper denied"},
		{[]string{"DE", "BHARR"}, "DE BHARR all-upper denied"},
		// allowTitlecase=false
		{[]string{"Ar", "Ais"}, "Ar Ais titlecase denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
	}

	// NewIrishHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewIrishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<Adv:Dir>"), "wired hybrid ar: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</Adv:Dir>"), "wired hybrid ais: %v", got[1])
}
