package nl

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenDutchGlobalChunker_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenDutchGlobalChunker applies Java constructor settings).
	r := strings.NewReader("Foo Bar\n")
	c, err := OpenDutchGlobalChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling)
	require.Contains(t, c.Lines, "Foo Bar")
	// spelling_global has no Dutch /N expansion markers; plain phrase only.
	require.NotContains(t, c.Lines, "Foo Barn")
}

func TestDutchGlobalChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverDutchGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	a := DutchGlobalChunker()
	b := DutchGlobalChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official GlobalChunker entries (spelling_global.txt)
	require.Contains(t, a.Lines, "Microsoft Entra")
	require.Contains(t, a.Lines, "Google Maps")
	require.Contains(t, a.Lines, "picture alliance")
	// Bare "New York" is not listed; multi-token New York Times is.
	require.NotContains(t, a.Lines, "New York")
	require.Contains(t, a.Lines, "New York Times")

	// Wired on NewDutchHybridDisambiguator
	d := NewDutchHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker)
	require.Same(t, a, d.GlobalChunker)
}

func TestDutchGlobalChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverDutchGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	// Isolate GlobalChunker stage (do not re-claim multiwords Chunker / Rules).
	g := DutchGlobalChunker()
	require.NotNil(t, g)
	d := &DutchHybridDisambiguator{GlobalChunker: g}

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
		// (Dutch multiwords with allowFirstCapitalized=true would match.)
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

	// NewDutchHybridDisambiguator wires GlobalChunker and still ignores official phrases.
	full := NewDutchHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	requireAllContentIgnored(t, out, true, "wired NewDutchHybridDisambiguator Google Maps")
}
