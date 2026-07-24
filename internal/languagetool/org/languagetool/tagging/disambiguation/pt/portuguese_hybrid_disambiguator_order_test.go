package pt

// Outcome twins for PortugueseHybridDisambiguator full stage order:
// Java PortugueseHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
// i.e. spelling_global → /pt/multiwords.txt (setRemovePreviousTags + setIgnoreSpelling)
//     → XmlRuleDisambiguator(lang, true).
// Official portuguese.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/global/XML stage leaves).
//
// Differs from Spanish hybrid:
//   - allowAllUppercase=true on both chunkers; allowTitlecase=true on both
//   - setIgnoreSpelling(true) on BOTH chunkerGlobal and chunker (ES: neither)

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requirePTHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverPortugueseGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	if DiscoverPortugueseMultiwords() == "" {
		t.Skip("official pt/multiwords.txt not discoverable")
	}
	if DiscoverPortugueseDisambiguationXML() == "" {
		t.Skip("official pt/disambiguation.xml not discoverable")
	}
	if DiscoverGlobalDisambiguationXML() == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
}

// TestNewPortugueseHybridDisambiguator_WiresAllThreeStages proves Java constructor
// eagerly builds GlobalChunker, multiwords Chunker, and XmlRuleDisambiguator when
// the same official resources Java loads are present.
func TestNewPortugueseHybridDisambiguator_WiresAllThreeStages(t *testing.T) {
	requirePTHybridResources(t)

	g := PortugueseGlobalChunker()
	mw := PortugueseMultiWordChunker()
	xml := PortugueseXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewPortugueseHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker, "chunkerGlobal = MultiWordChunker.getInstance(/spelling_global.txt, false, true, true, NPCN000)")
	require.NotNil(t, d.Chunker, "chunker = MultiWordChunker.getInstance(/pt/multiwords.txt, true, true, true) + setRemovePreviousTags + setIgnoreSpelling")
	require.NotNil(t, d.Rules, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.Same(t, g, d.GlobalChunker)
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Portuguese settings (no invent): ignoreSpelling on BOTH chunkers; multiwords removePreviousTags.
	require.True(t, g.AddIgnoreSpelling, "Portuguese chunkerGlobal.setIgnoreSpelling(true)")
	require.False(t, g.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	require.True(t, mw.RemovePreviousTags, "Portuguese chunker.setRemovePreviousTags(true)")
	require.True(t, mw.AddIgnoreSpelling, "Portuguese chunker.setIgnoreSpelling(true)")
}

// TestPortugueseHybridDisambiguator_OrderGlobalThenMultiwordThenXML proves stage
// isolation vs full Java order with Java-visible POS / ignore_spelling outcomes.
func TestPortugueseHybridDisambiguator_OrderGlobalThenMultiwordThenXML(t *testing.T) {
	requirePTHybridResources(t)

	g := PortugueseGlobalChunker()
	mw := PortugueseMultiWordChunker()
	xml := PortugueseXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	// Isolation hybrids (only one stage) vs full order.
	onlyGlobal := &PortugueseHybridDisambiguator{GlobalChunker: g}
	onlyMulti := &PortugueseHybridDisambiguator{Chunker: mw}
	onlyXML := &PortugueseHybridDisambiguator{Rules: xml}
	// Reverse chunk stages for order proof (multiword then global — opposite of Java).
	reverseChunks := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return g.Disambiguate(mw.Disambiguate(sent))
	}
	full := NewPortugueseHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Global-only phrase: "Microsoft Entra" in spelling_global, not pt/multiwords ---
	// Global alone → open/close <NPCN000></NPCN000> + ignore_spelling
	// Multiword alone → no multiword POS, no ignore
	// Full hybrid (global→multiword removePreviousTags) → plain NPCN000 NPCN000 + ignore
	// Wrong chunk order (multiword→global) → angle tags remain (global last, no flatten after)
	{
		label := "Microsoft Entra"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		require.Len(t, gotG, 2, label)
		require.True(t, hasExactPOS(gotG[0], "<NPCN000>"), "%s global-only open: %v", label, gotG[0])
		require.True(t, hasExactPOS(gotG[1], "</NPCN000>"), "%s global-only close: %v", label, gotG[1])
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global-only ignore")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		for i, tags := range gotM {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasExactPOS(tags, "<NPCN000>") ||
				hasExactPOS(tags, "</NPCN000>") || hasAnyAnglePOS(tags),
				"%s multiword-only token[%d] must have no global/multiword POS, got %v", label, i, tags)
		}
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), false, label+" multiword-only no ignore")

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
			"%s full hybrid Microsoft flattened (global then removePreviousTags): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPCN000"),
			"%s full hybrid Entra flattened: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must not keep angle tags: %v %v", label, gotFull[0], gotFull[1])
		// PT: both chunkers setIgnoreSpelling — full hybrid ignores matched global phrases.
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")

		// Reverse chunk order differs: multiword no-op then global leaves angles
		gotRev := contentPOSTags(reverseChunks(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<NPCN000>"),
			"%s reverse order leaves open angle (proves multiword-after-global needed): %v", label, gotRev[0])
		require.True(t, hasExactPOS(gotRev[1], "</NPCN000>"),
			"%s reverse order leaves close angle: %v", label, gotRev[1])
	}

	// --- (2) Multiword-only phrase: "Bin Laden" in pt/multiwords.txt, not spelling_global ---
	// Official line: Bin Laden\tNPMS000 → after removePreviousTags: NPMS000 NPMS000 + ignore
	{
		label := "Bin Laden"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Bin", "Laden"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "NPMS000") || hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no ignore")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "NPMS000"), "%s multiword-only Bin: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NPMS000"), "%s multiword-only Laden: %v", label, gotM[1])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]))
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), true, label+" multiword-only ignore")

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "NPMS000"), "%s full hybrid Bin: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPMS000"), "%s full hybrid Laden: %v", label, gotFull[1])
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")
	}

	// Multiword NC phrase (Romance getNextPosTag): fair play;NCMS000 → NCMS000 AQ0MS0
	{
		label := "fair play"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("fair", "play"))
		}
		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "NCMS000") || hasExactPOS(tags, "AQ0MS0") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NCMS000"), "%s multiword-only fair: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "AQ0MS0"), "%s multiword-only play: %v", label, gotM[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NCMS000"), "%s full hybrid fair: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "AQ0MS0"), "%s full hybrid play: %v", label, gotFull[1])
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")
	}

	// --- (3) Shared phrase: "Peter Pan" listed in BOTH global and multiwords ---
	// Official multiwords: Peter Pan;NPMS000_
	// Official global: Peter Pan → NPCN000 open/close
	// Java order: global tags first, multiword adds NPMS000_, removePreviousTags → NPMS000_
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
		require.True(t, hasExactPOS(gotM[0], "NPMS000_"), "%s multiword-only: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NPMS000_"), "%s multiword-only: %v", label, gotM[1])

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPMS000_"),
			"%s full hybrid prefers multiword NPMS000_ after global+removePreviousTags: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPMS000_"),
			"%s full hybrid Pan: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must flatten angles: %v %v", label, gotFull[0], gotFull[1])
		// Must not leave only NPCN000 from global without multiword tag.
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")
	}

	// --- (4) XML-only effects: pt UNIVERSITY_OF + global proper noun ---
	// Chunkers do not match these surfaces; XML stage sets ignore_spelling.
	{
		// pt rule: case_sensitive University of
		sentUni := tokenSentence("University", "of")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentUni), "University", "of")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentUni), "University", "of")
		requireIgnored(t, onlyXML.Disambiguate(sentUni), "University", "of")
		requireIgnored(t, full.Disambiguate(sentUni), "University", "of")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// pt ROMAN_NUMBER_IGNORE_SPELLING: XIV (XML after chunkers still fires)
		sentXIV := tokenSentence("XIV")
		requireIgnored(t, full.Disambiguate(sentXIV), "XIV")
	}
}

// TestPortugueseHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkPOS proves
// XML last does not wipe multiword/global POS on official surfaces.
func TestPortugueseHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkPOS(t *testing.T) {
	requirePTHybridResources(t)
	full := NewPortugueseHybridDisambiguator()
	// Peter Pan multiword POS survives XML stage.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPMS000_"), "Peter after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMS000_"), "Pan after full order: %v", got[1])
	// Microsoft Entra flattened global POS survives XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")))
	got = contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPCN000"), "Microsoft after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPCN000"), "Entra after full order: %v", got[1])
	// Bin Laden multiword POS survives XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Bin", "Laden")))
	got = contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPMS000"), "Bin after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMS000"), "Laden after full order: %v", got[1])
}
