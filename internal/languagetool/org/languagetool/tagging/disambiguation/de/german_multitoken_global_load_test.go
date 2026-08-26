package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenGermanMultitokenGlobal_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenGermanMultitokenGlobal applies Java constructor settings).
	r := strings.NewReader("Foo Bar\n")
	c, err := OpenGermanMultitokenGlobal(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling)
	require.Contains(t, c.Lines, "Foo Bar")
	// spelling_global has no German /N expansion markers; plain phrase only.
	require.NotContains(t, c.Lines, "Foo Barn")
}

func TestGermanMultitokenGlobal_ProcessCachedOfficial(t *testing.T) {
	if DiscoverGermanMultitokenGlobal() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	a := GermanMultitokenGlobal()
	b := GermanMultitokenGlobal()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official MultitokenGlobal entries (spelling_global.txt)
	require.Contains(t, a.Lines, "Microsoft Entra")
	require.Contains(t, a.Lines, "Google Maps")
	require.Contains(t, a.Lines, "P. Sherman 42 Wallaby Way")
	require.Contains(t, a.Lines, "picture alliance")
	// Bare "New York" is not listed; multi-token New York Times is.
	require.NotContains(t, a.Lines, "New York")
	require.Contains(t, a.Lines, "New York Times")

	// Wired on NewGermanRuleDisambiguator
	d := NewGermanRuleDisambiguator()
	require.NotNil(t, d.MultitokenGlobal)
	require.Same(t, a, d.MultitokenGlobal)
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

func TestGermanMultitokenGlobal_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverGermanMultitokenGlobal() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	// Isolate MultitokenGlobal stage (do not re-claim MultitokenIgnore).
	g := GermanMultitokenGlobal()
	require.NotNil(t, g)
	d := &GermanRuleDisambiguator{MultitokenGlobal: g}

	// Official multi-token phrases → IsIgnoredBySpeller on all content tokens.
	cases := []struct {
		parts []string
		want  bool
		label string
	}{
		{[]string{"Microsoft", "Entra"}, true, "Microsoft Entra"},
		{[]string{"Google", "Maps"}, true, "Google Maps"},
		{[]string{"New", "York", "Times"}, true, "New York Times"},
		// Phrase with punctuation as listed in spelling_global.txt
		{[]string{"P.", "Sherman", "42", "Wallaby", "Way"}, true, "P. Sherman 42 Wallaby Way"},
		// Negative: random non-listed multi-token phrase
		{[]string{"Zxqwv", "Plmnb"}, false, "random non-listed"},
		// Official casing: "picture alliance" matches as written
		{[]string{"picture", "alliance"}, true, "picture alliance exact"},
		// allowFirstCapitalized=false: first-cap of a lowercase official entry
		// is NOT generated — "Picture alliance" must not match MultitokenGlobal.
		// (MultitokenIgnore with allowFirstCapitalized=true would match.)
		{[]string{"Picture", "alliance"}, false, "Picture alliance first-cap denied"},
		// allowAllUppercase=true still accepts all-caps of official phrases
		{[]string{"GOOGLE", "MAPS"}, true, "GOOGLE MAPS all-upper"},
		// Wrong case on a proper-cased official entry: no match
		{[]string{"microsoft", "Entra"}, false, "microsoft Entra lower-first denied"},
	}
	for _, tc := range cases {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		requireAllContentIgnored(t, out, tc.want, tc.label)
	}

	// NewGermanRuleDisambiguator wires MultitokenGlobal and still ignores official phrases.
	full := NewGermanRuleDisambiguator()
	require.NotNil(t, full.MultitokenGlobal)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	requireAllContentIgnored(t, out, true, "wired NewGermanRuleDisambiguator Google Maps")
}
