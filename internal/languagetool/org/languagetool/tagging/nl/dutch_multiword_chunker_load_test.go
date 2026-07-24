package nl

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenDutchMultiWordChunker_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenDutchMultiWordChunker applies Java constructor settings).
	// Official nl/multiwords.txt is phrase-only (no tab/tag column).
	r := strings.NewReader("Foo Bar\n")
	c, err := OpenDutchMultiWordChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling)
	require.Contains(t, c.Lines, "Foo Bar")
}

func TestDutchMultiWordChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverDutchMultiwords() == "" {
		t.Skip("official nl/multiwords.txt not discoverable")
	}
	a := DutchMultiWordChunker()
	b := DutchMultiWordChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official multiwords phrases (phrase-only lines, no tabs)
	require.Contains(t, a.Lines, "A Clockwork Orange")
	require.Contains(t, a.Lines, "A fortiori")
	require.Contains(t, a.Lines, "A priori")
	require.Contains(t, a.Lines, "New York Knicks")
	require.Contains(t, a.Lines, "a fortiori")
	require.Contains(t, a.Lines, "a priori")

	// Wired on NewDutchHybridDisambiguator
	d := NewDutchHybridDisambiguator()
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

func requireAllContentIgnored(t *testing.T, out *languagetool.AnalyzedSentence, want bool, label string) {
	t.Helper()
	toks := out.GetTokens()
	for i, tr := range toks {
		if i == 0 || tr.IsWhitespace() {
			continue
		}
		if want {
			require.True(t, tr.IsIgnoredBySpeller(), "%s token[%d]=%q", label, i, tr.GetToken())
		} else {
			require.False(t, tr.IsIgnoredBySpeller(), "%s token[%d]=%q should NOT ignore", label, i, tr.GetToken())
		}
	}
}

func TestDutchMultiWordChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverDutchMultiwords() == "" {
		t.Skip("official nl/multiwords.txt not discoverable")
	}
	// Isolate Chunker stage (do not re-claim GlobalChunker / Rules).
	c := DutchMultiWordChunker()
	require.NotNil(t, c)
	d := &DutchHybridDisambiguator{Chunker: c}

	cases := []struct {
		parts []string
		want  bool
		label string
	}{
		{[]string{"A", "Clockwork", "Orange"}, true, "A Clockwork Orange"},
		{[]string{"A", "fortiori"}, true, "A fortiori"},
		{[]string{"A", "priori"}, true, "A priori"},
		{[]string{"New", "York", "Knicks"}, true, "New York Knicks"},
		// Official lowercase entries
		{[]string{"a", "fortiori"}, true, "a fortiori exact"},
		{[]string{"a", "priori"}, true, "a priori exact"},
		// Negatives: random non-listed multi-token
		{[]string{"Zxqwv", "Plmnb"}, false, "random non-listed"},
		// allowFirstCapitalized=true: first-cap of lowercase official entry
		// (also listed as "A fortiori" officially — either path matches)
		{[]string{"A", "fortiori"}, true, "A fortiori first-cap allowed"},
		// allowAllUppercase=true
		{[]string{"NEW", "YORK", "KNICKS"}, true, "NEW YORK KNICKS all-upper"},
		{[]string{"A", "FORTIORI"}, true, "A FORTIORI all-upper"},
		// Wrong case on proper-cased official: middle lower denied when not listed
		{[]string{"new", "york", "knicks"}, false, "new york knicks lower denied"},
	}
	for _, tc := range cases {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		requireAllContentIgnored(t, out, tc.want, tc.label)
	}

	// NewDutchHybridDisambiguator wires Chunker and still ignores official phrases.
	full := NewDutchHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("New", "York", "Knicks")))
	requireAllContentIgnored(t, out, true, "wired NewDutchHybridDisambiguator New York Knicks")
}
