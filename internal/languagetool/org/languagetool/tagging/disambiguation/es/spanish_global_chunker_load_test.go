package es

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenSpanishGlobalChunker_SettingsAndNoIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenSpanishGlobalChunker applies Java constructor settings).
	r := strings.NewReader("Foo Bar\n")
	c, err := OpenSpanishGlobalChunker(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.False(t, c.AddIgnoreSpelling, "Spanish GlobalChunker does NOT setIgnoreSpelling")
	require.False(t, c.RemovePreviousTags, "Spanish GlobalChunker does NOT setRemovePreviousTags")
	require.Contains(t, c.Lines, "Foo Bar")
	// spelling_global has no /N expansion markers; plain phrase only.
	require.NotContains(t, c.Lines, "Foo Barn")

	// DefaultTag NPCN000 → open/close angle tags (not tagForNotAddingTags, not empty).
	d := &SpanishHybridDisambiguator{GlobalChunker: c}
	out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Foo", "Bar")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<NPCN000>"), "open tag: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</NPCN000>"), "close tag: %v", got[1])
	// No ignore-spelling side effect
	for i, tr := range out.GetTokens() {
		if i == 0 || tr.IsWhitespace() {
			continue
		}
		require.False(t, tr.IsIgnoredBySpeller(), "token %q must not ignore spelling", tr.GetToken())
	}
}

func TestSpanishGlobalChunker_ProcessCachedOfficial(t *testing.T) {
	if DiscoverSpanishGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	a := SpanishGlobalChunker()
	b := SpanishGlobalChunker()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official GlobalChunker entries (spelling_global.txt)
	require.Contains(t, a.Lines, "Microsoft Entra")
	require.Contains(t, a.Lines, "Google Maps")
	require.Contains(t, a.Lines, "picture alliance")
	// Bare "New York" is not listed; multi-token New York Times is.
	require.NotContains(t, a.Lines, "New York")
	require.Contains(t, a.Lines, "New York Times")
	require.False(t, a.AddIgnoreSpelling)
	require.False(t, a.RemovePreviousTags)

	// Wired on NewSpanishHybridDisambiguator
	d := NewSpanishHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker)
	require.Same(t, a, d.GlobalChunker)
	// Rules still left for separate sector
	require.Nil(t, d.Rules)
}

func TestSpanishGlobalChunker_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverSpanishGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	// Isolate GlobalChunker stage (do not re-claim multiwords Chunker / Rules).
	g := SpanishGlobalChunker()
	require.NotNil(t, g)
	d := &SpanishHybridDisambiguator{GlobalChunker: g}

	// Official multi-token phrases → open/close <NPCN000></NPCN000> on first/last content tokens.
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
		{[]string{"picture", "alliance"}, true, "picture alliance exact"},
		// allowAllUppercase=true still accepts all-caps of official phrases
		{[]string{"GOOGLE", "MAPS"}, true, "GOOGLE MAPS all-upper"},
		// Negatives
		{[]string{"Zxqwv", "Plmnb"}, false, "random non-listed"},
		// allowFirstCapitalized=false: first-cap of a lowercase official entry
		// is NOT generated — "Picture alliance" must not match GlobalChunker.
		{[]string{"Picture", "alliance"}, false, "Picture alliance first-cap denied"},
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
		// Never ignore spelling from ES GlobalChunker
		for i, tr := range out.GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s token %q must not ignore spelling", tc.label, tr.GetToken())
		}
	}

	// NewSpanishHybridDisambiguator wires GlobalChunker first; multiwords Chunker then
	// runs setRemovePreviousTags → <NPCN000></NPCN000> flattens to plain NPCN000 NPCN000
	// (Java order: chunkerGlobal → chunker(multiwords with removePreviousTags)).
	full := NewSpanishHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	if full.Chunker != nil {
		require.True(t, hasExactPOS(got[0], "NPCN000"), "wired hybrid Google (flattened): %v", got[0])
		require.True(t, hasExactPOS(got[1], "NPCN000"), "wired hybrid Maps (flattened): %v", got[1])
		require.False(t, hasAnyAnglePOS(got[0]) || hasAnyAnglePOS(got[1]),
			"removePreviousTags should flatten angle tags: %v %v", got[0], got[1])
	} else {
		require.True(t, hasExactPOS(got[0], "<NPCN000>"), "wired hybrid Google: %v", got[0])
		require.True(t, hasExactPOS(got[1], "</NPCN000>"), "wired hybrid Maps: %v", got[1])
	}
}
