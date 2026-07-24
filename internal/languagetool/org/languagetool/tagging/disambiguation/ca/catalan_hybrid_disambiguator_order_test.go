package ca

// Outcome twins for CatalanHybridDisambiguator full stage order:
// Java CatalanHybridDisambiguator.disambiguate:
//   multitokenDisambiguator.disambiguate(
//     disambiguator.disambiguate(
//       chunker.disambiguate(
//         chunkerGlobal.disambiguate(input))))
// i.e. spelling_global (NPCN000; NO ignoreSpelling; NO removePreviousTags) →
//   /ca/multiwords.txt (setRemovePreviousTags; NO setIgnoreSpelling) →
//   XmlRuleDisambiguator(lang, true) →
//   CatalanMultitokenDisambiguator (no-op when IsMisspelled/speller nil).
// Official catalan.dict / multitoken speller not required: token-built
// AnalyzedSentence patterns (same helpers as ACCEPTed multiword/global/XML leaves).
//
// Differs from FR (tagForNotAddingTags + ignore on global) and PT (ignore on both):
//   - GlobalChunker: DefaultTag NPCN000 open/close; no ignore; no removePreviousTags
//   - Multiwords: removePreviousTags + NO ignoreSpelling
//   - Fourth stage Multitoken (like DE rule disambiguator multitoken tail)
//   - allowFirstCapitalized: global=false, multiwords=true

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireCAHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverCatalanGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	if DiscoverCatalanMultiwords() == "" {
		t.Skip("official ca/multiwords.txt not discoverable")
	}
	if DiscoverCatalanDisambiguationXML() == "" {
		t.Skip("official ca/disambiguation.xml not discoverable")
	}
	if DiscoverGlobalDisambiguationXML() == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
}

// TestNewCatalanHybridDisambiguator_WiresAllFourStages proves Java constructor
// eagerly builds GlobalChunker, multiwords Chunker, XmlRuleDisambiguator, and
// CatalanMultitokenDisambiguator when the same official resources Java loads are present.
func TestNewCatalanHybridDisambiguator_WiresAllFourStages(t *testing.T) {
	requireCAHybridResources(t)

	g := CatalanGlobalChunker()
	mw := CatalanMultiWordChunker()
	xml := CatalanXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewCatalanHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker,
		"chunkerGlobal = MultiWordChunker.getInstance(/spelling_global.txt, false, true, false, NPCN000)")
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/ca/multiwords.txt, true, true, false) + setRemovePreviousTags(true)")
	require.NotNil(t, d.Rules, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.NotNil(t, d.Multitoken, "multitokenDisambiguator = new CatalanMultitokenDisambiguator()")
	require.Same(t, g, d.GlobalChunker)
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Multitoken is always a CatalanMultitokenDisambiguator with nil IsMisspelled
	// (Java no-ops when speller == null; do not invent a fake dictionary).
	mt, ok := d.Multitoken.(*CatalanMultitokenDisambiguator)
	require.True(t, ok, "Multitoken field must be *CatalanMultitokenDisambiguator")
	require.Nil(t, mt.IsMisspelled, "IsMisspelled nil → identity like Java speller==null")

	// Catalan settings (no invent):
	// global: NPCN000, NO ignoreSpelling, NO removePreviousTags
	// multiwords: removePreviousTags, NO ignoreSpelling
	require.False(t, g.AddIgnoreSpelling, "Catalan chunkerGlobal does NOT setIgnoreSpelling")
	require.False(t, g.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	// DefaultTag NPCN000 is proven by open/close <NPCN000> outcomes in order isolation tests (not tagForNotAddingTags/_NONE_).
	require.True(t, mw.RemovePreviousTags, "Catalan chunker.setRemovePreviousTags(true)")
	require.False(t, mw.AddIgnoreSpelling, "Catalan multiwords does NOT setIgnoreSpelling")
}

// TestCatalanHybridDisambiguator_OrderGlobalThenMultiwordThenXMLThenMultitoken
// proves stage isolation vs full Java order with Java-visible POS / ignore outcomes.
func TestCatalanHybridDisambiguator_OrderGlobalThenMultiwordThenXMLThenMultitoken(t *testing.T) {
	requireCAHybridResources(t)

	g := CatalanGlobalChunker()
	mw := CatalanMultiWordChunker()
	xml := CatalanXmlRuleDisambiguator()
	mt := NewCatalanMultitokenDisambiguator() // IsMisspelled nil → identity
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyGlobal := &CatalanHybridDisambiguator{GlobalChunker: g}
	onlyMulti := &CatalanHybridDisambiguator{Chunker: mw}
	onlyXML := &CatalanHybridDisambiguator{Rules: xml}
	onlyMT := &CatalanHybridDisambiguator{Multitoken: mt}
	// Reverse chunk stages for order proof (multiword then global — opposite of Java).
	reverseChunks := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return g.Disambiguate(mw.Disambiguate(sent))
	}
	// Manual Java order: Multitoken(XML(mw(global(input))))
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return mt.Disambiguate(xml.Disambiguate(mw.Disambiguate(g.Disambiguate(sent))))
	}
	full := NewCatalanHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)
	require.NotNil(t, full.Multitoken)

	// --- (1) Global-only phrase: "Google Maps" in spelling_global, not ca/multiwords ---
	// Global alone → open/close <NPCN000></NPCN000>
	// Multiword alone → no chunk POS
	// Full hybrid (global→multiword removePreviousTags) → plain NPCN000 NPCN000
	// Wrong chunk order (multiword→global) → angle tags remain
	// Multitoken alone (no speller) → identity
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

		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no chunk POS, got %v", label, i, tags)
		}

		// Multitoken alone with nil IsMisspelled: identity (no invent POS)
		gotMT := contentPOSTags(onlyMT.Disambiguate(fresh()))
		for i, tags := range gotMT {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasExactPOS(tags, "NPCNM00") || hasAnyAnglePOS(tags),
				"%s multitoken-only token[%d] must not invent POS, got %v", label, i, tags)
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

		// javaOrder composition matches full
		gotJO := contentPOSTags(javaOrder(fresh()))
		require.True(t, hasExactPOS(gotJO[0], "NPCN000"), "%s javaOrder: %v", label, gotJO[0])
		require.True(t, hasExactPOS(gotJO[1], "NPCN000"), "%s javaOrder: %v", label, gotJO[1])

		// Never ignore spelling from CA GlobalChunker / multiwords (no setIgnoreSpelling)
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid token %q must not ignore spelling via chunkers", label, tr.GetToken())
		}

		// Without multiword: angles remain (no removePreviousTags after global)
		noMulti := &CatalanHybridDisambiguator{GlobalChunker: g, Rules: xml, Multitoken: mt}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotNoM[0], "<NPCN000>"),
			"%s without multiword leaves open angle: %v", label, gotNoM[0])
		require.True(t, hasExactPOS(gotNoM[1], "</NPCN000>"),
			"%s without multiword leaves close angle: %v", label, gotNoM[1])
	}

	// Global-only "Microsoft Entra" (same pattern; not in multiwords)
	{
		label := "Microsoft Entra"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra"))
		}
		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotG[0], "<NPCN000>"), "%s global: %v", label, gotG[0])
		require.True(t, hasExactPOS(gotG[1], "</NPCN000>"), "%s global: %v", label, gotG[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPCN000"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPCN000"), "%s full: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]))

		// allowFirstCapitalized=false: first-cap of lowercase official entry denied
		// (picture alliance exact matches; Picture alliance does not)
	}

	// Global casing: "picture alliance" matches; "Picture alliance" does not
	{
		label := "picture alliance"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("picture", "alliance"))
		}
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Picture", "alliance"))
		}
		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotG[0], "<NPCN000>"), "%s global exact: %v", label, gotG[0])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPCN000"), "%s full: %v", label, gotFull[0])

		gotCapG := contentPOSTags(onlyGlobal.Disambiguate(freshCap()))
		for i, tags := range gotCapG {
			require.False(t, hasExactPOS(tags, "<NPCN000>") || hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
				"Picture alliance first-cap denied global token[%d]: %v", i, tags)
		}
		gotCapFull := contentPOSTags(full.Disambiguate(freshCap()))
		for i, tags := range gotCapFull {
			require.False(t, hasExactPOS(tags, "NPCN000") || hasAnyAnglePOS(tags),
				"Picture alliance full no match token[%d]: %v", i, tags)
		}
	}

	// --- (2) Multiword-only phrase: "uilleann pipes" in ca/multiwords, not spelling_global ---
	// Official line: uilleann pipes;NCFN000 → after removePreviousTags: NCFN000 AQ0FN0
	// NO setIgnoreSpelling on Catalan multiwords → content tokens not ignored.
	{
		label := "uilleann pipes"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("uilleann", "pipes"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "NCFN000") || hasExactPOS(tags, "AQ0FN0") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "NCFN000"), "%s multiword-only uilleann: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "AQ0FN0"), "%s multiword-only pipes: %v", label, gotM[1])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]),
			"%s multiword removePreviousTags flattens angles: %v %v", label, gotM[0], gotM[1])

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "NCFN000"), "%s full hybrid uilleann: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "AQ0FN0"), "%s full hybrid pipes: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]))

		// Without multiword stage: global+XML+MT must not invent multiword POS.
		noMulti := &CatalanHybridDisambiguator{GlobalChunker: g, Rules: xml, Multitoken: mt}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "NCFN000") || hasExactPOS(tags, "AQ0FN0") || hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		// allowFirstCapitalized=true on multiwords: Uilleann pipes matches
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Uilleann", "pipes"))
		}
		gotCap := contentPOSTags(full.Disambiguate(freshCap()))
		require.True(t, hasExactPOS(gotCap[0], "NCFN000"), "Uilleann pipes full hybrid first-cap: %v", gotCap[0])
		require.True(t, hasExactPOS(gotCap[1], "AQ0FN0"), "Uilleann pipes full hybrid first-cap: %v", gotCap[1])

		// Never ignore from multiwords
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid must not ignore via multiwords", label)
		}
	}

	// Multiword-only "Agnes Callard" (NPFSSP0) and "time lapse" (NCMS000 → NCMS000 AQ0MS0)
	{
		label := "Agnes Callard"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Agnes", "Callard"))
		}
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NPFSSP0"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NPFSSP0"), "%s multi: %v", label, gotM[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NPFSSP0"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPFSSP0"), "%s full: %v", label, gotFull[1])
	}
	{
		label := "time lapse"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("time", "lapse"))
		}
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NCMS000"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "AQ0MS0"), "%s multi: %v", label, gotM[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NCMS000"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "AQ0MS0"), "%s full: %v", label, gotFull[1])
	}

	// --- (3) Shared phrase: "Peter Pan" listed in BOTH global and multiwords ---
	// Official multiwords: Peter Pan;NPMNSP0
	// Official global: Peter Pan → NPCN000 open/close
	// Java order: global tags first, multiword adds NPMNSP0, removePreviousTags → NPMNSP0
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

		// Without multiword: ignore still? No ignore on CA — but angles from global.
		noMulti := &CatalanHybridDisambiguator{GlobalChunker: g, Rules: xml, Multitoken: mt}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotNoM[0], "<NPCN000>") || hasExactPOS(gotNoM[0], "NPCN000"),
			"%s without multiword still has global POS: %v", label, gotNoM[0])
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "NPMNSP0"),
				"%s without multiword token[%d] must not have multiword POS: %v", label, i, tags)
		}

		// Without global: multiword POS still
		noGlobal := &CatalanHybridDisambiguator{Chunker: mw, Rules: xml, Multitoken: mt}
		gotNoG := contentPOSTags(noGlobal.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotNoG[0], "NPMNSP0"), label+" without global still multi POS")
	}

	// --- (4) XML-only effects: ca HAHAHA + global proper noun ---
	// Chunkers do not set ignore_spelling; XML stage does.
	// Multitoken with nil IsMisspelled does not clear ignore.
	{
		// ca rule: ha(ha)+ → hahaha
		sentHA := tokenSentence("hahaha")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentHA), "hahaha")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentHA), "hahaha")
		requireNotIgnored(t, onlyMT.Disambiguate(sentHA), "hahaha")
		requireIgnored(t, onlyXML.Disambiguate(sentHA), "hahaha")
		requireIgnored(t, full.Disambiguate(sentHA), "hahaha")
		requireIgnored(t, javaOrder(sentHA), "hahaha")

		// Without XML: chunkers + multitoken must not ignore hahaha
		noXML := &CatalanHybridDisambiguator{GlobalChunker: g, Chunker: mw, Multitoken: mt}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("hahaha")), "hahaha")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// ca mes_info: més info → ignore on info only (XML after chunkers still fires)
		sentInfo := tokenSentence("més", "info")
		outInfo := full.Disambiguate(sentInfo)
		requireNotIgnored(t, outInfo, "més")
		requireIgnored(t, outInfo, "info")
	}
}

// TestCatalanHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockOutcomes proves
// XML last (before Multitoken) does not wipe multiword/global POS; Multitoken no-op
// does not invent tags on already-tagged content.
func TestCatalanHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockOutcomes(t *testing.T) {
	requireCAHybridResources(t)
	full := NewCatalanHybridDisambiguator()

	// Global-only flattened POS survives XML + Multitoken.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "NPCN000"), "Google after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPCN000"), "Maps after full order: %v", got[1])
	require.False(t, hasAnyAnglePOS(got[0]) || hasAnyAnglePOS(got[1]))

	// Multiword-only POS survives XML + Multitoken.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("uilleann", "pipes")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "NCFN000"), "uilleann after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "AQ0FN0"), "pipes after full order: %v", got[1])

	// Shared: multiword POS survives.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Peter", "Pan")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "NPMNSP0"), "Peter after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NPMNSP0"), "Pan after full order: %v", got[1])

	// XML effects still fire after chunkers (before Multitoken no-op).
	requireIgnored(t, full.Disambiguate(tokenSentence("hahaha")), "hahaha")
	requireIgnored(t, full.Disambiguate(tokenSentence("QB|LT")), "QB|LT")
}

// TestCatalanHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == Multitoken(XML(mw(g(input)))) for official isolation surfaces.
// With Multitoken IsMisspelled nil (identity), also equals XML(mw(g(input))).
func TestCatalanHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireCAHybridResources(t)
	g := CatalanGlobalChunker()
	mw := CatalanMultiWordChunker()
	xml := CatalanXmlRuleDisambiguator()
	mt := NewCatalanMultitokenDisambiguator() // nil IsMisspelled → identity
	full := NewCatalanHybridDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)
	require.NotNil(t, full.Multitoken)

	type caseT struct {
		parts    []string
		label    string
		wantPOS0 string // empty = no multiword/global invent required beyond no-angle check
		wantPOS1 string
		// wantGlobalFlatten: expect plain NPCN000 on both (global-only after removePreviousTags)
		wantGlobalFlatten bool
	}
	cases := []caseT{
		{[]string{"Google", "Maps"}, "Google Maps", "NPCN000", "NPCN000", true},
		{[]string{"Microsoft", "Entra"}, "Microsoft Entra", "NPCN000", "NPCN000", true},
		{[]string{"picture", "alliance"}, "picture alliance", "NPCN000", "NPCN000", true},
		{[]string{"Picture", "alliance"}, "Picture alliance", "", "", false}, // first-cap denied
		{[]string{"uilleann", "pipes"}, "uilleann pipes", "NCFN000", "AQ0FN0", false},
		{[]string{"Agnes", "Callard"}, "Agnes Callard", "NPFSSP0", "NPFSSP0", false},
		{[]string{"time", "lapse"}, "time lapse", "NCMS000", "AQ0MS0", false},
		{[]string{"Peter", "Pan"}, "Peter Pan", "NPMNSP0", "NPMNSP0", false},
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed", "", "", false},
	}
	for _, tc := range cases {
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...))
		}
		outFull := full.Disambiguate(fresh())
		// Full four-stage Java order
		outManual4 := mt.Disambiguate(xml.Disambiguate(mw.Disambiguate(g.Disambiguate(fresh()))))
		// First three stages only (Multitoken identity when IsMisspelled nil)
		outManual3 := xml.Disambiguate(mw.Disambiguate(g.Disambiguate(fresh())))

		// Per-token ignore parity between full hybrid and explicit composition.
		ft, m4, m3 := outFull.GetTokens(), outManual4.GetTokens(), outManual3.GetTokens()
		require.Equal(t, len(ft), len(m4), tc.label+" token count full vs 4-stage")
		require.Equal(t, len(ft), len(m3), tc.label+" token count full vs 3-stage")
		for i := range ft {
			if i == 0 || ft[i].IsWhitespace() {
				continue
			}
			require.Equal(t, ft[i].IsIgnoredBySpeller(), m4[i].IsIgnoredBySpeller(),
				"%s token[%d]=%q ignore parity full vs javaOrder4", tc.label, i, ft[i].GetToken())
			require.Equal(t, ft[i].IsIgnoredBySpeller(), m3[i].IsIgnoredBySpeller(),
				"%s token[%d]=%q ignore parity full vs javaOrder3 (MT identity)", tc.label, i, ft[i].GetToken())
		}

		gotFull := contentPOSTags(outFull)
		gotM4 := contentPOSTags(outManual4)
		gotM3 := contentPOSTags(outManual3)
		require.Equal(t, len(gotFull), len(gotM4), tc.label+" content POS count")
		require.Equal(t, len(gotFull), len(gotM3), tc.label+" content POS count 3")

		if tc.wantPOS0 != "" {
			require.True(t, hasExactPOS(gotFull[0], tc.wantPOS0), "%s full POS0: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotM4[0], tc.wantPOS0), "%s manual4 POS0: %v", tc.label, gotM4[0])
			require.True(t, hasExactPOS(gotM3[0], tc.wantPOS0), "%s manual3 POS0: %v", tc.label, gotM3[0])
			require.True(t, hasExactPOS(gotFull[1], tc.wantPOS1), "%s full POS1: %v", tc.label, gotFull[1])
			require.True(t, hasExactPOS(gotM4[1], tc.wantPOS1), "%s manual4 POS1: %v", tc.label, gotM4[1])
			require.True(t, hasExactPOS(gotM3[1], tc.wantPOS1), "%s manual3 POS1: %v", tc.label, gotM3[1])
			if tc.wantGlobalFlatten {
				require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
					"%s must flatten angles: %v %v", tc.label, gotFull[0], gotFull[1])
			}
		} else {
			// Non-listed / denied: no multiword invent, no angle invent from multiwords stage.
			for i, tags := range gotFull {
				require.False(t, hasExactPOS(tags, "NPCN000") || hasExactPOS(tags, "<NPCN000>") ||
					hasExactPOS(tags, "</NPCN000>") || hasAnyAnglePOS(tags) || hasExactPOS(tags, "NPCNM00"),
					"%s token[%d] must not invent chunk/multitoken POS: %v", tc.label, i, tags)
			}
		}
		// Per-token POS set parity (membership).
		for i := range gotFull {
			require.ElementsMatch(t, gotFull[i], gotM4[i],
				"%s token[%d] POS parity full vs javaOrder4", tc.label, i)
			require.ElementsMatch(t, gotFull[i], gotM3[i],
				"%s token[%d] POS parity full vs javaOrder3 (MT identity)", tc.label, i)
		}
	}
}

// TestCatalanHybridDisambiguator_MultitokenNoOpWithoutSpeller proves that with
// IsMisspelled nil (Java speller==null), Multitoken is identity and does not invent
// NPCNM00 on untagged multi-token surfaces.
func TestCatalanHybridDisambiguator_MultitokenNoOpWithoutSpeller(t *testing.T) {
	requireCAHybridResources(t)
	full := NewCatalanHybridDisambiguator()
	mt, ok := full.Multitoken.(*CatalanMultitokenDisambiguator)
	require.True(t, ok)
	require.Nil(t, mt.IsMisspelled)

	// Untagged non-listed titlecase phrase: Multitoken must not invent NPCNM00 without speller.
	fresh := languagetool.NewAnalyzedSentence(multiwordTokens("Zxqwv", "Plmnb"))
	out := full.Disambiguate(fresh)
	for i, tags := range contentPOSTags(out) {
		require.False(t, hasExactPOS(tags, "NPCNM00"),
			"token[%d] Multitoken must not invent NPCNM00 without speller: %v", i, tags)
	}

	// leave-one-out: hybrid without Multitoken equals full when Multitoken is identity
	g := CatalanGlobalChunker()
	mw := CatalanMultiWordChunker()
	xml := CatalanXmlRuleDisambiguator()
	noMT := &CatalanHybridDisambiguator{GlobalChunker: g, Chunker: mw, Rules: xml}
	for _, parts := range [][]string{
		{"Google", "Maps"},
		{"uilleann", "pipes"},
		{"Peter", "Pan"},
		{"hahaha"},
	} {
		var freshSent *languagetool.AnalyzedSentence
		if len(parts) == 1 {
			freshSent = tokenSentence(parts[0])
		} else {
			freshSent = languagetool.NewAnalyzedSentence(multiwordTokens(parts...))
		}
		outFull := full.Disambiguate(freshSent)
		// rebuild for noMT
		if len(parts) == 1 {
			freshSent = tokenSentence(parts[0])
		} else {
			freshSent = languagetool.NewAnalyzedSentence(multiwordTokens(parts...))
		}
		outNoMT := noMT.Disambiguate(freshSent)
		require.ElementsMatch(t,
			flattenPOS(contentPOSTags(outFull)),
			flattenPOS(contentPOSTags(outNoMT)),
			"without Multitoken POS parity for %v", parts)
	}
}

func flattenPOS(all [][]string) []string {
	var out []string
	for _, tags := range all {
		out = append(out, tags...)
	}
	return out
}
