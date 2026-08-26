package nl

// Outcome twins for DutchHybridDisambiguator full stage order:
// Java DutchHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
// i.e. spelling_global (tagForNotAddingTags + setIgnoreSpelling) →
//     /nl/multiwords.txt (tagForNotAddingTags + setIgnoreSpelling; NO setRemovePreviousTags) →
//     XmlRuleDisambiguator(lang, true).
// Official dutch.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/global/XML stage leaves).
//
// Differs from Spanish/Portuguese hybrid:
//   - both chunkers use MultiWordChunker.tagForNotAddingTags (no invent open/close POS)
//   - setIgnoreSpelling(true) on BOTH chunkers
//   - NO setRemovePreviousTags on multiwords (ES/PT set it true → angle flatten)
//   - allowFirstCapitalized: global=false, multiwords=true

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func requireNLHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverDutchGlobalChunker() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	if DiscoverDutchMultiwords() == "" {
		t.Skip("official nl/multiwords.txt not discoverable")
	}
	if DiscoverDutchDisambiguationXML() == "" {
		t.Skip("official nl/disambiguation.xml not discoverable")
	}
	if DiscoverGlobalDisambiguationXML() == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
}

func contentPOSTags(out *languagetool.AnalyzedSentence) [][]string {
	var all [][]string
	for i, tr := range out.GetTokens() {
		if i == 0 || tr.IsWhitespace() {
			continue
		}
		var tags []string
		for _, r := range tr.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
		all = append(all, tags)
	}
	return all
}

func hasAnyAnglePOS(tags []string) bool {
	for _, p := range tags {
		if strings.Contains(p, "<") {
			return true
		}
	}
	return false
}

func requireNoChunkPOS(t *testing.T, out *languagetool.AnalyzedSentence, label string) {
	t.Helper()
	for i, tags := range contentPOSTags(out) {
		require.False(t, hasAnyAnglePOS(tags),
			"%s token[%d] tagForNotAddingTags must not invent angle POS: %v", label, i, tags)
		for _, p := range tags {
			require.NotEqual(t, disambiguation.TagForNotAddingTags, p,
				"%s token[%d] must not surface internal _NONE_ tag: %v", label, i, tags)
			// Romance-style multiword tags must not appear under Dutch tagForNotAddingTags.
			require.False(t, p == "NPCN000" || strings.HasPrefix(p, "<NPCN") || strings.HasPrefix(p, "</NPCN"),
				"%s token[%d] must not invent NPCN000-style chunk POS: %v", label, i, tags)
		}
	}
}

// TestNewDutchHybridDisambiguator_WiresAllThreeStages proves Java constructor
// eagerly builds GlobalChunker, multiwords Chunker, and XmlRuleDisambiguator when
// the same official resources Java loads are present — with Dutch flags.
func TestNewDutchHybridDisambiguator_WiresAllThreeStages(t *testing.T) {
	requireNLHybridResources(t)

	g := DutchGlobalChunker()
	mw := DutchMultiWordChunker()
	xml := DutchXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewDutchHybridDisambiguator()
	require.NotNil(t, d.GlobalChunker,
		"chunkerGlobal = MultiWordChunker.getInstance(/spelling_global.txt, false, true, false, tagForNotAddingTags)")
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/nl/multiwords.txt, true, true, false, tagForNotAddingTags)")
	require.NotNil(t, d.Rules, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.Same(t, g, d.GlobalChunker)
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Dutch settings (no invent): ignoreSpelling on BOTH; NO removePreviousTags on either.
	require.True(t, g.AddIgnoreSpelling, "Dutch chunkerGlobal.setIgnoreSpelling(true)")
	require.False(t, g.RemovePreviousTags, "GlobalChunker does NOT setRemovePreviousTags")
	require.True(t, mw.AddIgnoreSpelling, "Dutch chunker.setIgnoreSpelling(true)")
	require.False(t, mw.RemovePreviousTags, "Dutch multiwords does NOT setRemovePreviousTags (unlike ES/PT)")
}

// TestDutchHybridDisambiguator_OrderGlobalThenMultiwordThenXML proves stage
// isolation vs full Java order with Java-visible ignore_spelling / tagForNotAddingTags outcomes.
func TestDutchHybridDisambiguator_OrderGlobalThenMultiwordThenXML(t *testing.T) {
	requireNLHybridResources(t)

	g := DutchGlobalChunker()
	mw := DutchMultiWordChunker()
	xml := DutchXmlRuleDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	// Isolation hybrids (only one stage) vs full order.
	onlyGlobal := &DutchHybridDisambiguator{GlobalChunker: g}
	onlyMulti := &DutchHybridDisambiguator{Chunker: mw}
	onlyXML := &DutchHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(g.Disambiguate(sent)))
	}
	// Wrong: multiword before global (still ignore-commutative for tagForNotAddingTags,
	// but used with isolation to prove each stage is required).
	full := NewDutchHybridDisambiguator()
	require.NotNil(t, full.GlobalChunker)
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Global-only phrase: "Microsoft Entra" in spelling_global, not nl/multiwords ---
	// Global alone → ignore_spelling; no invent POS (tagForNotAddingTags)
	// Multiword alone → no ignore
	// XML alone → no ignore (no matching rule on this surface alone)
	// Full hybrid → ignore (global stage ran)
	// Manual javaOrder → same as full
	// Without global stage → no ignore
	{
		label := "Microsoft Entra"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra"))
		}

		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, true, label+" global-only ignore")
		requireNoChunkPOS(t, outG, label+" global-only")

		outM := onlyMulti.Disambiguate(fresh())
		requireAllContentIgnored(t, outM, false, label+" multiword-only no ignore")
		requireNoChunkPOS(t, outM, label+" multiword-only")

		outX := onlyXML.Disambiguate(fresh())
		requireAllContentIgnored(t, outX, false, label+" xml-only no ignore")
		requireNoChunkPOS(t, outX, label+" xml-only")

		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore")
		requireNoChunkPOS(t, outFull, label+" full hybrid")

		outJO := javaOrder(fresh())
		requireAllContentIgnored(t, outJO, true, label+" javaOrder ignore")
		requireNoChunkPOS(t, outJO, label+" javaOrder")

		// Without GlobalChunker: multiword+XML must not ignore this global-only surface.
		noGlobal := &DutchHybridDisambiguator{Chunker: mw, Rules: xml}
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
		requireNoChunkPOS(t, full.Disambiguate(fresh()), label+" full")

		// first-cap denied by GlobalChunker (allowFirstCapitalized=false) and absent from multiwords
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshCap()), false, "Picture alliance first-cap denied global")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(freshCap()), false, "Picture alliance not in multiwords")
		requireAllContentIgnored(t, full.Disambiguate(freshCap()), false, "Picture alliance full no match")
	}

	// --- (2) Multiword-only phrase: "A fortiori" / "carpe diem" in nl/multiwords, not spelling_global ---
	{
		label := "A fortiori"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("A", "fortiori"))
		}

		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no ignore")
		outM := onlyMulti.Disambiguate(fresh())
		requireAllContentIgnored(t, outM, true, label+" multiword-only ignore")
		requireNoChunkPOS(t, outM, label+" multiword-only")

		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore")
		requireNoChunkPOS(t, outFull, label+" full hybrid")

		// Without multiword stage: global+XML must not ignore multiword-only surface.
		noMulti := &DutchHybridDisambiguator{GlobalChunker: g, Rules: xml}
		requireAllContentIgnored(t, noMulti.Disambiguate(fresh()), false,
			label+" without multiword must not ignore")

		// Lowercase official multiword entry
		freshLower := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("a", "fortiori"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshLower()), false, "a fortiori global no")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(freshLower()), true, "a fortiori multi yes")
		requireAllContentIgnored(t, full.Disambiguate(freshLower()), true, "a fortiori full yes")
	}

	{
		label := "carpe diem"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("carpe", "diem"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(fresh()), true, label+" multiword-only yes")
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full yes")
		requireNoChunkPOS(t, full.Disambiguate(fresh()), label+" full")
	}

	// --- (3) Shared phrase: "Google Maps" listed in BOTH global and multiwords ---
	// Both stages ignore; neither invents POS (tagForNotAddingTags). No removePreviousTags flatten.
	{
		label := "Google Maps"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps"))
		}

		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, true, label+" global-only")
		requireNoChunkPOS(t, outG, label+" global-only")

		outM := onlyMulti.Disambiguate(fresh())
		requireAllContentIgnored(t, outM, true, label+" multiword-only")
		requireNoChunkPOS(t, outM, label+" multiword-only")

		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid")
		requireNoChunkPOS(t, outFull, label+" full hybrid")

		// allowAllUppercase=true on both chunkers
		freshUpper := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("GOOGLE", "MAPS"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshUpper()), true, "GOOGLE MAPS global")
		requireAllContentIgnored(t, onlyMulti.Disambiguate(freshUpper()), true, "GOOGLE MAPS multi")
		requireAllContentIgnored(t, full.Disambiguate(freshUpper()), true, "GOOGLE MAPS full")
	}

	// --- (4) XML-only effects: nl ROADS / global proper noun / roman ---
	// Chunkers do not match these surfaces; XML stage sets ignore_spelling.
	// Proves XML runs after chunkers in full hybrid (and isolation).
	{
		// nl ROADS: A12
		sentA12 := tokenSentence("A12")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentA12), "A12")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentA12), "A12")
		requireIgnored(t, onlyXML.Disambiguate(sentA12), "A12")
		requireIgnored(t, full.Disambiguate(sentA12), "A12")
		requireIgnored(t, javaOrder(sentA12), "A12")

		// Without XML: chunkers alone must not ignore A12
		noXML := &DutchHybridDisambiguator{GlobalChunker: g, Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("A12")), "A12")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// nl IGNORE_SPELLER_ROMAN_NUMBERS: XIV (XML after chunkers still fires)
		sentXIV := tokenSentence("XIV")
		requireIgnored(t, full.Disambiguate(sentXIV), "XIV")

		// PLANES: PH-ABC
		requireIgnored(t, full.Disambiguate(tokenSentence("PH-ABC")), "PH-ABC")
	}
}

// TestDutchHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkIgnore proves
// XML last does not wipe ignore_spelling set by multiword/global stages.
func TestDutchHybridDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkIgnore(t *testing.T) {
	requireNLHybridResources(t)
	full := NewDutchHybridDisambiguator()

	// Global-only ignore survives XML stage.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")))
	requireAllContentIgnored(t, out, true, "Microsoft Entra after full order")
	requireNoChunkPOS(t, out, "Microsoft Entra after full")

	// Multiword-only ignore survives XML stage.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("A", "fortiori")))
	requireAllContentIgnored(t, out, true, "A fortiori after full order")
	requireNoChunkPOS(t, out, "A fortiori after full")

	// Shared phrase ignore survives XML; still no invent POS.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps")))
	requireAllContentIgnored(t, out, true, "Google Maps after full order")
	requireNoChunkPOS(t, out, "Google Maps after full")

	// XML effects still fire after chunkers.
	requireIgnored(t, full.Disambiguate(tokenSentence("A12")), "A12")
	requireIgnored(t, full.Disambiguate(tokenSentence("QB|LT")), "QB|LT")
}

// TestDutchHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == xml(mw(g(input))) for official isolation surfaces.
func TestDutchHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireNLHybridResources(t)
	g := DutchGlobalChunker()
	mw := DutchMultiWordChunker()
	xml := DutchXmlRuleDisambiguator()
	full := NewDutchHybridDisambiguator()
	require.NotNil(t, g)
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		parts  []string
		label  string
		wantIg bool
	}
	cases := []caseT{
		{[]string{"Microsoft", "Entra"}, "Microsoft Entra", true},
		{[]string{"A", "fortiori"}, "A fortiori", true},
		{[]string{"Google", "Maps"}, "Google Maps", true},
		{[]string{"carpe", "diem"}, "carpe diem", true},
		{[]string{"picture", "alliance"}, "picture alliance", true},
		{[]string{"Picture", "alliance"}, "Picture alliance", false},
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed", false},
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
		requireNoChunkPOS(t, outFull, tc.label+" full")
		requireNoChunkPOS(t, outManual, tc.label+" manual")
	}
}
