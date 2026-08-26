package es

// Outcome twins for SpanishHybridDisambiguator full stage order:
// Java SpanishHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
// i.e. spelling_global → /es/multiwords.txt (setRemovePreviousTags) → XmlRuleDisambiguator(lang, true).
// Official spanish.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/global/XML stage leaves).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireESHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverSpanishGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	if DiscoverSpanishMultiwords() == "" {
		t.Skip("official es/multiwords.txt not discoverable")
	}
	if DiscoverSpanishDisambiguationXML() == "" {
		t.Skip("official es/disambiguation.xml not discoverable")
	}
	if DiscoverGlobalDisambiguationXML() == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
}

// TestNewSpanishHybridDisambiguator_WiresAllThreeStages proves Java constructor
// eagerly builds GlobalChunker, multiwords Chunker, and XmlRuleDisambiguator when
// the same official resources Java loads are present.
func TestNewSpanishHybridDisambiguator_WiresAllThreeStages(t *testing.T) {
	requireESHybridResources(t)

	g := SpanishGlobalChunker()
	mw := SpanishMultiWordChunker()
	xml := SpanishXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewSpanishHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker, "chunkerGlobal = MultiWordChunker.getInstance(/spelling_global.txt, …, NPCN000)")
	require.NotNil(t, d.Chunker, "chunker = MultiWordChunker.getInstance(/es/multiwords.txt, …) + setRemovePreviousTags(true)")
	require.NotNil(t, d.Rules, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.Same(t, g, d.GlobalChunker)
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Spanish settings (no invent): global no ignore/remove; multiwords removePreviousTags, no ignore.
	require.False(t, g.AddIgnoreSpelling)
	require.False(t, g.RemovePreviousTags)
	require.True(t, mw.RemovePreviousTags)
	require.False(t, mw.AddIgnoreSpelling)
}

// TestSpanishHybridDisambiguator_OrderGlobalThenMultiwordThenXML proves stage
// isolation vs full Java order with Java-visible POS / ignore_spelling outcomes.
func TestSpanishHybridDisambiguator_OrderGlobalThenMultiwordThenXML(t *testing.T) {
	requireESHybridResources(t)

	g := SpanishGlobalChunker()
	mw := SpanishMultiWordChunker()
	xml := SpanishXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	// Isolation hybrids (only one stage) vs full order.
	onlyGlobal := &SpanishHybridDisambiguator{GlobalChunker: g}
	onlyMulti := &SpanishHybridDisambiguator{Chunker: mw}
	onlyXML := &SpanishHybridDisambiguator{Rules: xml}
	// Reverse chunk stages for order proof (multiword then global — opposite of Java).
	reverseChunks := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		// multiword first, then global (opposite of Java)
		return g.Disambiguate(mw.Disambiguate(sent))
	}
	full := NewSpanishHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Global-only phrase: "Google Maps" in spelling_global, not multiwords ---
	// Global alone → open/close <NPCN000></NPCN000>
	// Multiword alone → no multiword POS
	// Full hybrid (global→multiword removePreviousTags) → plain NPCN000 NPCN000
	// Wrong chunk order (multiword→global) → angle tags remain (global last, no flatten after)
	{
		label := "Google Maps"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		require.Len(t, gotG, 2, label)
		require.True(t, hasExactPOS(gotG[0], "<NPCN000>"), "%s global-only open: %v", label, gotG[0])
		require.True(t, hasExactPOS(gotG[1], "</NPCN000>"), "%s global-only close: %v", label, gotG[1])

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		for i, tags := range gotM {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasExactPOS(tags, "<NPCN000>") ||
				hasExactPOS(tags, "</NPCN000>") || hasAnyAnglePOS(tags),
				"%s multiword-only token[%d] must have no global/multiword POS, got %v", label, i, tags)
		}

		// XML alone: no POS invent for this surface; no angle tags
		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no chunk POS, got %v", label, i, tags)
		}

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "NPCN000"),
			"%s full hybrid Google flattened (global then removePreviousTags): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPCN000"),
			"%s full hybrid Maps flattened: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must not keep angle tags: %v %v", label, gotFull[0], gotFull[1])

		// Reverse chunk order differs: multiword no-op then global leaves angles
		gotRev := contentPOSTags(reverseChunks(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<NPCN000>"),
			"%s reverse order leaves open angle (proves multiword-after-global needed): %v", label, gotRev[0])
		require.True(t, hasExactPOS(gotRev[1], "</NPCN000>"),
			"%s reverse order leaves close angle: %v", label, gotRev[1])

		// Never ignore spelling from ES GlobalChunker / multiwords (no setIgnoreSpelling)
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid token %q must not ignore spelling via chunkers", label, tr.GetToken())
		}
	}

	// --- (2) Multiword-only phrase: "time lapse" in es/multiwords.txt, not spelling_global ---
	// Official line: time lapse;NCMS000 → after removePreviousTags: NCMS000 AQ0MS0
	{
		label := "time lapse"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("time", "lapse"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "NCMS000") || hasExactPOS(tags, "AQ0MS0") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "NCMS000"), "%s multiword-only time: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "AQ0MS0"), "%s multiword-only lapse: %v", label, gotM[1])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]))

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "NCMS000"), "%s full hybrid time: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "AQ0MS0"), "%s full hybrid lapse: %v", label, gotFull[1])
	}

	// --- (3) Shared phrase: "Peter Pan" listed in BOTH global and multiwords ---
	// Official multiwords: Peter Pan;NPMNSP0
	// Official global: Peter Pan → NPCN000 open/close
	// Java order: global tags first, multiword adds NPMNSP0, removePreviousTags → NPMNSP0
	// (multiword tag preferred over low-priority NPCN000 when both present).
	{
		label := "Peter Pan"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotG[0], "<NPCN000>"), "%s global-only: %v", label, gotG[0])
		require.True(t, hasExactPOS(gotG[1], "</NPCN000>"), "%s global-only: %v", label, gotG[1])

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NPMNSP0"), "%s multiword-only: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NPMNSP0"), "%s multiword-only: %v", label, gotM[1])

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPMNSP0"),
			"%s full hybrid prefers multiword NPMNSP0 after global+removePreviousTags: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPMNSP0"),
			"%s full hybrid Pan: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must flatten angles: %v %v", label, gotFull[0], gotFull[1])
		// Must not leave only NPCN000 from global without multiword tag.
		// (plain NPCN000 would mean multiword stage was skipped after global flatten)
	}

	// --- (4) XML-only effects: es ABBREVIATIONS short date + global proper noun ---
	// Chunkers do not set ignore_spelling; XML stage does.
	{
		// es rule: case_sensitive short date 15E
		sent15 := tokenSentence("15E")
		outG := onlyGlobal.Disambiguate(sent15)
		outM := onlyMulti.Disambiguate(sent15)
		requireNotIgnored(t, outG, "15E")
		requireNotIgnored(t, outM, "15E")
		requireIgnored(t, onlyXML.Disambiguate(sent15), "15E")
		requireIgnored(t, full.Disambiguate(sent15), "15E")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// UNIDADES_SI marker min (XML after chunkers still fires)
		sentMin := tokenSentence("30", "min")
		requireIgnored(t, full.Disambiguate(sentMin), "min")
	}
}

// TestSpanishHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkPOS proves
// XML last does not wipe multiword POS on official multiword surfaces.
func TestSpanishHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkPOS(t *testing.T) {
	requireESHybridResources(t)
	full := NewSpanishHybridDisambiguator()
	// Peter Pan multiword POS survives XML stage.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPMNSP0"), "Peter after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMNSP0"), "Pan after full order: %v", got[1])
	// Google Maps flattened global POS survives XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	got = contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPCN000"), "Google after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPCN000"), "Maps after full order: %v", got[1])
}
