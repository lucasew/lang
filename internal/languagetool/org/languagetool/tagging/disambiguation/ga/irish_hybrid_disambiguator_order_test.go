package ga

// Outcome twins for IrishHybridDisambiguator full stage order:
// Java IrishHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(input))
// i.e. MultiWordChunker.getInstance("/ga/multiwords.txt") defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling) FIRST, then
// XmlRuleDisambiguator(Irish.getInstance(), useGlobal=false).
//
// CRITICAL: multiword→XML (Romance order; same as Galician/Russian/ES;
// opposite of Polish/Swedish XML→multiword).
// Official ga.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/XML stage leaves).
//
// Note: official ga/disambiguation.xml has no multiword-flatten rules and no
// ignore_spelling actions; multiword open/close angle tags survive full hybrid.
// XML contributes immunize / add / replace on text-only official rules.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireGAHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverIrishMultiwords() == "" {
		t.Skip("official ga/multiwords.txt not discoverable")
	}
	if DiscoverIrishDisambiguationXML() == "" {
		t.Skip("official ga/disambiguation.xml not discoverable")
	}
}

func requireImmunized(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.True(t, tr.IsImmunized(), "%q must be immunized", s)
	}
}

func requireNotImmunized(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.False(t, tr.IsImmunized(), "%q must not be immunized", s)
	}
}

func requireNotIgnored(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.False(t, tr.IsIgnoredBySpeller(), "%q must not be ignore_spelling", s)
	}
}

// TestNewIrishHybridDisambiguator_WiresBothStages proves Java constructor
// eagerly builds multiwords Chunker and XmlRuleDisambiguator when the same
// official resources Java loads are present — with Irish flags.
func TestNewIrishHybridDisambiguator_WiresBothStages(t *testing.T) {
	requireGAHybridResources(t)

	mw := IrishMultiWordChunker()
	xml := IrishXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewIrishHybridDisambiguator()
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/ga/multiwords.txt) defaults F,F,F")
	require.NotNil(t, d.Rules,
		"disambiguator = new XmlRuleDisambiguator(Irish.getInstance()) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Irish multiwords defaults (no invent):
	// NO setRemovePreviousTags, NO setIgnoreSpelling
	// allowFirstCapitalized/allowAllUppercase/allowTitlecase false (outcome-tested)
	require.False(t, mw.RemovePreviousTags, "Irish multiwords does NOT setRemovePreviousTags")
	require.False(t, mw.AddIgnoreSpelling, "Irish multiwords does NOT setIgnoreSpelling")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official ga pack loads ~100+ expanded rules (NUM_DIG_ORD, RTE_PONC_IE, DE_SHIOR, …).
	require.GreaterOrEqual(t, len(xml.Rules), 100,
		"Irish XmlRuleDisambiguator must load official ga/disambiguation.xml rules")
	require.LessOrEqual(t, len(xml.Rules), 200,
		"Irish XmlRuleDisambiguator useGlobal=false should not append global pack")
}

// TestIrishHybridDisambiguator_OrderMultiwordThenXML proves stage isolation vs
// full Java order with Java-visible POS / immunize / add outcomes.
// Order is multiword first → XML second (NOT XML→multiword like Polish/Swedish).
func TestIrishHybridDisambiguator_OrderMultiwordThenXML(t *testing.T) {
	requireGAHybridResources(t)

	mw := IrishMultiWordChunker()
	xml := IrishXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyMulti := &IrishHybridDisambiguator{Chunker: mw}
	onlyXML := &IrishHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(sent))
	}
	// Reverse of Java (Polish-style XML→multiword) — used for leave-one-out / order contrast.
	reverseOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return mw.Disambiguate(xml.Disambiguate(sent))
	}
	full := NewIrishHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Multiword-only phrase: "ar ais" in ga/multiwords as Adv:Dir ---
	// Multiword alone → <Adv:Dir> / </Adv:Dir> (no removePreviousTags)
	// XML alone → no multiword POS
	// Full hybrid (multiword then XML) → multiword angles survive (no flatten rules)
	// Without multiword → no multiword POS
	{
		label := "ar ais"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais"))
		}

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "<Adv:Dir>"), "%s multiword-only ar: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</Adv:Dir>"), "%s multiword-only ais: %v", label, gotM[1])

		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<Adv:Dir>") || hasExactPOS(tags, "</Adv:Dir>") ||
				hasExactPOS(tags, "Adv:Dir") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "<Adv:Dir>"),
			"%s full hybrid ar (multiword then XML, angles survive): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "</Adv:Dir>"),
			"%s full hybrid ais: %v", label, gotFull[1])

		gotJO := contentPOSTags(javaOrder(fresh()))
		require.True(t, hasExactPOS(gotJO[0], "<Adv:Dir>"), "%s javaOrder ar: %v", label, gotJO[0])
		require.True(t, hasExactPOS(gotJO[1], "</Adv:Dir>"), "%s javaOrder ais: %v", label, gotJO[1])

		// Reverse still keeps multiword POS (XML does not wipe Adv:Dir on this surface).
		// Order proof for independent stages: composition + leave-one-out + call order.
		gotRev := contentPOSTags(reverseOrder(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<Adv:Dir>"),
			"%s reverse still has multiword open (stages independent on this surface): %v", label, gotRev[0])
		require.True(t, hasExactPOS(gotRev[1], "</Adv:Dir>"),
			"%s reverse multiword close: %v", label, gotRev[1])

		// Leave multiword out: XML-only hybrid has no multiword POS.
		noMulti := &IrishHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "Adv:Dir"),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		// Leave XML out: multiword close remains angle.
		noXML := &IrishHybridDisambiguator{Chunker: mw}
		gotNoX := contentPOSTags(noXML.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotNoX[0], "<Adv:Dir>"),
			"%s without XML keeps multiword open: %v", label, gotNoX[0])
		require.True(t, hasExactPOS(gotNoX[1], "</Adv:Dir>"),
			"%s without XML keeps multiword close: %v", label, gotNoX[1])

		// No setIgnoreSpelling on Irish multiwords.
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid token %q must not ignore spelling via multiwords", label, tr.GetToken())
		}
	}

	// Multiword 2-token Prep:Cmpd: "ar feadh" / "de bharr" / "chun go"
	{
		for _, tc := range []struct {
			parts []string
			open  string
			close string
			label string
		}{
			{[]string{"ar", "feadh"}, "<Prep:Cmpd>", "</Prep:Cmpd>", "ar feadh"},
			{[]string{"de", "bharr"}, "<Prep:Cmpd>", "</Prep:Cmpd>", "de bharr"},
			{[]string{"chun", "go"}, "<Conj:Subord>", "</Conj:Subord>", "chun go"},
			{[]string{"a", "lán"}, "<Subst:Noun:Sg>", "</Subst:Noun:Sg>", "a lán"},
		} {
			fresh := func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...))
			}
			gotFull := contentPOSTags(full.Disambiguate(fresh()))
			require.True(t, hasExactPOS(gotFull[0], tc.open), "%s full open: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotFull[len(gotFull)-1], tc.close),
				"%s full close: %v", tc.label, gotFull[len(gotFull)-1])
			gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
			require.True(t, hasExactPOS(gotM[0], tc.open), "%s multi-only: %v", tc.label, gotM[0])
			gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
			for i, tags := range gotX {
				require.False(t, hasAnyAnglePOS(tags),
					"%s xml-only token[%d]: %v", tc.label, i, tags)
			}
		}
	}

	// Multiword 3-token: "mar a déarfá" → Cmc open/close; interior empty on multi-only
	{
		label := "mar a déarfá"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("mar", "a", "déarfá"))
		}
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 3, label)
		require.True(t, hasExactPOS(gotM[0], "<Cmc>"), "%s multi open: %v", label, gotM[0])
		require.False(t, hasAnyAnglePOS(gotM[1]), "%s multi interior: %v", label, gotM[1])
		require.True(t, hasExactPOS(gotM[2], "</Cmc>"), "%s multi close: %v", label, gotM[2])

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 3, label)
		require.True(t, hasExactPOS(gotFull[0], "<Cmc>"), "%s full mar: %v", label, gotFull[0])
		require.False(t, hasAnyAnglePOS(gotFull[1]), "%s full interior a: %v", label, gotFull[1])
		require.True(t, hasExactPOS(gotFull[2], "</Cmc>"), "%s full déarfá: %v", label, gotFull[2])
	}

	// Multiword 4-token: "ar chor ar bith" → Adv:Gn
	{
		label := "ar chor ar bith"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "chor", "ar", "bith"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 4, label)
		require.True(t, hasExactPOS(gotFull[0], "<Adv:Gn>"), "%s full open: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[3], "</Adv:Gn>"), "%s full close: %v", label, gotFull[3])
		for i := 1; i < 3; i++ {
			require.False(t, hasAnyAnglePOS(gotFull[i]),
				"%s full interior token[%d]: %v", label, i, gotFull[i])
		}
		// XML alone (filterall AR_CHOR_AR_BITH needs POS from dict) → no invent angles
		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		for i, tags := range gotX {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "Adv:Gn"),
				"%s xml-only token[%d]: %v", label, i, tags)
		}
	}

	// Casing flags: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false
	// Official multiwords list lowercase "ar ais" only — capital forms denied under F,F,F.
	{
		gotAllCap := contentPOSTags(full.Disambiguate(
			languagetool.NewAnalyzedSentence(multiwordTokens("AR", "AIS"))))
		for i, tags := range gotAllCap {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "Adv:Dir"),
				"AR AIS all-caps denied full hybrid token[%d]: %v", i, tags)
		}

		gotFirstCap := contentPOSTags(full.Disambiguate(
			languagetool.NewAnalyzedSentence(multiwordTokens("Ar", "ais"))))
		for i, tags := range gotFirstCap {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "Adv:Dir"),
				"Ar ais first-cap denied full hybrid token[%d]: %v", i, tags)
		}

		gotTitle := contentPOSTags(full.Disambiguate(
			languagetool.NewAnalyzedSentence(multiwordTokens("Ar", "Ais"))))
		for i, tags := range gotTitle {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "Adv:Dir"),
				"Ar Ais titlecase denied full hybrid token[%d]: %v", i, tags)
		}

		// Listed lowercase still matches.
		gotListed := contentPOSTags(full.Disambiguate(
			languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais"))))
		require.True(t, hasExactPOS(gotListed[0], "<Adv:Dir>"), "ar ais listed: %v", gotListed[0])
	}

	// --- (2) XML-only effects: add / replace / immunize (token-built; no ga.dict) ---
	// Chunker does not invent Num:Dig:Ord / immunize / replace; XML stage does.
	{
		// NUM_DIG_ORD: 6ú → Num:Dig:Ord
		sent6 := tokenSentence("6ú")
		gotM := contentPOSTags(onlyMulti.Disambiguate(sent6))
		require.False(t, hasExactPOS(gotM[0], "Num:Dig:Ord"),
			"6ú multiword-only must not invent Num:Dig:Ord: %v", gotM[0])
		gotX := contentPOSTags(onlyXML.Disambiguate(tokenSentence("6ú")))
		require.True(t, hasExactPOS(gotX[0], "Num:Dig:Ord"), "6ú xml-only: %v", gotX[0])
		gotFull := contentPOSTags(full.Disambiguate(tokenSentence("6ú")))
		require.True(t, hasExactPOS(gotFull[0], "Num:Dig:Ord"), "6ú full hybrid: %v", gotFull[0])
		gotJO := contentPOSTags(javaOrder(tokenSentence("6ú")))
		require.True(t, hasExactPOS(gotJO[0], "Num:Dig:Ord"), "6ú javaOrder: %v", gotJO[0])
		// leave-one-out: without Rules, multiword alone must not invent
		noXML := &IrishHybridDisambiguator{Chunker: mw}
		gotNoX := contentPOSTags(noXML.Disambiguate(tokenSentence("6ú")))
		require.False(t, hasExactPOS(gotNoX[0], "Num:Dig:Ord"), "6ú without XML: %v", gotNoX[0])

		// NUM_DIG_ORD_OBS: 6adh
		got6adh := contentPOSTags(full.Disambiguate(tokenSentence("6adh")))
		require.True(t, hasExactPOS(got6adh[0], "Num:Dig:Ord"), "6adh full: %v", got6adh[0])

		// DE_SHIOR: de + shíor → replace shíor with Subst:Noun:Sg:Len
		sentDS := tokenSentence("de", "shíor")
		outM := onlyMulti.Disambiguate(sentDS)
		// multiword alone: not a multiword phrase; no invent replace
		require.False(t, hasExactPOS(contentPOSTags(outM)[1], "Subst:Noun:Sg:Len"),
			"de shíor multi-only must not invent Subst:Noun:Sg:Len: %v", contentPOSTags(outM)[1])
		outX := onlyXML.Disambiguate(tokenSentence("de", "shíor"))
		require.Contains(t, posTagsOn(tokenBySurface(outX, "shíor")), "Subst:Noun:Sg:Len")
		outFull := full.Disambiguate(tokenSentence("de", "shíor"))
		require.Contains(t, posTagsOn(tokenBySurface(outFull, "shíor")), "Subst:Noun:Sg:Len")
		outJO := javaOrder(tokenSentence("de", "shíor"))
		require.Contains(t, posTagsOn(tokenBySurface(outJO, "shíor")), "Subst:Noun:Sg:Len")
		require.False(t, hasExactPOS(contentPOSTags(noXML.Disambiguate(tokenSentence("de", "shíor")))[1], "Subst:Noun:Sg:Len"),
			"de shíor without XML must not invent replace")

		// RTE_PONC_IE: rte.ie immunize (spacebefore=no) — XML only
		freshRTE := func() *languagetool.AnalyzedSentence {
			return tokenSentenceNoSpace("rte", ".", "ie")
		}
		requireNotImmunized(t, onlyMulti.Disambiguate(freshRTE()), "rte", "ie")
		requireImmunized(t, onlyXML.Disambiguate(freshRTE()), "rte", "ie")
		requireImmunized(t, full.Disambiguate(freshRTE()), "rte", "ie")
		requireImmunized(t, javaOrder(freshRTE()), "rte", "ie")
		requireNotImmunized(t, noXML.Disambiguate(freshRTE()), "rte", "ie")

		// Word-repeat immunize: ciaróg ciaróg (CIAROG_CIAROG) — XML only
		sentCiar := tokenSentence("ciaróg", "ciaróg")
		requireNotImmunized(t, onlyMulti.Disambiguate(sentCiar), "ciaróg")
		requireImmunized(t, onlyXML.Disambiguate(tokenSentence("ciaróg", "ciaróg")), "ciaróg")
		requireImmunized(t, full.Disambiguate(tokenSentence("ciaróg", "ciaróg")), "ciaróg")
		requireNotImmunized(t, noXML.Disambiguate(tokenSentence("ciaróg", "ciaróg")), "ciaróg")

		// Multiword stage does not set ignore_spelling (no setIgnoreSpelling; ga XML has none).
		requireNotIgnored(t, full.Disambiguate(tokenSentence("ar", "ais")), "ar", "ais")
		requireNotIgnored(t, full.Disambiguate(tokenSentence("6ú")), "6ú")
	}

	// Dual-stage composition: multiword phrase + XML add on separate tokens.
	// "ar ais 6ú" — multiword tags ar/ais; XML adds Num:Dig:Ord on 6ú.
	// full == XML(multiword(input)); leave-one-out isolates each stage.
	{
		label := "ar ais 6ú dual-stage"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais", "6ú"))
		}

		outM := onlyMulti.Disambiguate(fresh())
		gotM := contentPOSTags(outM)
		require.True(t, hasExactPOS(gotM[0], "<Adv:Dir>"), "%s multi open: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</Adv:Dir>"), "%s multi close: %v", label, gotM[1])
		require.False(t, hasExactPOS(gotM[2], "Num:Dig:Ord"),
			"%s multi-only 6ú must not invent Num:Dig:Ord: %v", label, gotM[2])

		outX := onlyXML.Disambiguate(fresh())
		gotX := contentPOSTags(outX)
		for i := 0; i < 2; i++ {
			require.False(t, hasAnyAnglePOS(gotX[i]),
				"%s xml-only ar/ais no multiword angles: %v", label, gotX[i])
		}
		require.True(t, hasExactPOS(gotX[2], "Num:Dig:Ord"), "%s xml-only 6ú: %v", label, gotX[2])

		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		require.True(t, hasExactPOS(gotFull[0], "<Adv:Dir>"), "%s full ar: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "</Adv:Dir>"), "%s full ais: %v", label, gotFull[1])
		require.True(t, hasExactPOS(gotFull[2], "Num:Dig:Ord"), "%s full 6ú: %v", label, gotFull[2])

		outJO := javaOrder(fresh())
		gotJO := contentPOSTags(outJO)
		for i := range gotFull {
			require.ElementsMatch(t, gotFull[i], gotJO[i],
				"%s POS parity full vs javaOrder token[%d]", label, i)
		}

		// Reverse: stages independent on these surfaces → same POS membership.
		gotRev := contentPOSTags(reverseOrder(fresh()))
		for i := range gotFull {
			require.ElementsMatch(t, gotFull[i], gotRev[i],
				"%s reverse POS (independent stages) token[%d]", label, i)
		}
	}
}

// TestIrishHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS proves
// multiword first (before XML) attaches official multiword POS and XML still fires after
// (immunize / add / replace).
func TestIrishHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS(t *testing.T) {
	requireGAHybridResources(t)
	full := NewIrishHybridDisambiguator()

	// Multiword angles survive full order (no ga multiword-flatten rules).
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<Adv:Dir>"), "ar after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</Adv:Dir>"), "ais after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("a", "lán")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<Subst:Noun:Sg>"), "a after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</Subst:Noun:Sg>"), "lán after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("chun", "go")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<Conj:Subord>"), "chun after full: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</Conj:Subord>"), "go after full: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("mar", "a", "déarfá")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<Cmc>"), "mar after full: %v", got[0])
	require.True(t, hasExactPOS(got[2], "</Cmc>"), "déarfá after full: %v", got[2])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("ar", "chor", "ar", "bith")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<Adv:Gn>"), "ar after full: %v", got[0])
	require.True(t, hasExactPOS(got[3], "</Adv:Gn>"), "bith after full: %v", got[3])

	// XML effects still fire after multiword (multiword no-op on these surfaces).
	got6 := contentPOSTags(full.Disambiguate(tokenSentence("6ú")))
	require.True(t, hasExactPOS(got6[0], "Num:Dig:Ord"), "6ú after full order: %v", got6[0])
	requireImmunized(t, full.Disambiguate(tokenSentenceNoSpace("rte", ".", "ie")), "rte", "ie")
	require.Contains(t, posTagsOn(tokenBySurface(
		full.Disambiguate(tokenSentence("de", "shíor")), "shíor")), "Subst:Noun:Sg:Len")
}

// TestIrishHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == XML(multiword(input)) for official isolation surfaces.
func TestIrishHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireGAHybridResources(t)
	mw := IrishMultiWordChunker()
	xml := IrishXmlRuleDisambiguator()
	full := NewIrishHybridDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantPOS0 string // empty = no multiword POS required
		wantPOS1 string
		wantPOS  map[string]string // surface → required POS
		wantImm  map[string]bool   // surface → immunize
		noAngles bool              // require no multiword angle POS
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais")) },
			label:    "ar ais",
			wantPOS0: "<Adv:Dir>",
			wantPOS1: "</Adv:Dir>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("a", "lán")) },
			label:    "a lán",
			wantPOS0: "<Subst:Noun:Sg>",
			wantPOS1: "</Subst:Noun:Sg>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "feadh")) },
			label:    "ar feadh",
			wantPOS0: "<Prep:Cmpd>",
			wantPOS1: "</Prep:Cmpd>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("chun", "go")) },
			label:    "chun go",
			wantPOS0: "<Conj:Subord>",
			wantPOS1: "</Conj:Subord>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("de", "bharr")) },
			label:    "de bharr",
			wantPOS0: "<Prep:Cmpd>",
			wantPOS1: "</Prep:Cmpd>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("mar", "a", "déarfá")) },
			label:    "mar a déarfá",
			wantPOS0: "<Cmc>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "chor", "ar", "bith")) },
			label:    "ar chor ar bith",
			wantPOS0: "<Adv:Gn>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("AR", "AIS")) },
			label:    "AR AIS all-caps denied",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Ar", "ais")) },
			label:    "Ar ais first-cap denied",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Ar", "Ais")) },
			label:    "Ar Ais titlecase denied",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Zxqwv", "Plmnb")) },
			label:    "random non-listed",
			noAngles: true,
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("6ú") },
			label:   "6ú",
			wantPOS: map[string]string{"6ú": "Num:Dig:Ord"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("6adh") },
			label:   "6adh",
			wantPOS: map[string]string{"6adh": "Num:Dig:Ord"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("de", "shíor") },
			label:   "de shíor",
			wantPOS: map[string]string{"shíor": "Subst:Noun:Sg:Len"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentenceNoSpace("rte", ".", "ie") },
			label:   "rte.ie",
			wantImm: map[string]bool{"rte": true, "ie": true},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("ciaróg", "ciaróg") },
			label:   "ciaróg ciaróg",
			wantImm: map[string]bool{"ciaróg": true},
		},
		{
			// dual-stage: multiword + XML add
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais", "6ú")) },
			label:    "ar ais 6ú",
			wantPOS0: "<Adv:Dir>",
			wantPOS1: "</Adv:Dir>",
			wantPOS:  map[string]string{"6ú": "Num:Dig:Ord"},
		},
	}
	for _, tc := range cases {
		outFull := full.Disambiguate(tc.fresh())
		// Java: disambiguator.disambiguate(chunker.disambiguate(input))
		outManual := xml.Disambiguate(mw.Disambiguate(tc.fresh()))

		gotFull := contentPOSTags(outFull)
		gotManual := contentPOSTags(outManual)
		require.Equal(t, len(gotFull), len(gotManual), tc.label+" content POS count")
		if tc.wantPOS0 != "" {
			require.True(t, hasExactPOS(gotFull[0], tc.wantPOS0), "%s full POS0: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotManual[0], tc.wantPOS0), "%s manual POS0: %v", tc.label, gotManual[0])
			if tc.wantPOS1 != "" {
				require.True(t, hasExactPOS(gotFull[1], tc.wantPOS1), "%s full POS1: %v", tc.label, gotFull[1])
				require.True(t, hasExactPOS(gotManual[1], tc.wantPOS1), "%s manual POS1: %v", tc.label, gotManual[1])
			}
		} else if tc.noAngles || (len(tc.wantPOS) == 0 && len(tc.wantImm) == 0) {
			for i, tags := range gotFull {
				require.False(t, hasAnyAnglePOS(tags),
					"%s full token[%d] must have no multiword angle POS: %v", tc.label, i, tags)
			}
		}
		// Per-token POS membership parity.
		for i := range gotFull {
			require.ElementsMatch(t, gotFull[i], gotManual[i],
				"%s token[%d] POS parity full vs javaOrder", tc.label, i)
		}
		// Ignore / immunize parity.
		ft, mt := outFull.GetTokens(), outManual.GetTokens()
		require.Equal(t, len(ft), len(mt), tc.label+" token count")
		for i := range ft {
			if i == 0 || ft[i].IsWhitespace() {
				continue
			}
			require.Equal(t, ft[i].IsIgnoredBySpeller(), mt[i].IsIgnoredBySpeller(),
				"%s token[%d]=%q ignore parity full vs javaOrder", tc.label, i, ft[i].GetToken())
			require.Equal(t, ft[i].IsImmunized(), mt[i].IsImmunized(),
				"%s token[%d]=%q immunize parity full vs javaOrder", tc.label, i, ft[i].GetToken())
		}
		for surface, want := range tc.wantImm {
			trF := tokenBySurface(outFull, surface)
			trM := tokenBySurface(outManual, surface)
			require.NotNil(t, trF, "%s full missing %q", tc.label, surface)
			require.NotNil(t, trM, "%s manual missing %q", tc.label, surface)
			require.Equal(t, want, trF.IsImmunized(), "%s full %q immunize", tc.label, surface)
			require.Equal(t, want, trM.IsImmunized(), "%s manual %q immunize", tc.label, surface)
		}
		for surface, wantPOS := range tc.wantPOS {
			trF := tokenBySurface(outFull, surface)
			trM := tokenBySurface(outManual, surface)
			require.NotNil(t, trF, "%s full missing %q", tc.label, surface)
			require.NotNil(t, trM, "%s manual missing %q", tc.label, surface)
			var tagsF, tagsM []string
			for _, r := range trF.GetReadings() {
				if r != nil && r.GetPOSTag() != nil {
					tagsF = append(tagsF, *r.GetPOSTag())
				}
			}
			for _, r := range trM.GetReadings() {
				if r != nil && r.GetPOSTag() != nil {
					tagsM = append(tagsM, *r.GetPOSTag())
				}
			}
			require.True(t, hasExactPOS(tagsF, wantPOS), "%s full %q want %s in %v", tc.label, surface, wantPOS, tagsF)
			require.True(t, hasExactPOS(tagsM, wantPOS), "%s manual %q want %s in %v", tc.label, surface, wantPOS, tagsM)
		}
	}
}

// TestIrishHybridDisambiguator_StageOrderIsMultiwordThenXML proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestIrishHybridDisambiguator_StageOrderIsMultiwordThenXML(t *testing.T) {
	requireGAHybridResources(t)
	mw := IrishMultiWordChunker()
	xml := IrishXmlRuleDisambiguator()
	full := NewIrishHybridDisambiguator()

	// Multiword surface: only Chunker produces multiword POS; full order keeps angles.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("ar", "ais"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "<Adv:Dir>"), "full: %v", gotFull[0])

		// Without Chunker → no multiword POS
		noChunk := &IrishHybridDisambiguator{Rules: xml}
		gotNoC := contentPOSTags(noChunk.Disambiguate(fresh()))
		for i, tags := range gotNoC {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "Adv:Dir"),
				"without chunker token[%d]: %v", i, tags)
		}
		// Only Chunker → multiword angles
		onlyC := &IrishHybridDisambiguator{Chunker: mw}
		gotOnlyC := contentPOSTags(onlyC.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyC[0], "<Adv:Dir>"), "only chunker: %v", gotOnlyC[0])
		require.True(t, hasExactPOS(gotOnlyC[1], "</Adv:Dir>"), "only chunker close: %v", gotOnlyC[1])
	}

	// XML-only surface: only Rules sets immunize (RTE_PONC_IE).
	{
		freshRTE := func() *languagetool.AnalyzedSentence {
			return tokenSentenceNoSpace("rte", ".", "ie")
		}
		requireImmunized(t, full.Disambiguate(freshRTE()), "rte", "ie")
		requireNotImmunized(t, (&IrishHybridDisambiguator{Chunker: mw}).Disambiguate(freshRTE()), "rte", "ie")
		requireImmunized(t, (&IrishHybridDisambiguator{Rules: xml}).Disambiguate(freshRTE()), "rte", "ie")
	}

	// XML NUM_DIG_ORD: only Rules invents Num:Dig:Ord.
	{
		gotFull := contentPOSTags(full.Disambiguate(tokenSentence("6ú")))
		require.True(t, hasExactPOS(gotFull[0], "Num:Dig:Ord"), "full 6ú: %v", gotFull[0])
		gotNoXML := contentPOSTags((&IrishHybridDisambiguator{Chunker: mw}).Disambiguate(tokenSentence("6ú")))
		require.False(t, hasExactPOS(gotNoXML[0], "Num:Dig:Ord"), "chunker-only 6ú: %v", gotNoXML[0])
		gotOnlyXML := contentPOSTags((&IrishHybridDisambiguator{Rules: xml}).Disambiguate(tokenSentence("6ú")))
		require.True(t, hasExactPOS(gotOnlyXML[0], "Num:Dig:Ord"), "xml-only 6ú: %v", gotOnlyXML[0])
	}

	// Call-order: Chunker then Rules (Java nested call: outer=disambiguator, inner=chunker).
	{
		var order []string
		rulesStub := &gaOrderStage{name: "rules", order: &order}
		chunkStub := &gaOrderStage{name: "chunker", order: &order}
		d := &IrishHybridDisambiguator{Rules: rulesStub, Chunker: chunkStub}
		d.Disambiguate(tokenSentence("x"))
		require.Equal(t, []string{"chunker", "rules"}, order,
			"Java: disambiguator.disambiguate(chunker.disambiguate(input)) → multiword then XML")
	}
}

// gaOrderStage records Disambiguate call order for stage-order proof.
type gaOrderStage struct {
	name  string
	order *[]string
}

func (s *gaOrderStage) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s.order != nil {
		*s.order = append(*s.order, s.name)
	}
	return input
}

var _ interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
} = (*gaOrderStage)(nil)
