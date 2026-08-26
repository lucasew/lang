package de

// Outcome twins for GermanRuleDisambiguator full stage order:
// Java GermanRuleDisambiguator.disambiguate(input, callback):
//   disambiguator.disambiguate(
//     multitokenSpeller2.disambiguate(           // multitoken-suggest.txt
//       multitokenSpeller3.disambiguate(         // spelling_global.txt
//         multitokenSpeller.disambiguate(input), // multitoken-ignore.txt
//       ),
//     ),
//   )
// i.e. ignore → global → suggest → XmlRuleDisambiguator(lang, true).
//
// All three MultiWordChunkers:
//   tagForNotAddingTags + setIgnoreSpelling(true); NO setRemovePreviousTags.
// Flags: ignore/suggest allowFirstCapitalized=true; global allowFirstCapitalized=false.
// Official german.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed ignore/global/suggest/XML stage leaves).
//
// Closest to Dutch hybrid (tagForNotAddingTags + ignore on chunkers) but four stages.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func requireDEHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverGermanMultitokenIgnore() == "" {
		t.Skip("official de/multitoken-ignore.txt not discoverable")
	}
	if DiscoverGermanMultitokenGlobal() == "" {
		t.Skip("official spelling_global.txt not discoverable")
	}
	if DiscoverGermanMultitokenSuggest() == "" {
		t.Skip("official de/multitoken-suggest.txt not discoverable")
	}
	if DiscoverGermanDisambiguationXML() == "" {
		t.Skip("official de/disambiguation.xml not discoverable")
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

// requireNoChunkPOS asserts German MultiWordChunker tagForNotAddingTags:
// no angle POS, no surface _NONE_, no Romance NPCN000 invent.
func requireNoChunkPOS(t *testing.T, out *languagetool.AnalyzedSentence, label string) {
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

// TestNewGermanRuleDisambiguator_WiresAllFourStages proves Java constructor
// eagerly builds MultitokenIgnore, MultitokenGlobal, MultitokenSuggest, and
// XmlRuleDisambiguator when the same official resources Java loads are present.
func TestNewGermanRuleDisambiguator_WiresAllFourStages(t *testing.T) {
	requireDEHybridResources(t)

	ign := GermanMultitokenIgnore()
	g := GermanMultitokenGlobal()
	sug := GermanMultitokenSuggest()
	xml := GermanXmlRuleDisambiguator()
	require.NotNil(t, ign)
	require.NotNil(t, g)
	require.NotNil(t, sug)
	require.NotNil(t, xml)

	d := NewGermanRuleDisambiguator()
	require.NotNil(t, d.MultitokenIgnore,
		"multitokenSpeller = MultiWordChunker.getInstance(/de/multitoken-ignore.txt, true, true, false, tagForNotAddingTags)")
	require.NotNil(t, d.MultitokenGlobal,
		"multitokenSpeller3 = MultiWordChunker.getInstance(/spelling_global.txt, false, true, false, tagForNotAddingTags)")
	require.NotNil(t, d.MultitokenSuggest,
		"multitokenSpeller2 = MultiWordChunker.getInstance(/de/multitoken-suggest.txt, true, true, false, tagForNotAddingTags)")
	require.NotNil(t, d.Rules, "disambiguator = new XmlRuleDisambiguator(lang, true)")
	require.Same(t, ign, d.MultitokenIgnore)
	require.Same(t, g, d.MultitokenGlobal)
	require.Same(t, sug, d.MultitokenSuggest)
	require.Same(t, xml, d.Rules)

	// German settings (no invent): ignoreSpelling on ALL three chunkers; no removePreviousTags.
	require.True(t, ign.AddIgnoreSpelling, "German multitokenSpeller.setIgnoreSpelling(true)")
	require.False(t, ign.RemovePreviousTags, "MultitokenIgnore does NOT setRemovePreviousTags")
	require.True(t, g.AddIgnoreSpelling, "German multitokenSpeller3.setIgnoreSpelling(true)")
	require.False(t, g.RemovePreviousTags, "MultitokenGlobal does NOT setRemovePreviousTags")
	require.True(t, sug.AddIgnoreSpelling, "German multitokenSpeller2.setIgnoreSpelling(true)")
	require.False(t, sug.RemovePreviousTags, "MultitokenSuggest does NOT setRemovePreviousTags")
}

// TestGermanRuleDisambiguator_OrderIgnoreGlobalSuggestXML proves stage
// isolation vs full Java order with Java-visible ignore_spelling / tagForNotAddingTags outcomes.
func TestGermanRuleDisambiguator_OrderIgnoreGlobalSuggestXML(t *testing.T) {
	requireDEHybridResources(t)

	ign := GermanMultitokenIgnore()
	g := GermanMultitokenGlobal()
	sug := GermanMultitokenSuggest()
	xml := GermanXmlRuleDisambiguator()
	require.NotNil(t, ign)
	require.NotNil(t, g)
	require.NotNil(t, sug)
	require.NotNil(t, xml)

	// Isolation hybrids (only one stage) vs full order.
	onlyIgnore := &GermanRuleDisambiguator{MultitokenIgnore: ign}
	onlyGlobal := &GermanRuleDisambiguator{MultitokenGlobal: g}
	onlySuggest := &GermanRuleDisambiguator{MultitokenSuggest: sug}
	onlyXML := &GermanRuleDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		// ignore → global → suggest → XML
		return xml.Disambiguate(sug.Disambiguate(g.Disambiguate(ign.Disambiguate(sent))))
	}
	full := NewGermanRuleDisambiguator()
	require.NotNil(t, full.MultitokenIgnore)
	require.NotNil(t, full.MultitokenGlobal)
	require.NotNil(t, full.MultitokenSuggest)
	require.NotNil(t, full.Rules)

	// --- (1) Ignore-only phrase: "3-adische System" in multitoken-ignore, not global/suggest ---
	// Ignore alone → ignore_spelling; no invent POS
	// Global/Suggest/XML alone → no ignore
	// Full hybrid → ignore (ignore stage ran)
	// Without ignore stage → no ignore
	{
		label := "3-adische System"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("3-adische", "System"))
		}

		outI := onlyIgnore.Disambiguate(fresh())
		requireAllContentIgnored(t, outI, true, label+" ignore-only ignore")
		requireNoChunkPOS(t, outI, label+" ignore-only")

		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, false, label+" global-only no ignore")
		requireNoChunkPOS(t, outG, label+" global-only")

		outS := onlySuggest.Disambiguate(fresh())
		requireAllContentIgnored(t, outS, false, label+" suggest-only no ignore")
		requireNoChunkPOS(t, outS, label+" suggest-only")

		outX := onlyXML.Disambiguate(fresh())
		requireAllContentIgnored(t, outX, false, label+" xml-only no ignore")
		requireNoChunkPOS(t, outX, label+" xml-only")

		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore")
		requireNoChunkPOS(t, outFull, label+" full hybrid")

		outJO := javaOrder(fresh())
		requireAllContentIgnored(t, outJO, true, label+" javaOrder ignore")
		requireNoChunkPOS(t, outJO, label+" javaOrder")

		// Without MultitokenIgnore: global+suggest+XML must not ignore this ignore-only surface.
		noIgnore := &GermanRuleDisambiguator{MultitokenGlobal: g, MultitokenSuggest: sug, Rules: xml}
		requireAllContentIgnored(t, noIgnore.Disambiguate(fresh()), false,
			label+" without ignore must not ignore")
	}

	// Ignore-only "Kelassurier Mauer" + /N expansion "Kelassurier Mauern"
	{
		label := "Kelassurier Mauer"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Kelassurier", "Mauer"))
		}
		requireAllContentIgnored(t, onlyIgnore.Disambiguate(fresh()), true, label+" ignore-only")
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no")
		requireAllContentIgnored(t, onlySuggest.Disambiguate(fresh()), false, label+" suggest-only no")
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid")
		requireNoChunkPOS(t, full.Disambiguate(fresh()), label+" full")

		freshN := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Kelassurier", "Mauern"))
		}
		requireAllContentIgnored(t, onlyIgnore.Disambiguate(freshN()), true, "Kelassurier Mauern /N")
		requireAllContentIgnored(t, full.Disambiguate(freshN()), true, "Kelassurier Mauern full")
		// Wrong suffix not listed
		freshWrong := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Kelassurier", "Mauers"))
		}
		requireAllContentIgnored(t, onlyIgnore.Disambiguate(freshWrong()), false, "Kelassurier Mauers")
		requireAllContentIgnored(t, full.Disambiguate(freshWrong()), false, "Kelassurier Mauers full")
	}

	// --- (2) Global-only phrase: "Microsoft Entra" in spelling_global, not ignore/suggest ---
	{
		label := "Microsoft Entra"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra"))
		}

		requireAllContentIgnored(t, onlyIgnore.Disambiguate(fresh()), false, label+" ignore-only no ignore")
		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, true, label+" global-only ignore")
		requireNoChunkPOS(t, outG, label+" global-only")

		requireAllContentIgnored(t, onlySuggest.Disambiguate(fresh()), false, label+" suggest-only no ignore")
		requireAllContentIgnored(t, onlyXML.Disambiguate(fresh()), false, label+" xml-only no ignore")

		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore")
		requireNoChunkPOS(t, outFull, label+" full hybrid")

		// Without MultitokenGlobal: ignore+suggest+XML must not ignore this global-only surface.
		noGlobal := &GermanRuleDisambiguator{MultitokenIgnore: ign, MultitokenSuggest: sug, Rules: xml}
		requireAllContentIgnored(t, noGlobal.Disambiguate(fresh()), false,
			label+" without global must not ignore")
	}

	// Global-only casing: allowFirstCapitalized=false on MultitokenGlobal
	// "picture alliance" matches; "Picture alliance" does not (and is not in ignore/suggest).
	{
		label := "picture alliance"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("picture", "alliance"))
		}
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Picture", "alliance"))
		}

		requireAllContentIgnored(t, onlyIgnore.Disambiguate(fresh()), false, label+" ignore no")
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), true, label+" global exact")
		requireAllContentIgnored(t, onlySuggest.Disambiguate(fresh()), false, label+" suggest no")
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full hybrid")
		requireNoChunkPOS(t, full.Disambiguate(fresh()), label+" full")

		// first-cap denied by MultitokenGlobal (allowFirstCapitalized=false)
		// and absent from MultitokenIgnore/Suggest.
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshCap()), false, "Picture alliance first-cap denied global")
		requireAllContentIgnored(t, onlyIgnore.Disambiguate(freshCap()), false, "Picture alliance not in ignore")
		requireAllContentIgnored(t, onlySuggest.Disambiguate(freshCap()), false, "Picture alliance not in suggest")
		requireAllContentIgnored(t, full.Disambiguate(freshCap()), false, "Picture alliance full no match")
	}

	// --- (3) Suggest-only phrase: "New York" in multitoken-suggest, not ignore/global ---
	// (global has "New York Times" etc., not bare "New York")
	{
		label := "New York"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("New", "York"))
		}

		requireAllContentIgnored(t, onlyIgnore.Disambiguate(fresh()), false, label+" ignore-only no ignore")
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global-only no ignore")
		outS := onlySuggest.Disambiguate(fresh())
		requireAllContentIgnored(t, outS, true, label+" suggest-only ignore")
		requireNoChunkPOS(t, outS, label+" suggest-only")

		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid ignore")
		requireNoChunkPOS(t, outFull, label+" full hybrid")

		// Without MultitokenSuggest: ignore+global+XML must not ignore this suggest-only surface.
		noSuggest := &GermanRuleDisambiguator{MultitokenIgnore: ign, MultitokenGlobal: g, Rules: xml}
		requireAllContentIgnored(t, noSuggest.Disambiguate(fresh()), false,
			label+" without suggest must not ignore")

		// /S expansion + allowFirstCapitalized on suggest (a cappella / A cappella)
		freshS := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("New", "Yorks"))
		}
		requireAllContentIgnored(t, onlySuggest.Disambiguate(freshS()), true, "New Yorks /S")
		requireAllContentIgnored(t, full.Disambiguate(freshS()), true, "New Yorks full")

		freshA := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("a", "cappella"))
		}
		freshACap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("A", "cappella"))
		}
		requireAllContentIgnored(t, onlySuggest.Disambiguate(freshA()), true, "a cappella suggest")
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshA()), false, "a cappella global no")
		requireAllContentIgnored(t, full.Disambiguate(freshA()), true, "a cappella full")
		// MultitokenSuggest allowFirstCapitalized=true (unlike MultitokenGlobal)
		requireAllContentIgnored(t, onlySuggest.Disambiguate(freshACap()), true, "A cappella first-cap allowed suggest")
		requireAllContentIgnored(t, full.Disambiguate(freshACap()), true, "A cappella full")
	}

	// Additional suggest-only surfaces used by stage leaf tests.
	for _, parts := range [][]string{
		{"à", "la", "carte"},
		{"Alma", "Mater"},
		{"Deus", "ex", "Machina"},
		{"Osama", "bin", "Laden"},
		{"Human-centered", "Design"},
	} {
		label := strings.Join(parts, " ")
		fresh := func(p []string) func() *languagetool.AnalyzedSentence {
			return func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens(p...))
			}
		}(parts)
		requireAllContentIgnored(t, onlyIgnore.Disambiguate(fresh()), false, label+" ignore no")
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(fresh()), false, label+" global no")
		requireAllContentIgnored(t, onlySuggest.Disambiguate(fresh()), true, label+" suggest yes")
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, label+" full yes")
		requireNoChunkPOS(t, full.Disambiguate(fresh()), label+" full")
	}

	// --- (4) Global-only "Google Maps" (not in DE ignore/suggest) + all-upper ---
	{
		label := "Google Maps"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Google", "Maps"))
		}
		requireAllContentIgnored(t, onlyIgnore.Disambiguate(fresh()), false, label+" ignore no")
		outG := onlyGlobal.Disambiguate(fresh())
		requireAllContentIgnored(t, outG, true, label+" global-only")
		requireNoChunkPOS(t, outG, label+" global-only")
		requireAllContentIgnored(t, onlySuggest.Disambiguate(fresh()), false, label+" suggest no")
		outFull := full.Disambiguate(fresh())
		requireAllContentIgnored(t, outFull, true, label+" full hybrid")
		requireNoChunkPOS(t, outFull, label+" full hybrid")

		// allowAllUppercase=true on MultitokenGlobal
		freshUpper := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("GOOGLE", "MAPS"))
		}
		requireAllContentIgnored(t, onlyGlobal.Disambiguate(freshUpper()), true, "GOOGLE MAPS global")
		requireAllContentIgnored(t, full.Disambiguate(freshUpper()), true, "GOOGLE MAPS full")
	}

	// --- (5) XML-only effects: de pack + global proper noun ---
	// Chunkers do not match these surfaces; XML stage sets ignore_spelling.
	// Proves XML runs after chunkers in full hybrid (and isolation).
	{
		// de ZWEIPFÜNDER: 2-Pfünder
		sent2p := tokenSentence("2-Pfünder")
		requireNotIgnored(t, onlyIgnore.Disambiguate(sent2p), "2-Pfünder")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sent2p), "2-Pfünder")
		requireNotIgnored(t, onlySuggest.Disambiguate(sent2p), "2-Pfünder")
		requireIgnored(t, onlyXML.Disambiguate(sent2p), "2-Pfünder")
		requireIgnored(t, full.Disambiguate(sent2p), "2-Pfünder")
		requireIgnored(t, javaOrder(sent2p), "2-Pfünder")

		// Without XML: chunkers alone must not ignore 2-Pfünder
		noXML := &GermanRuleDisambiguator{
			MultitokenIgnore:  ign,
			MultitokenGlobal:  g,
			MultitokenSuggest: sug,
		}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("2-Pfünder")), "2-Pfünder")

		// global XML: literal QB|LT
		sentQB := tokenSentence("QB|LT")
		requireNotIgnored(t, onlyIgnore.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlyGlobal.Disambiguate(sentQB), "QB|LT")
		requireNotIgnored(t, onlySuggest.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, onlyXML.Disambiguate(sentQB), "QB|LT")
		requireIgnored(t, full.Disambiguate(sentQB), "QB|LT")

		// de ENGLISCHE_WOERTER: Bingewatching
		requireIgnored(t, full.Disambiguate(tokenSentence("Bingewatching")), "Bingewatching")

		// de KREUCHT_UND_FLEUCHT: marker only on kreucht
		sentK := full.Disambiguate(tokenSentence("kreucht", "und", "fleucht"))
		requireIgnored(t, sentK, "kreucht")
		requireNotIgnored(t, sentK, "und", "fleucht")
	}
}

// TestGermanRuleDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkIgnore proves
// XML last does not wipe ignore_spelling set by ignore/global/suggest stages.
func TestGermanRuleDisambiguator_ChunkerBeforeXML_DoesNotBlockChunkIgnore(t *testing.T) {
	requireDEHybridResources(t)
	full := NewGermanRuleDisambiguator()

	// Ignore-only ignore survives XML stage.
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("3-adische", "System")))
	requireAllContentIgnored(t, out, true, "3-adische System after full order")
	requireNoChunkPOS(t, out, "3-adische System after full")

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Kelassurier", "Mauer")))
	requireAllContentIgnored(t, out, true, "Kelassurier Mauer after full order")
	requireNoChunkPOS(t, out, "Kelassurier Mauer after full")

	// Global-only ignore survives XML stage.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("Microsoft", "Entra")))
	requireAllContentIgnored(t, out, true, "Microsoft Entra after full order")
	requireNoChunkPOS(t, out, "Microsoft Entra after full")

	// Suggest-only ignore survives XML stage.
	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("New", "York")))
	requireAllContentIgnored(t, out, true, "New York after full order")
	requireNoChunkPOS(t, out, "New York after full")

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("à", "la", "carte")))
	requireAllContentIgnored(t, out, true, "à la carte after full order")
	requireNoChunkPOS(t, out, "à la carte after full")

	// XML effects still fire after chunkers.
	requireIgnored(t, full.Disambiguate(tokenSentence("2-Pfünder")), "2-Pfünder")
	requireIgnored(t, full.Disambiguate(tokenSentence("QB|LT")), "QB|LT")
}

// TestGermanRuleDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == xml(suggest(global(ignore(input)))) for official isolation surfaces.
func TestGermanRuleDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireDEHybridResources(t)
	ign := GermanMultitokenIgnore()
	g := GermanMultitokenGlobal()
	sug := GermanMultitokenSuggest()
	xml := GermanXmlRuleDisambiguator()
	full := NewGermanRuleDisambiguator()
	require.NotNil(t, ign)
	require.NotNil(t, g)
	require.NotNil(t, sug)
	require.NotNil(t, xml)

	type caseT struct {
		parts  []string
		label  string
		wantIg bool
	}
	cases := []caseT{
		// ignore-only
		{[]string{"3-adische", "System"}, "3-adische System", true},
		{[]string{"Kelassurier", "Mauer"}, "Kelassurier Mauer", true},
		{[]string{"Kelassurier", "Mauern"}, "Kelassurier Mauern", true},
		{[]string{"Kelassurier", "Mauers"}, "Kelassurier Mauers", false},
		// global-only
		{[]string{"Microsoft", "Entra"}, "Microsoft Entra", true},
		{[]string{"Google", "Maps"}, "Google Maps", true},
		{[]string{"picture", "alliance"}, "picture alliance", true},
		{[]string{"Picture", "alliance"}, "Picture alliance", false},
		// suggest-only
		{[]string{"New", "York"}, "New York", true},
		{[]string{"New", "Yorks"}, "New Yorks", true},
		{[]string{"à", "la", "carte"}, "à la carte", true},
		{[]string{"Alma", "Mater"}, "Alma Mater", true},
		{[]string{"a", "cappella"}, "a cappella", true},
		{[]string{"A", "cappella"}, "A cappella", true},
		// negative
		{[]string{"Zxqwv", "Plmnb"}, "random non-listed", false},
	}
	for _, tc := range cases {
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...))
		}
		outFull := full.Disambiguate(fresh())
		outManual := xml.Disambiguate(sug.Disambiguate(g.Disambiguate(ign.Disambiguate(fresh()))))
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

	// XML-only surfaces: composition parity for ignore flags set only by Rules stage.
	xmlCases := []struct {
		words  []string
		label  string
		wantIg map[string]bool
	}{
		{[]string{"2-Pfünder"}, "2-Pfünder", map[string]bool{"2-Pfünder": true}},
		{[]string{"QB|LT"}, "QB|LT", map[string]bool{"QB|LT": true}},
		{[]string{"Bingewatching"}, "Bingewatching", map[string]bool{"Bingewatching": true}},
		{[]string{"kreucht", "und", "fleucht"}, "kreucht und fleucht",
			map[string]bool{"kreucht": true, "und": false, "fleucht": false}},
	}
	for _, tc := range xmlCases {
		outFull := full.Disambiguate(tokenSentence(tc.words...))
		outManual := xml.Disambiguate(sug.Disambiguate(g.Disambiguate(ign.Disambiguate(tokenSentence(tc.words...)))))
		for surface, want := range tc.wantIg {
			trF := tokenBySurface(outFull, surface)
			trM := tokenBySurface(outManual, surface)
			require.NotNil(t, trF, "%s full missing %q", tc.label, surface)
			require.NotNil(t, trM, "%s manual missing %q", tc.label, surface)
			require.Equal(t, want, trF.IsIgnoredBySpeller(), "%s full %q", tc.label, surface)
			require.Equal(t, want, trM.IsIgnoredBySpeller(), "%s manual %q", tc.label, surface)
		}
	}
}

// TestGermanRuleDisambiguator_StageOrderIsIgnoreThenGlobalThenSuggestThenXML
// proves each stage occupies its Java slot via leave-one-out isolation.
func TestGermanRuleDisambiguator_StageOrderIsIgnoreThenGlobalThenSuggestThenXML(t *testing.T) {
	requireDEHybridResources(t)
	ign := GermanMultitokenIgnore()
	g := GermanMultitokenGlobal()
	sug := GermanMultitokenSuggest()
	xml := GermanXmlRuleDisambiguator()
	full := NewGermanRuleDisambiguator()

	// Unique surfaces: only the corresponding stage can set ignore.
	type stageCase struct {
		parts     []string
		label     string
		needField string // which single field is required for ignore
	}
	cases := []stageCase{
		{[]string{"3-adische", "System"}, "ignore-only", "ignore"},
		{[]string{"Microsoft", "Entra"}, "global-only", "global"},
		{[]string{"New", "York"}, "suggest-only", "suggest"},
	}
	for _, tc := range cases {
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...))
		}
		// Full pipeline ignores.
		requireAllContentIgnored(t, full.Disambiguate(fresh()), true, tc.label+" full")

		// Leave out the required stage → no ignore.
		switch tc.needField {
		case "ignore":
			d := &GermanRuleDisambiguator{MultitokenGlobal: g, MultitokenSuggest: sug, Rules: xml}
			requireAllContentIgnored(t, d.Disambiguate(fresh()), false, tc.label+" without ignore")
		case "global":
			d := &GermanRuleDisambiguator{MultitokenIgnore: ign, MultitokenSuggest: sug, Rules: xml}
			requireAllContentIgnored(t, d.Disambiguate(fresh()), false, tc.label+" without global")
		case "suggest":
			d := &GermanRuleDisambiguator{MultitokenIgnore: ign, MultitokenGlobal: g, Rules: xml}
			requireAllContentIgnored(t, d.Disambiguate(fresh()), false, tc.label+" without suggest")
		}

		// Only the required stage alone → ignore.
		switch tc.needField {
		case "ignore":
			requireAllContentIgnored(t, (&GermanRuleDisambiguator{MultitokenIgnore: ign}).Disambiguate(fresh()), true, tc.label+" only ignore")
		case "global":
			requireAllContentIgnored(t, (&GermanRuleDisambiguator{MultitokenGlobal: g}).Disambiguate(fresh()), true, tc.label+" only global")
		case "suggest":
			requireAllContentIgnored(t, (&GermanRuleDisambiguator{MultitokenSuggest: sug}).Disambiguate(fresh()), true, tc.label+" only suggest")
		}
	}

	// XML-only: only Rules stage ignores 2-Pfünder.
	sent := tokenSentence("2-Pfünder")
	requireNotIgnored(t, (&GermanRuleDisambiguator{
		MultitokenIgnore: ign, MultitokenGlobal: g, MultitokenSuggest: sug,
	}).Disambiguate(sent), "2-Pfünder")
	requireIgnored(t, (&GermanRuleDisambiguator{Rules: xml}).Disambiguate(sent), "2-Pfünder")
	requireIgnored(t, full.Disambiguate(sent), "2-Pfünder")
}
