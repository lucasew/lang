package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenEnglishGlobalChunker_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenEnglishGlobalChunker applies Java constructor settings).
	r := strings.NewReader("Foo Bar\n")
	c, err := OpenEnglishGlobalChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling, "Java EnglishHybridDisambiguator.chunkerGlobal.setIgnoreSpelling(true)")
	require.False(t, c.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	require.Contains(t, c.Lines, "Foo Bar")
	// spelling_global has no /N expansion markers; plain phrase only.
	require.NotContains(t, c.Lines, "Foo Barn")
}

func TestEnglishGlobalChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverEnglishGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	a := EnglishGlobalChunker()
	b := EnglishGlobalChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official GlobalChunker entries (spelling_global.txt)
	require.Contains(t, a.Lines, "Microsoft Entra")
	require.Contains(t, a.Lines, "Google Maps")
	require.Contains(t, a.Lines, "picture alliance")
	require.Contains(t, a.Lines, "Picture Alliance")
	require.Contains(t, a.Lines, "acid house")
	// Titlecase of "acid house" is not a separate official line (titlecase variant only if allowTitlecase).
	require.NotContains(t, a.Lines, "Acid House")
	// Bare "New York" is not listed; multi-token New York Times is.
	require.NotContains(t, a.Lines, "New York")
	require.Contains(t, a.Lines, "New York Times")
	require.True(t, a.AddIgnoreSpelling)
	require.False(t, a.RemovePreviousTags)

	// Wired on DefaultEnglishHybridDisambiguator
	d := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker)
	require.Same(t, a, d.GlobalChunker)
}

func TestEnglishGlobalChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverEnglishGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	// Isolate GlobalChunker stage (do not re-claim multiwords Chunker / Rules).
	g := EnglishGlobalChunker()
	require.NotNil(t, g)
	d := &EnglishHybridDisambiguator{GlobalChunker: g}

	// Official multi-token phrases → IsIgnoredBySpeller on all content tokens.
	// tagForNotAddingTags: no invent open/close POS like normal multiwords.
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
		// Official listed titlecase form (separate line in spelling_global.txt)
		{[]string{"Picture", "Alliance"}, true, "Picture Alliance listed"},
		// allowFirstCapitalized=true (EN differs from FR GlobalChunker=false):
		// first-cap of a lowercase official entry IS generated.
		// "acid house" is listed; "Acid house" is not a separate line but is generated.
		{[]string{"acid", "house"}, true, "acid house exact"},
		{[]string{"Acid", "house"}, true, "Acid house first-cap allowed"},
		// allowTitlecase=false: full titlecase of lower official entry denied
		// ("Acid House" is not listed and not generated when allowTitlecase=false)
		{[]string{"Acid", "House"}, false, "Acid House titlecase denied"},
		// allowAllUppercase=true still accepts all-caps of official phrases
		{[]string{"GOOGLE", "MAPS"}, true, "GOOGLE MAPS all-upper"},
		// Wrong case on a proper-cased official entry: no match
		{[]string{"microsoft", "Entra"}, false, "microsoft Entra lower-first denied"},
	}
	for _, tc := range cases {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		requireAllContentIgnored(t, out, tc.want, tc.label)
		// tagForNotAddingTags: never invent multiword open/close angle POS
		got := contentPOSTags(out)
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"%s token[%d] tagForNotAddingTags must not invent angle POS: %v", tc.label, i, tags)
		}
	}

	// DefaultEnglishHybridDisambiguator wires GlobalChunker and still ignores official phrases.
	full := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	requireAllContentIgnored(t, out, true, "wired DefaultEnglishHybridDisambiguator Google Maps")
}
