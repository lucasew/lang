package fr

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenFrenchGlobalChunker_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenFrenchGlobalChunker applies Java constructor settings).
	r := strings.NewReader("Foo Bar\n")
	c, err := OpenFrenchGlobalChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling, "Java FrenchHybridDisambiguator.chunkerGlobal.setIgnoreSpelling(true)")
	require.False(t, c.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	require.Contains(t, c.Lines, "Foo Bar")
	// spelling_global has no /N expansion markers; plain phrase only.
	require.NotContains(t, c.Lines, "Foo Barn")
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

func TestFrenchGlobalChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverFrenchGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	a := FrenchGlobalChunker()
	b := FrenchGlobalChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official GlobalChunker entries (spelling_global.txt)
	require.Contains(t, a.Lines, "Microsoft Entra")
	require.Contains(t, a.Lines, "Google Maps")
	require.Contains(t, a.Lines, "picture alliance")
	// Bare "New York" is not listed; multi-token New York Times is.
	require.NotContains(t, a.Lines, "New York")
	require.Contains(t, a.Lines, "New York Times")
	require.True(t, a.AddIgnoreSpelling)
	require.False(t, a.RemovePreviousTags)

	// Wired on NewFrenchHybridDisambiguator
	d := NewFrenchHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker)
	require.Same(t, a, d.GlobalChunker)
}

func TestFrenchGlobalChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverFrenchGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	// Isolate GlobalChunker stage (do not re-claim multiwords Chunker / Rules).
	g := FrenchGlobalChunker()
	require.NotNil(t, g)
	d := &FrenchHybridDisambiguator{GlobalChunker: g}

	// Official multi-token phrases → IsIgnoredBySpeller on all content tokens.
	cases := []struct {
		parts []string
		want  bool
		label string
	}{
		{[]string{"Microsoft", "Entra"}, true, "Microsoft Entra"},
		{[]string{"Google", "Maps"}, true, "Google Maps"},
		{[]string{"New", "York", "Times"}, true, "New York Times"},
		// Negative: random non-listed multi-token phrase
		{[]string{"Zxqwv", "Plmnb"}, false, "random non-listed"},
		// Official casing: "picture alliance" matches as written
		{[]string{"picture", "alliance"}, true, "picture alliance exact"},
		// allowFirstCapitalized=false: first-cap of a lowercase official entry
		// is NOT generated — "Picture alliance" must not match GlobalChunker.
		// (French multiwords with allowFirstCapitalized=true would match.)
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

	// NewFrenchHybridDisambiguator wires GlobalChunker and still ignores official phrases.
	full := NewFrenchHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	requireAllContentIgnored(t, out, true, "wired NewFrenchHybridDisambiguator Google Maps")
}
