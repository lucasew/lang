package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenPortugueseGlobalChunker_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenPortugueseGlobalChunker applies Java constructor settings).
	r := strings.NewReader("Foo Bar\n")
	c, err := OpenPortugueseGlobalChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling, "Portuguese GlobalChunker setIgnoreSpelling(true)")
	require.False(t, c.RemovePreviousTags, "Portuguese GlobalChunker does NOT setRemovePreviousTags")
	require.Contains(t, c.Lines, "Foo Bar")
	// spelling_global has no /N expansion markers; plain phrase only.
	require.NotContains(t, c.Lines, "Foo Barn")

	// DefaultTag NPCN000 → open/close angle tags (not tagForNotAddingTags, not empty).
	d := &PortugueseHybridDisambiguator{GlobalChunker: c}
	out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Foo", "Bar")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<NPCN000>"), "open tag: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</NPCN000>"), "close tag: %v", got[1])
	requireAllContentIgnored(t, out, true, "Foo Bar ignore spelling")
}

func TestPortugueseGlobalChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverPortugueseGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	a := PortugueseGlobalChunker()
	b := PortugueseGlobalChunker()
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

	// Wired on NewPortugueseHybridDisambiguator
	d := NewPortugueseHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker)
	require.Same(t, a, d.GlobalChunker)
	// Rules stay nil until XmlRule sector
	require.Nil(t, d.Rules)
	// Chunker still wired from multiwords when discoverable
	if PortugueseMultiWordChunker() != nil {
		require.NotNil(t, d.Chunker)
		require.Same(t, PortugueseMultiWordChunker(), d.Chunker)
	}
}

func TestPortugueseGlobalChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverPortugueseGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	// Isolate GlobalChunker stage (do not re-claim multiwords Chunker / Rules).
	g := PortugueseGlobalChunker()
	require.NotNil(t, g)
	d := &PortugueseHybridDisambiguator{GlobalChunker: g}

	// Official multi-token phrases → open/close <NPCN000></NPCN000> on first/last content tokens
	// + IsIgnoredBySpeller true when matched (setIgnoreSpelling).
	type phraseCase struct {
		parts     []string
		wantMatch bool
		label     string
	}
	cases := []phraseCase{
		{[]string{"Microsoft", "Entra"}, true, "Microsoft Entra"},
		{[]string{"Google", "Maps"}, true, "Google Maps"},
		{[]string{"New", "York", "Times"}, true, "New York Times"},
		// Official casing: "picture alliance" matches as written
		// (file also lists "Picture Alliance" as a separate line — not a lettercase variant).
		{[]string{"picture", "alliance"}, true, "picture alliance exact"},
		{[]string{"Picture", "Alliance"}, true, "Picture Alliance official line"},
		// allowAllUppercase=true still accepts all-caps of official phrases
		{[]string{"GOOGLE", "MAPS"}, true, "GOOGLE MAPS all-upper"},
		// all-lower official multi-token without titlecase/first-cap twin lines
		{[]string{"acid", "house"}, true, "acid house exact"},
		// Negatives
		{[]string{"Zxqwv", "Plmnb"}, false, "random non-listed"},
		// allowFirstCapitalized=false: first-cap of a lowercase official entry
		// is NOT generated — "Acid house" must not match GlobalChunker.
		{[]string{"Acid", "house"}, false, "Acid house first-cap denied"},
		// allowTitlecase=true is inert without allowFirstCapitalized (Java nests titlecase):
		// "Acid House" is not generated as a variant of "acid house".
		{[]string{"Acid", "House"}, false, "Acid House titlecase denied without first-cap"},
		// Wrong case on a proper-cased official entry: no match
		{[]string{"microsoft", "Entra"}, false, "microsoft Entra lower-first denied"},
	}
	for _, tc := range cases {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		got := contentPOSTags(out)
		require.Len(t, got, len(tc.parts), "%s content token count", tc.label)
		if tc.wantMatch {
			require.True(t, hasExactPOS(got[0], "<NPCN000>"),
				"%s first want <NPCN000> in %v", tc.label, got[0])
			last := len(got) - 1
			require.True(t, hasExactPOS(got[last], "</NPCN000>"),
				"%s last want </NPCN000> in %v", tc.label, got[last])
			// Middle content tokens (if any) need not carry NPCN000 open/close;
			// MultiWordChunker only annotates first and last of the space span.
		} else {
			for i, tags := range got {
				require.False(t, hasExactPOS(tags, "<NPCN000>") || hasExactPOS(tags, "</NPCN000>") ||
					hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
					"%s token[%d] should have no GlobalChunker POS, got %v", tc.label, i, tags)
			}
		}
		requireAllContentIgnored(t, out, tc.wantMatch, tc.label+" ignore spelling")
	}

	// NewPortugueseHybridDisambiguator wires GlobalChunker first; multiwords Chunker then
	// runs setRemovePreviousTags → <NPCN000></NPCN000> flattens to plain NPCN000 NPCN000
	// (Java order: chunkerGlobal → chunker(multiwords with removePreviousTags)).
	// Use Microsoft Entra: official in spelling_global, not listed in pt/multiwords.txt
	// (Google Maps is also in multiwords and would be re-tagged there).
	full := NewPortugueseHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	if full.Chunker != nil {
		require.True(t, hasExactPOS(got[0], "NPCN000"), "wired hybrid Microsoft (flattened): %v", got[0])
		require.True(t, hasExactPOS(got[1], "NPCN000"), "wired hybrid Entra (flattened): %v", got[1])
		require.False(t, hasAnyAnglePOS(got[0]) || hasAnyAnglePOS(got[1]),
			"removePreviousTags should flatten angle tags: %v %v", got[0], got[1])
	} else {
		require.True(t, hasExactPOS(got[0], "<NPCN000>"), "wired hybrid Microsoft: %v", got[0])
		require.True(t, hasExactPOS(got[1], "</NPCN000>"), "wired hybrid Entra: %v", got[1])
	}
	requireAllContentIgnored(t, out, true, "wired hybrid Microsoft Entra ignore")
}
