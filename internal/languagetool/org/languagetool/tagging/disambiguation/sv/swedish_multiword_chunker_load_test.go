package sv

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenSwedishMultiWordChunker_SettingsDefaults(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenSwedishMultiWordChunker applies Java constructor defaults).
	// Official sv/multiwords.txt is phrase\ttag (default separator).
	r := strings.NewReader("en passant\tNN:OF:SIN:NOM:UTR\n")
	c, err := OpenSwedishMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.False(t, c.RemovePreviousTags, "Swedish multiwords does NOT setRemovePreviousTags")
	require.False(t, c.AddIgnoreSpelling, "Swedish multiwords does NOT setIgnoreSpelling")
	require.Contains(t, c.Lines, "en passant\tNN:OF:SIN:NOM:UTR")
}

func TestSwedishMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverSwedishMultiwords() == "" {
		t.Skip("official sv/multiwords.txt not discoverable")
	}
	a := SwedishMultiWordChunker()
	b := SwedishMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (from multiwords.txt; not invented)
	require.Contains(t, a.Lines, "...\tELLIPS")
	require.Contains(t, a.Lines, "en passant\tNN:OF:SIN:NOM:UTR")
	require.Contains(t, a.Lines, "Sri Lanka\tPM:NOM")
	require.Contains(t, a.Lines, "ad hoc\tAB")
	require.Contains(t, a.Lines, "vice versa\tAB")
	require.Contains(t, a.Lines, "New York\tPM:NOM")
	require.False(t, a.RemovePreviousTags)
	require.False(t, a.AddIgnoreSpelling)

	// Wired on NewSwedishHybridDisambiguator
	d := NewSwedishHybridDisambiguator()
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
// for no-space multiwords such as official "...\tELLIPS".
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

func TestSwedishMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverSwedishMultiwords() == "" {
		t.Skip("official sv/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage.
	c := SwedishMultiWordChunker()
	require.NotNil(t, c)
	d := &SwedishHybridDisambiguator{Chunker: c}

	// Defaults: no setRemovePreviousTags → open/close angle tags on first/last content tokens.
	type phraseCase struct {
		parts    []string
		wantOpen string
		wantEnd  string
		label    string
	}
	positives := []phraseCase{
		{[]string{"en", "passant"}, "<NN:OF:SIN:NOM:UTR>", "</NN:OF:SIN:NOM:UTR>", "en passant"},
		{[]string{"Sri", "Lanka"}, "<PM:NOM>", "</PM:NOM>", "Sri Lanka"},
		{[]string{"ad", "hoc"}, "<AB>", "</AB>", "ad hoc"},
		{[]string{"vice", "versa"}, "<AB>", "</AB>", "vice versa"},
		{[]string{"New", "York"}, "<PM:NOM>", "</PM:NOM>", "New York"},
		{[]string{"Costa", "Rica"}, "<PM:NOM>", "</PM:NOM>", "Costa Rica"},
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

	// No-space multiword: official "...\tELLIPS"
	{
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokensNoSpace(".", ".", ".")))
		got := contentPOSTags(out)
		require.Len(t, got, 3, "...")
		require.True(t, hasExactPOS(got[0], "<ELLIPS>"), "first . open: %v", got[0])
		require.True(t, hasExactPOS(got[2], "</ELLIPS>"), "last . close: %v", got[2])
		require.False(t, hasAnyAnglePOS(got[1]), "middle . no angle: %v", got[1])
	}

	// Spaced ellipsis: official ". . .\tELLIPS"
	{
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(".", ".", ".")))
		got := contentPOSTags(out)
		require.Len(t, got, 3, ". . .")
		require.True(t, hasExactPOS(got[0], "<ELLIPS>"), "spaced first . open: %v", got[0])
		require.True(t, hasExactPOS(got[2], "</ELLIPS>"), "spaced last . close: %v", got[2])
	}

	// Negatives: non-listed sequences and casing denied by defaults F/F/F
	negatives := []struct {
		parts []string
		label string
	}{
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed"},
		// allowFirstCapitalized=false
		{[]string{"En", "passant"}, "En passant first-cap denied"},
		// allowAllUppercase=false
		{[]string{"EN", "PASSANT"}, "EN PASSANT all-upper denied"},
		// allowTitlecase=false
		{[]string{"En", "Passant"}, "En Passant titlecase denied"},
	}
	for _, tc := range negatives {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"%s token[%d] should have no multiword POS, got %v", tc.label, i, tags)
		}
	}

	// NewSwedishHybridDisambiguator wires Chunker and still tags official phrases.
	full := NewSwedishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("en", "passant")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<NN:OF:SIN:NOM:UTR>"), "wired hybrid en: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</NN:OF:SIN:NOM:UTR>"), "wired hybrid passant: %v", got[1])
}
