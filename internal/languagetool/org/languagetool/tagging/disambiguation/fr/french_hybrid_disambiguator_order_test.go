package fr

// Outcome twins for FrenchHybridDisambiguator full stage order:
// Java FrenchHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
// i.e. spelling_global (tagForNotAddingTags + setIgnoreSpelling) →
//     /fr/multiwords.txt (setRemovePreviousTags; NO setIgnoreSpelling) →
//     XmlRuleDisambiguator(lang, true).
// Official french.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/global/XML stage leaves).
//
// Differs from ES/PT/NL hybrids:
//   - GlobalChunker: tagForNotAddingTags + setIgnoreSpelling(true) (like NL global; unlike ES/PT NPCN000)
//   - Multiwords: normal open/close tags + setRemovePreviousTags(true); NO setIgnoreSpelling
//     (like ES multiwords; unlike NL which ignores + tagForNotAddingTags on multiwords)
//   - allowFirstCapitalized: global=false, multiwords=true

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func requireFRHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverFrenchGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	if DiscoverFrenchMultiwords() == "" {
		t.Skip("official fr/multiwords.txt not discoverable")
	}
	if DiscoverFrenchDisambiguationXML() == "" {
		t.Skip("official fr/disambiguation.xml not discoverable")
	}
	if DiscoverGlobalDisambiguationXML() == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
}

// requireNoGlobalInventPOS asserts French GlobalChunker tagForNotAddingTags:
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

// TestNewFrenchHybridDisambiguator_WiresAllThreeStages proves Java constructor
// eagerly builds GlobalChunker, multiwords Chunker, and XmlRuleDisambiguator when
// the same official resources Java loads are present — with French flags.
func TestNewFrenchHybridDisambiguator_WiresAllThreeStages(t *testing.T) {
	requireFRHybridResources(t)

	g := FrenchGlobalChunker()
	mw := FrenchMultiWordChunker()
	xml := FrenchXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewFrenchHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker,
		"chunkerGlobal = MultiWordChunker.getInstance(/spelling_global.txt, false, true, false, tagForNotAddingTags)")
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/fr/multiwords.txt, true, true, false) + setRemovePreviousTags(true)")
	require.NotNil(t, d.Rules, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.Same(t, g, d.GlobalChunker)
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// French settings (no invent):
	// global: ignoreSpelling + NO removePreviousTags
	// multiwords: removePreviousTags + NO ignoreSpelling
	require.True(t, g.AddIgnoreSpelling, "French chunkerGlobal.setIgnoreSpelling(true)")
	require.False(t, g.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	require.True(t, mw.RemovePreviousTags, "French chunker.setRemovePreviousTags(true)")
	require.False(t, mw.AddIgnoreSpelling, "French multiwords does NOT setIgnoreSpelling (unlike global/NL)")
}

// TestFrenchHybridDisambiguator_OrderGlobalThenMultiwordThenXML proves stage
// isolation vs full Java order with Java-visible POS / ignore_spelling outcomes.
func TestFrenchHybridDisambiguator_OrderGlobalThenMultiwordThenXML(t *testing.T) {
	requireFRHybridResources(t)

	g := FrenchGlobalChunker()
	mw := FrenchMultiWordChunker()
	xml := FrenchXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	// Isolation hybrids (only one stage) vs full order.
	onlyGlobal := &FrenchHybridDisambiguator{GlobalChunker: g}
	onlyMulti := &FrenchHybridDisambiguator{Chunker: mw}
	onlyXML := &FrenchHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(g.Disambiguate(sent)))
	}
	full := NewFrenchHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Global-only phrase: "Microsoft Entra" in spelling_global, not fr/multiwords ---
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
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "Z e sp") || hasExactPOS(tags, "Z m sp"),
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
		noGlobal := &FrenchHybridDisambiguator{Chunker: mw, Rules: xml}
		requireAllContentIgnored(t, noGlobal.Disambiguate(fresh()), false,
			label+" without global must not ignore")
	}

	// Global-only casing flags: allowFirstCapitalized=false on GlobalChunker
	// "picture alliance" matches; "Picture alliance" does not (and is not in multiwords).
	{
		label := "picture alliance"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("picture", "alliance"))
		}
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Picture", "alliance"))
		}

		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global exact")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), false, label+" multi no match")
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid")
		requireNoGlobalInventPOS(t, full.Disambiguate(fresh()), label+" full")

		// first-cap denied by GlobalChunker (allowFirstCapitalized=false) and absent from multiwords
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshCap()), false, "Picture alliance first-cap denied global")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(freshCap()), false, "Picture alliance not in multiwords")
		requireAllContentIgnored(t, full.Disambiguate(freshCap()), false, "Picture alliance full no match")
	}

	// --- (2) Multiword-only phrase: "home page" in fr/multiwords, not spelling_global ---
	// Official line: home page;N f s → after removePreviousTags: N f s / J f s
	// NO setIgnoreSpelling on French multiwords → content tokens not ignored.
	{
		label := "home page"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("home", "page"))
		}

		gotG := contentPOSTags(onlyGlobal.Disambiguate(fresh()))
		for i, tags := range gotG {
			require.False(t, hasExactPOS(tags, "N f s") || hasExactPOS(tags, "J f s") || hasAnyAnglePOS(tags),
				"%s global-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no ignore")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "N f s"), "%s multiword-only home: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "J f s"), "%s multiword-only page: %v", label, gotM[1])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]),
			"%s multiword removePreviousTags flattens angles: %v %v", label, gotM[0], gotM[1])
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), false,
			label+" multiword-only NO ignore (French multiwords has no setIgnoreSpelling)")

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "N f s"), "%s full hybrid home: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "J f s"), "%s full hybrid page: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]))
		requireAllContentIgnored(t, full.Disambiguate(fresh()), false,
			label+" full hybrid NO ignore (global miss + multiword no ignoreSpelling)")

		// Without multiword stage: global+XML must not invent multiword POS.
		noMulti := &FrenchHybridDisambiguator{GlobalChunker: g, Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "N f s") || hasExactPOS(tags, "J f s") || hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		// allowFirstCapitalized=true on multiwords: Home page matches multiword stage
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Home", "page"))
		}
		gotCap := contentPOSTags(full.Disambiguate(freshCap()))
		require.True(t, hasExactPOS(gotCap[0], "N f s"), "Home page full hybrid first-cap: %v", gotCap[0])
		require.True(t, hasExactPOS(gotCap[1], "J f s"), "Home page full hybrid first-cap: %v", gotCap[1])
	}

	// Multiword-only "point presse" (N m s → N m s / J m s) and "capture d'écran"
	{
		label := "point presse"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("point", "presse"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global no")
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "N m s"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "J m s"), "%s multi: %v", label, gotM[1])
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "N m s"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "J m s"), "%s full: %v", label, gotFull[1])
		requireAllContentIgnored(t, full.Disambiguate(fresh()), false, label+" full no ignore")
	}

	// --- (3) Shared phrase: "Google Maps" listed in BOTH global and multiwords ---
	// Official multiwords: Google Maps; Z e sp → after removePreviousTags: Z e sp Z e sp
	// Official global: tagForNotAddingTags + ignore
	// Java order: global ignores first; multiword adds POS; no invent from global
	{
		label := "Google Maps"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps"))
		}

		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, true, label+" global-only")
		requireNoGlobalInventPOS(t, outG, label+" global-only")

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "Z e sp"), "%s multiword-only: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "Z e sp"), "%s multiword-only: %v", label, gotM[1])
		require.False(t, hasAnyAnglePOS(gotM[0]) || hasAnyAnglePOS(gotM[1]))
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), false,
			label+" multiword-only NO ignore")

		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		require.True(t, hasExactPOS(gotFull[0], "Z e sp"),
			"%s full hybrid multiword POS after global ignore: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "Z e sp"),
			"%s full hybrid Maps: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must flatten multiword angles: %v %v", label, gotFull[0], gotFull[1])
		// Global stage ran → ignore survives multiword (multiword does not clear ignore).
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore from global")

		// Without multiword: ignore still, but no Z e sp POS.
		noMulti := &FrenchHybridDisambiguator{GlobalChunker: g, Rules: xml}
		outNoM := noMulti.Disambiguate(fresh())
		requireAllContentIgnored(t, outNoM, true, label+" without multi ignore")
		for i, tags := range contentPOSTags(outNoM) {
			require.False(t, hasExactPOS(tags, "Z e sp") || hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must not have multiword POS: %v", label, i, tags)
		}

		// Without global: multiword POS, no ignore.
		noGlobal := &FrenchHybridDisambiguator{Chunker: mw, Rules: xml}
		outNoG := noGlobal.Disambiguate(fresh())
		gotNoG := contentPOSTags(outNoG)
		require.True(t, hasExactPOS(gotNoG[0], "Z e sp"), label+" without global still multi POS")
		requireAllContentIgnored(t, outNoG, false, label+" without global no ignore")

		// allowAllUppercase=true on both chunkers.
		// Official multiwords.txt lists Google Maps twice (tab Z m s earlier; ; Z e sp later).
		// Exact "Google Maps" keeps later Z e sp; all-upper GOOGLE MAPS keeps first-entry Z m s
		// because getTokenLettercaseVariants skips keys already present in the map.
		freshUpper := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("GOOGLE", "MAPS"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshUpper()), true, "GOOGLE MAPS global")
		gotUpperM := contentPOSTags(onlyMulti.Disambiguate(freshUpper()))
		require.True(t, hasExactPOS(gotUpperM[0], "Z m s"), "GOOGLE MAPS multi (first official entry all-upper): %v", gotUpperM[0])
		require.True(t, hasExactPOS(gotUpperM[1], "Z m s"), "GOOGLE MAPS multi token1: %v", gotUpperM[1])
		outUpperFull := full.Disambiguate(freshUpper())
		requireAllContentIgnored(t, outUpperFull, true, "GOOGLE MAPS full ignore")
		gotUpperFull := contentPOSTags(outUpperFull)
		require.True(t, hasExactPOS(gotUpperFull[0], "Z m s"), "GOOGLE MAPS full POS: %v", gotUpperFull[0])
		require.True(t, hasExactPOS(gotUpperFull[1], "Z m s"), "GOOGLE MAPS full POS1: %v", gotUpperFull[1])
	}

	// Shared "Intel Core" (Z m sp on multiwords; global ignore)
	{
		label := "Intel Core"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Intel", "Core"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global")
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "Z m sp"), "%s multi: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "Z m sp"), "%s multi: %v", label, gotM[1])
		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		require.True(t, hasExactPOS(gotFull[0], "Z m sp"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "Z m sp"), "%s full: %v", label, gotFull[1])
		requireAllContentIgnored(t, outFull, true, label+" full ignore from global")
	}

	// --- (4) XML-only effects: fr IGNORE_SPELLING_OF_NUMBERS + global proper noun ---
	// Chunkers do not match these surfaces; XML stage sets ignore_spelling.
	// Proves XML runs after chunkers in full hybrid (and isolation).
	{
		// fr rule: [A-Z]\d+ → A4
		sentA4 := tokenSentence("A4")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentA4), "A4")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentA4), "A4")
		requireIgnored(t, onlyXML.Disambiguate(sentA4), "A4")
		requireIgnored(t, full.Disambiguate(sentA4), "A4")
		requireIgnored(t, javaOrder(sentA4), "A4")

		// Without XML: chunkers alone must not ignore A4
		noXML := &FrenchHybridDisambiguator{GlobalChunker: g, Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("A4")), "A4")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// fr 5e / 4x4 (XML after chunkers still fires)
		requireIgnored(t, full.Disambiguate(tokenSentence("5e")), "5e")
		requireIgnored(t, full.Disambiguate(tokenSentence("4x4")), "4x4")
	}
}

// TestFrenchHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockOutcomes proves
// XML last does not wipe ignore_spelling / multiword POS set by earlier stages.
func TestFrenchHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockOutcomes(t *testing.T) {
	requireFRHybridResources(t)
	full := NewFrenchHybridDisambiguator()

	// Global-only ignore survives XML; no invent POS.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")))
	requireAllContentIgnored(t, out, true, "Microsoft Entra after full order")
	requireNoGlobalInventPOS(t, out, "Microsoft Entra after full")

	// Multiword-only POS survives XML; no ignore from multiwords.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("home", "page")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "N f s"), "home after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "J f s"), "page after full order: %v", got[1])
	requireAllContentIgnored(t, out, false, "home page multiword no ignore after full")

	// Shared: ignore from global + multiword POS survive XML.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "Z e sp"), "Google after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "Z e sp"), "Maps after full order: %v", got[1])
	requireAllContentIgnored(t, out, true, "Google Maps ignore after full")

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Intel", "Core")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "Z m sp"), "Intel after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "Z m sp"), "Core after full order: %v", got[1])
	requireAllContentIgnored(t, out, true, "Intel Core ignore after full")

	// XML effects still fire after chunkers.
	requireIgnored(t, full.Disambiguate(tokenSentence("A4")), "A4")
	requireIgnored(t, full.Disambiguate(tokenSentence("QB|LT")), "QB|LT")
}

// TestFrenchHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == xml(mw(g(input))) for official isolation surfaces.
func TestFrenchHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireFRHybridResources(t)
	g := FrenchGlobalChunker()
	mw := FrenchMultiWordChunker()
	xml := FrenchXmlRuleDisambiguator()
	full := NewFrenchHybridDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		parts    []string
		label    string
		wantIg   bool
		wantPOS0 string // empty = no multiword POS required
		wantPOS1 string
	}
	cases := []caseT{
		{[]string{"Microsoft", "Entra"}, "Microsoft Entra", true, "", ""},
		{[]string{"picture", "alliance"}, "picture alliance", true, "", ""},
		{[]string{"Picture", "alliance"}, "Picture alliance", false, "", ""},
		{[]string{"home", "page"}, "home page", false, "N f s", "J f s"},
		{[]string{"point", "presse"}, "point presse", false, "N m s", "J m s"},
		{[]string{"Google", "Maps"}, "Google Maps", true, "Z e sp", "Z e sp"},
		{[]string{"Intel", "Core"}, "Intel Core", true, "Z m sp", "Z m sp"},
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed", false, "", ""},
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
		} else {
			// Global-only / non-listed: no multiword invent; tagForNotAddingTags on global.
			requireNoGlobalInventPOS(t, outFull, tc.label+" full")
			requireNoGlobalInventPOS(t, outManual, tc.label+" manual")
		}
		// Per-token POS set parity (order of readings may differ slightly — compare membership).
		for i := range gotFull {
			require.ElementsMatch(t, gotFull[i], gotManual[i],
				"%s token[%d] POS parity full vs javaOrder", tc.label, i)
		}
	}
}
