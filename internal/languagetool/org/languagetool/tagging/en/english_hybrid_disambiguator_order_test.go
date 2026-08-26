package en

// Outcome twins for EnglishHybridDisambiguator full stage order:
// Java EnglishHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
// i.e. spelling_global (tagForNotAddingTags + setIgnoreSpelling; allowFirstCapitalized=true) →
//     /en/multiwords.txt (setIgnoreSpelling + setRemovePreviousTags; allowFirstCapitalized=true) →
//     XmlRuleDisambiguator(lang, true).
// Official english.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/global/XML stage leaves).
//
// Differs from FR/NL hybrids:
//   - GlobalChunker: allowFirstCapitalized=true (FR/NL global=false)
//   - Multiwords: normal open/close tags + setRemovePreviousTags(true) + setIgnoreSpelling(true)
//     (FR multiwords: removePreviousTags only, no ignore; NL multiwords: ignore + tagForNotAddingTags)
// Differs from ES/PT hybrids:
//   - Global uses tagForNotAddingTags (not NPCN000 invent)
// EN multiwords ignore+removePreviousTags is closest to PT multiwords flags.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func requireENHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverEnglishGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	if DiscoverEnglishMultiwords() == "" {
		t.Skip("official en/multiwords.txt not discoverable")
	}
	if DiscoverEnglishDisambiguationXML() == "" {
		t.Skip("official en/disambiguation.xml not discoverable")
	}
	if DiscoverGlobalDisambiguationXML() == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
}

// requireNoGlobalInventPOS asserts English GlobalChunker tagForNotAddingTags:
// no angle POS, no surface _NONE_, no Romance NPCN000 invent.
func requireNoGlobalInventPOS(t *testing.T, out *languagetool.AnalyzedSentence, label string) {
	t.Helper()
	for i, tags := range contentPOSTags(out) {
		require.False(t, hasAnyAnglePOS(tags),
			"%s token[%d] tagForNotAddingTags must not invent angle POS: %v", label, i, tags)
		for _, p := range tags {
			require.NotEqual(t, disambiguation.TagForNotAddingTags, p,
				"%s token[%d] must not surface internal _NONE_ tag: %v", label, i, tags)
			require.False(t, p == "NPCN000" || strings.HasPrefix(p, "<NPCN") || strings.HasPrefix(p, "</NPCN"),
				"%s token[%d] must not invent NPCN000-style chunk POS: %v", label, i, tags)
		}
	}
}

// TestDefaultEnglishHybridDisambiguator_WiresAllThreeStages proves Java constructor
// field wiring (process-cached DefaultEnglishHybridDisambiguator) builds GlobalChunker,
// multiwords Chunker, and XmlRuleDisambiguator when official resources are present —
// with English flags.
func TestDefaultEnglishHybridDisambiguator_WiresAllThreeStages(t *testing.T) {
	requireENHybridResources(t)

	g := EnglishGlobalChunker()
	mw := EnglishMultiWordChunker()
	xml := EnglishHybridXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker,
		"chunkerGlobal = MultiWordChunker.getInstance(/spelling_global.txt, true, true, false, tagForNotAddingTags)")
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/en/multiwords.txt, true, true, false) + setIgnoreSpelling + setRemovePreviousTags")
	require.NotNil(t, d.RulesDisambiguator, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.Same(t, g, d.GlobalChunker)
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.RulesDisambiguator)

	// English settings (no invent):
	// global: tagForNotAddingTags + ignoreSpelling + NO removePreviousTags
	// multiwords: removePreviousTags + ignoreSpelling (both true)
	require.True(t, g.AddIgnoreSpelling, "English chunkerGlobal.setIgnoreSpelling(true)")
	require.False(t, g.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	require.True(t, mw.AddIgnoreSpelling, "English chunker.setIgnoreSpelling(true)")
	require.True(t, mw.RemovePreviousTags, "English chunker.setRemovePreviousTags(true)")
	// allowFirstCapitalized=true on BOTH is proven by outcome cases below (Acid house / Quid pro quo).
}

// TestEnglishHybridDisambiguator_OrderGlobalThenMultiwordThenXML proves stage
// isolation vs full Java order with Java-visible POS / ignore_spelling outcomes.
func TestEnglishHybridDisambiguator_OrderGlobalThenMultiwordThenXML(t *testing.T) {
	requireENHybridResources(t)

	g := EnglishGlobalChunker()
	mw := EnglishMultiWordChunker()
	xml := EnglishHybridXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	// Isolation hybrids (only one stage) vs full order.
	onlyGlobal := &EnglishHybridDisambiguator{GlobalChunker: g}
	onlyMulti := &EnglishHybridDisambiguator{Chunker: mw}
	onlyXML := &EnglishHybridDisambiguator{RulesDisambiguator: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(g.Disambiguate(sent)))
	}
	full := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.RulesDisambiguator)

	// --- (1) Global-only phrase: "Microsoft Entra" in spelling_global, not en/multiwords ---
	// Global alone → ignore_spelling; no invent POS (tagForNotAddingTags)
	// Multiword alone → no ignore, no multiword POS
	// XML alone → no ignore
	// Full hybrid → ignore (global); no invent POS
	// Without global → no ignore
	{
		label := "Microsoft Entra"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra"))
		}

		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, true, label+" global-only ignore")
		requireNoGlobalInventPOS(t, outG, label+" global-only")

		outM := onlyMulti.Disambiguate(fresh())
		requireAllContentIgnored(t, outM, false, label+" multiword-only no ignore")
		gotM := contentPOSTags(outM)
		require.Len(t, gotM, 2, label)
		for i, tags := range gotM {
			require.False(t, hasExactPOS(tags, "NNP") || hasExactPOS(tags, "NN") || hasAnyAnglePOS(tags),
				"%s multiword-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		outX := onlyXML.Disambiguate(fresh())
		requireAllContentIgnored(t, outX, false, label+" xml-only no ignore")
		requireNoGlobalInventPOS(t, outX, label+" xml-only")

		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore")
		requireNoGlobalInventPOS(t, outFull, label+" full hybrid")

		outJO := javaOrder(fresh())
		requireAllContentIgnored(t, outJO, true, label+" javaOrder ignore")
		requireNoGlobalInventPOS(t, outJO, label+" javaOrder")

		// Without GlobalChunker: multiword+XML must not ignore this global-only surface.
		noGlobal := &EnglishHybridDisambiguator{Chunker: mw, RulesDisambiguator: xml}
		requireAllContentIgnored(t, noGlobal.Disambiguate(fresh()), false,
			label+" without global must not ignore")
	}

	// Global-only casing: allowFirstCapitalized=true on GlobalChunker (EN ≠ FR/NL global=false).
	// "acid house" exact + "Acid house" first-cap match; "Acid House" titlecase denied.
	// None of these are in en/multiwords.
	{
		label := "acid house"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("acid", "house"))
		}
		freshFirstCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Acid", "house"))
		}
		freshTitle := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Acid", "House"))
		}

		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global exact")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), false, label+" multi no match")
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid")
		requireNoGlobalInventPOS(t, full.Disambiguate(fresh()), label+" full")

		// first-cap allowed by GlobalChunker (allowFirstCapitalized=true)
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshFirstCap()), true, "Acid house first-cap allowed global")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(freshFirstCap()), false, "Acid house not in multiwords")
		requireAllContentIgnored(t, full.Disambiguate(freshFirstCap()), true, "Acid house full first-cap")

		// allowTitlecase=false: full titlecase of lower official entry denied
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshTitle()), false, "Acid House titlecase denied global")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(freshTitle()), false, "Acid House not in multiwords")
		requireAllContentIgnored(t, full.Disambiguate(freshTitle()), false, "Acid House full no match")
	}

	// --- (2) Multiword-only phrase: "quid pro quo" / "status quo" / "Yom Kippur"
	// in en/multiwords, not spelling_global ---
	// Official: quid pro quo\tNN → after removePreviousTags: NN NN NN + ignore
	{
		label := "quid pro quo"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("quid", "pro", "quo"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "NN") || hasExactPOS(tags, "NNP") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no ignore")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 3, label)
		require.True(t, hasExactPOS(gotM[0], "NN"), "%s multiword-only quid: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NN"), "%s multiword-only pro: %v", label, gotM[1])
		require.True(t, hasExactPOS(gotM[2], "NN"), "%s multiword-only quo: %v", label, gotM[2])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]) || hasAnyAnglePOS(gotM[2]),
			"%s multiword removePreviousTags flattens angles: %v %v %v", label, gotM[0], gotM[1], gotM[2])
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), true,
			label+" multiword-only ignore (English multiwords setIgnoreSpelling)")

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 3, label)
		require.True(t, hasExactPOS(gotFull[0], "NN"), "%s full hybrid quid: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NN"), "%s full hybrid pro: %v", label, gotFull[1])
		require.True(t, hasExactPOS(gotFull[2], "NN"), "%s full hybrid quo: %v", label, gotFull[2])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]) || hasAnyAnglePOS(gotFull[2]))
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid ignore")

		// Without multiword stage: global+XML must not invent multiword POS or ignore.
		noMulti := &EnglishHybridDisambiguator{GlobalChunker: g, RulesDisambiguator: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "NN") || hasExactPOS(tags, "NNP") || hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}
		requireAllContentIgnored(t, noMulti.Disambiguate(fresh()), false,
			label+" without multiword must not ignore")

		// allowFirstCapitalized=true on multiwords: Quid pro quo matches multiword stage
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Quid", "pro", "quo"))
		}
		gotCap := contentPOSTags(full.Disambiguate(freshCap()))
		require.True(t, hasExactPOS(gotCap[0], "NN"), "Quid pro quo full hybrid first-cap: %v", gotCap[0])
		require.True(t, hasExactPOS(gotCap[1], "NN"), "Quid pro quo full hybrid first-cap: %v", gotCap[1])
		require.True(t, hasExactPOS(gotCap[2], "NN"), "Quid pro quo full hybrid first-cap: %v", gotCap[2])
		requireAllContentIgnored(t, full.Disambiguate(freshCap()), true, "Quid pro quo first-cap ignore")
	}

	// Multiword-only "status quo" (last exact-key write is NN) and "Yom Kippur" (NNP)
	{
		label := "status quo"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("status", "quo"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global no")
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NN"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NN"), "%s multi: %v", label, gotM[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NN"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NN"), "%s full: %v", label, gotFull[1])
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full ignore")
	}
	{
		label := "Yom Kippur"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Yom", "Kippur"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global no")
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NNP"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NNP"), "%s multi: %v", label, gotM[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "NNP"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NNP"), "%s full: %v", label, gotFull[1])
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full ignore")
	}

	// --- (3) Shared phrase: "Google Maps" listed in BOTH global and multiwords ---
	// Official multiwords: Google Maps\tNNP → after removePreviousTags: NNP NNP
	// Official global: tagForNotAddingTags + ignore
	// Java order: global ignores first; multiword adds POS + ignore; no invent from global
	{
		label := "Google Maps"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps"))
		}

		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, true, label+" global-only")
		requireNoGlobalInventPOS(t, outG, label+" global-only")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NNP"), "%s multiword-only: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NNP"), "%s multiword-only: %v", label, gotM[1])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]))
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), true,
			label+" multiword-only ignore")

		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		require.True(t, hasExactPOS(gotFull[0], "NNP"),
			"%s full hybrid multiword POS after global ignore: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NNP"),
			"%s full hybrid Maps: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must flatten multiword angles: %v %v", label, gotFull[0], gotFull[1])
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore")

		// Without multiword: ignore still (global), but no NNP POS.
		noMulti := &EnglishHybridDisambiguator{GlobalChunker: g, RulesDisambiguator: xml}
		outNoM := noMulti.Disambiguate(fresh())
		requireAllContentIgnored(t, outNoM, true, label+" without multi ignore")
		for i, tags := range contentPOSTags(outNoM) {
			require.False(t, hasExactPOS(tags, "NNP") || hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must not have multiword POS: %v", label, i, tags)
		}

		// Without global: multiword POS + ignore still (multiwords setIgnoreSpelling).
		noGlobal := &EnglishHybridDisambiguator{Chunker: mw, RulesDisambiguator: xml}
		outNoG := noGlobal.Disambiguate(fresh())
		gotNoG := contentPOSTags(outNoG)
		require.True(t, hasExactPOS(gotNoG[0], "NNP"), label+" without global still multi POS")
		requireAllContentIgnored(t, outNoG, true, label+" without global still multi ignore")

		// allowAllUppercase=true on both chunkers.
		freshUpper := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("GOOGLE", "MAPS"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshUpper()), true, "GOOGLE MAPS global")
		gotUpperM := contentPOSTags(onlyMulti.Disambiguate(freshUpper()))
		require.True(t, hasExactPOS(gotUpperM[0], "NNP"), "GOOGLE MAPS multi: %v", gotUpperM[0])
		require.True(t, hasExactPOS(gotUpperM[1], "NNP"), "GOOGLE MAPS multi token1: %v", gotUpperM[1])
		outUpperFull := full.Disambiguate(freshUpper())
		requireAllContentIgnored(t, outUpperFull, true, "GOOGLE MAPS full ignore")
		gotUpperFull := contentPOSTags(outUpperFull)
		require.True(t, hasExactPOS(gotUpperFull[0], "NNP"), "GOOGLE MAPS full POS: %v", gotUpperFull[0])
		require.True(t, hasExactPOS(gotUpperFull[1], "NNP"), "GOOGLE MAPS full POS1: %v", gotUpperFull[1])
	}

	// Shared "New York Post" (NNP on multiwords; global ignore) and "picture alliance"
	{
		label := "New York Post"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("New", "York", "Post"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global")
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NNP"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NNP"), "%s multi: %v", label, gotM[1])
		require.True(t, hasExactPOS(gotM[2], "NNP"), "%s multi: %v", label, gotM[2])
		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		require.True(t, hasExactPOS(gotFull[0], "NNP"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NNP"), "%s full: %v", label, gotFull[1])
		require.True(t, hasExactPOS(gotFull[2], "NNP"), "%s full: %v", label, gotFull[2])
		requireAllContentIgnored(t, outFull, true, label+" full ignore")
	}
	{
		label := "picture alliance"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("picture", "alliance"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global")
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "NNP"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "NNP"), "%s multi: %v", label, gotM[1])
		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		require.True(t, hasExactPOS(gotFull[0], "NNP"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NNP"), "%s full: %v", label, gotFull[1])
		requireAllContentIgnored(t, outFull, true, label+" full ignore from global+multi")
	}

	// --- (4) XML-only effects: en KUNG_FU + global proper noun ---
	// Chunkers do not match these surfaces; XML stage sets ignore_spelling.
	// Proves XML runs after chunkers in full hybrid (and isolation).
	{
		// en rule: kung fu → ignore_spelling
		sentKung := tokenSentence("kung", "fu")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentKung), "kung", "fu")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentKung), "kung", "fu")
		requireIgnored(t, onlyXML.Disambiguate(sentKung), "kung", "fu")
		requireIgnored(t, full.Disambiguate(sentKung), "kung", "fu")
		requireIgnored(t, javaOrder(sentKung), "kung", "fu")

		// Without XML: chunkers alone must not ignore kung fu
		noXML := &EnglishHybridDisambiguator{GlobalChunker: g, Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("kung", "fu")), "kung", "fu")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// en SPELLING_IN_VIVO / SPELLING_KETO (XML after chunkers still fires)
		requireIgnored(t, full.Disambiguate(tokenSentence("in", "vivo")), "in", "vivo")
		requireIgnored(t, full.Disambiguate(tokenSentence("keto", "diet")), "keto", "diet")

		// UNKNOWN_PCT via full hybrid RulesDisambiguator (POS, not ignore)
		sentDot := full.Disambiguate(tokenSentence("."))
		tr := tokenBySurface(sentDot, ".")
		require.NotNil(t, tr)
		require.Contains(t, posTagsOn(tr), "PCT")
	}
}

// TestEnglishHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockOutcomes proves
// XML last does not wipe ignore_spelling / multiword POS set by earlier stages.
func TestEnglishHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockOutcomes(t *testing.T) {
	requireENHybridResources(t)
	full := DefaultEnglishHybridDisambiguator()

	// Global-only ignore survives XML; no invent POS.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")))
	requireAllContentIgnored(t, out, true, "Microsoft Entra after full order")
	requireNoGlobalInventPOS(t, out, "Microsoft Entra after full")

	// Multiword-only POS + ignore survives XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("quid", "pro", "quo")))
	got := contentPOSTags(out)
	require.Len(t, got, 3)
	require.True(t, hasExactPOS(got[0], "NN"), "quid after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NN"), "pro after full order: %v", got[1])
	require.True(t, hasExactPOS(got[2], "NN"), "quo after full order: %v", got[2])
	requireAllContentIgnored(t, out, true, "quid pro quo multiword ignore after full")

	// Shared: ignore from global+multi + multiword POS survive XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "NNP"), "Google after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NNP"), "Maps after full order: %v", got[1])
	requireAllContentIgnored(t, out, true, "Google Maps ignore after full")

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("New", "York", "Post")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "NNP"), "New after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "NNP"), "York after full order: %v", got[1])
	require.True(t, hasExactPOS(got[2], "NNP"), "Post after full order: %v", got[2])
	requireAllContentIgnored(t, out, true, "New York Post ignore after full")

	// XML effects still fire after chunkers.
	requireIgnored(t, full.Disambiguate(tokenSentence("kung", "fu")), "kung", "fu")
	requireIgnored(t, full.Disambiguate(tokenSentence("QB|LT")), "QB|LT")
}

// TestEnglishHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == xml(mw(g(input))) for official isolation surfaces.
func TestEnglishHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireENHybridResources(t)
	g := EnglishGlobalChunker()
	mw := EnglishMultiWordChunker()
	xml := EnglishHybridXmlRuleDisambiguator()
	full := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		parts    []string
		label    string
		wantIg   bool
		wantPOS0 string // empty = no multiword POS required
		wantPOS1 string
		wantPOS2 string // optional third token
	}
	cases := []caseT{
		{[]string{"Microsoft", "Entra"}, "Microsoft Entra", true, "", "", ""},
		{[]string{"acid", "house"}, "acid house", true, "", "", ""},
		{[]string{"Acid", "house"}, "Acid house first-cap", true, "", "", ""},
		{[]string{"Acid", "House"}, "Acid House titlecase denied", false, "", "", ""},
		{[]string{"quid", "pro", "quo"}, "quid pro quo", true, "NN", "NN", "NN"},
		{[]string{"status", "quo"}, "status quo", true, "NN", "NN", ""},
		{[]string{"Yom", "Kippur"}, "Yom Kippur", true, "NNP", "NNP", ""},
		{[]string{"Google", "Maps"}, "Google Maps", true, "NNP", "NNP", ""},
		{[]string{"New", "York", "Post"}, "New York Post", true, "NNP", "NNP", "NNP"},
		{[]string{"picture", "alliance"}, "picture alliance", true, "NNP", "NNP", ""},
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed", false, "", "", ""},
	}
	for _, tc := range cases {
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...))
		}
		outFull := full.Disambiguate(fresh())
		outManual := xml.Disambiguate(mw.Disambiguate(g.Disambiguate(fresh())))
		requireAllContentIgnored(t, outFull, tc.wantIg, tc.label+" full")
		requireAllContentIgnored(t, outManual, tc.wantIg, tc.label+" javaOrder manual")

		// Per-token ignore parity between full hybrid and explicit composition.
		ft, mt := outFull.GetTokens(), outManual.GetTokens()
		require.Equal(t, len(ft), len(mt), tc.label+" token count")
		for i := range ft {
			if i == 0 || ft[i].IsWhitespace() {
				continue
			}
			require.Equal(t, ft[i].IsIgnoredBySpeller(), mt[i].IsIgnoredBySpeller(),
				"%s token[%d]=%q ignore parity full vs javaOrder", tc.label, i, ft[i].GetToken())
		}

		gotFull := contentPOSTags(outFull)
		gotManual := contentPOSTags(outManual)
		require.Equal(t, len(gotFull), len(gotManual), tc.label+" content POS count")
		if tc.wantPOS0 != "" {
			require.True(t, hasExactPOS(gotFull[0], tc.wantPOS0), "%s full POS0: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotManual[0], tc.wantPOS0), "%s manual POS0: %v", tc.label, gotManual[0])
			require.True(t, hasExactPOS(gotFull[1], tc.wantPOS1), "%s full POS1: %v", tc.label, gotFull[1])
			require.True(t, hasExactPOS(gotManual[1], tc.wantPOS1), "%s manual POS1: %v", tc.label, gotManual[1])
			if tc.wantPOS2 != "" {
				require.True(t, hasExactPOS(gotFull[2], tc.wantPOS2), "%s full POS2: %v", tc.label, gotFull[2])
				require.True(t, hasExactPOS(gotManual[2], tc.wantPOS2), "%s manual POS2: %v", tc.label, gotManual[2])
			}
		} else {
			// Global-only / non-listed: no multiword invent; tagForNotAddingTags on global.
			requireNoGlobalInventPOS(t, outFull, tc.label+" full")
			requireNoGlobalInventPOS(t, outManual, tc.label+" manual")
		}
		// Per-token POS set parity.
		for i := range gotFull {
			require.ElementsMatch(t, gotFull[i], gotManual[i],
				"%s token[%d] POS parity full vs javaOrder", tc.label, i)
		}
	}
}
