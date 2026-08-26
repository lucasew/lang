package uk

// Outcome twins for UkrainianHybridDisambiguator full stage order:
// Java UkrainianHybridDisambiguator.disambiguate:
//   preDisambiguate(input);  // SimpleDisambiguator + UK-specific retags
//   return disambiguator.disambiguate(chunker.disambiguate(input));
// i.e. PRE first, then UkrainianMultiwordChunker("/uk/multiwords.txt", allowFirstCapitalized=true),
// then XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT) with useGlobalDisambiguation=false.
//
// CRITICAL: pre → multiword → XML (UK-specific pre stage; multiword before XML like RU/GL/GA;
// opposite of Polish/Swedish XML→multiword).
// Official uk.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/XML stage leaves + Soft hybrid pre surfaces).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/stretchr/testify/require"
)

func requireUKHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverUkrainianMultiwords() == "" {
		t.Skip("official uk/multiwords.txt not discoverable")
	}
	if DiscoverUkrainianDisambiguationXML() == "" {
		t.Skip("official uk/disambiguation.xml not discoverable")
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

// ukJavaOrderComposition ports Java:
//
//	preDisambiguate(input);
//	return disambiguator.disambiguate(chunker.disambiguate(input));
//
// preOnly hybrid has Simple + UK retags, nil Chunker/Inner.
func ukJavaOrderComposition(
	preOnly *UkrainianHybridDisambiguator,
	mw, xml interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	},
	sent *languagetool.AnalyzedSentence,
) *languagetool.AnalyzedSentence {
	preOut := preOnly.Disambiguate(sent)
	return xml.Disambiguate(mw.Disambiguate(preOut))
}

// TestNewUkrainianHybridDisambiguator_WiresAllStages proves Java constructor
// eagerly builds SimpleDisambiguator, UkrainianMultiwordChunker (allowFirstCapitalized=true),
// and XmlRuleDisambiguator(useGlobal=false) when official resources are present.
func TestNewUkrainianHybridDisambiguator_WiresAllStages(t *testing.T) {
	requireUKHybridResources(t)

	mw := UkrainianMultiwordChunkerDefault()
	xml := UkrainianXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewUkrainianHybridDisambiguator()
	require.NotNil(t, d.Simple,
		"simpleDisambiguator = new SimpleDisambiguator()")
	require.NotNil(t, d.Chunker,
		"chunker = new UkrainianMultiwordChunker(/uk/multiwords.txt, true)")
	require.NotNil(t, d.Inner,
		"disambiguator = new XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Inner)

	// Ukrainian multiword flags (Java MultiWordChunker2(filename, allowFirstCapitalized=true)):
	// allowFirstCapitalized=true; WrapTag default true; NO setRemoveOtherReadings; NO setIgnoreSpelling.
	require.True(t, mw.AllowFirstCapitalized,
		"UkrainianMultiwordChunker(..., true) allowFirstCapitalized")
	require.True(t, mw.WrapTag, "MultiWordChunker2 default WrapTag=true")
	require.False(t, mw.RemoveOtherReadings, "UK multiwords does NOT setRemoveOtherReadings")
	require.False(t, mw.AddIgnoreSpelling, "UK multiwords does NOT setIgnoreSpelling")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official UK pack expands to ~400+ rules (see 3.A.4 ACCEPTed leaf).
	require.GreaterOrEqual(t, len(xml.Rules), 400,
		"Ukrainian XmlRuleDisambiguator must load official uk/disambiguation.xml rules")
	for _, r := range xml.Rules {
		require.NotNil(t, r)
		require.NotContains(t, r.GetID(), "GLOBAL_",
			"useGlobal=false must not load disambiguation-global.xml")
	}
}

// TestUkrainianHybridDisambiguator_OrderPreThenMultiwordThenXML proves stage isolation vs
// full Java order with Java-visible POS / immunize outcomes.
// Order is pre first → multiword second → XML third.
func TestUkrainianHybridDisambiguator_OrderPreThenMultiwordThenXML(t *testing.T) {
	requireUKHybridResources(t)

	mw := UkrainianMultiwordChunkerDefault()
	xml := UkrainianXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	// Isolation hybrids: UK retags always run in Disambiguate; omit Chunker/Inner/Simple as needed.
	// Pure multiword / pure XML isolation uses the stage directly (no hybrid pre).
	onlyMulti := mw
	onlyXML := xml
	preOnly := &UkrainianHybridDisambiguator{Simple: NewSimpleDisambiguator()}
	// Hybrid without multiword (pre + XML)
	noMulti := &UkrainianHybridDisambiguator{Simple: NewSimpleDisambiguator(), Inner: xml}
	// Hybrid without XML (pre + multiword)
	noXML := &UkrainianHybridDisambiguator{Simple: NewSimpleDisambiguator(), Chunker: mw}
	// Hybrid without Simple (retags still run; rare-form maps off)
	noSimple := &UkrainianHybridDisambiguator{Chunker: mw, Inner: xml}

	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return ukJavaOrderComposition(preOnly, mw, xml, sent)
	}
	// Reverse of Java multiword/XML (XML then multiword) after pre — order contrast.
	reverseMWXML := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		preOut := preOnly.Disambiguate(sent)
		return mw.Disambiguate(xml.Disambiguate(preOut))
	}

	full := NewUkrainianHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Inner)
	require.NotNil(t, full.Simple)

	// --- (1) Multiword phrases: MultiWordChunker2 WrapTag → <tag> on every content token ---
	// Multiword alone → <adv>/<insert>
	// XML alone → no multiword wrap
	// Full hybrid (pre→mw→XML) → multiword wrap survives
	// Without multiword → no wrap
	{
		for _, tc := range []struct {
			parts []string
			tag   string
			label string
		}{
			{[]string{"для", "годиться"}, "<adv>", "для годиться"},
			{[]string{"а", "капела"}, "<adv>", "а капела"},
			{[]string{"на", "жаль"}, "<insert>", "на жаль"},
			{[]string{"як", "правило"}, "<insert>", "як правило"},
			{[]string{"від", "і", "до"}, "<adv>", "від і до"},
			// allowFirstCapitalized=true: first-cap of lowercase official entry
			{[]string{"Для", "годиться"}, "<adv>", "Для годиться first-cap"},
			{[]string{"На", "жаль"}, "<insert>", "На жаль first-cap"},
		} {
			fresh := func() *languagetool.AnalyzedSentence { return mwSent(tc.parts...) }

			gotM := contentPOSTagsMW(onlyMulti.Disambiguate(fresh()))
			require.Len(t, gotM, len(tc.parts), tc.label+" multi token count")
			for i := range tc.parts {
				require.True(t, hasExactPOS(gotM[i], tc.tag),
					"%s multi-only token[%d] want %s in %v", tc.label, i, tc.tag, gotM[i])
			}

			gotX := contentPOSTagsMW(onlyXML.Disambiguate(fresh()))
			for i, tags := range gotX {
				require.False(t, hasExactPOS(tags, tc.tag) || hasAnyAnglePOS(tags),
					"%s xml-only token[%d] must have no multiword wrap, got %v", tc.label, i, tags)
			}

			gotFull := contentPOSTagsMW(full.Disambiguate(fresh()))
			require.Len(t, gotFull, len(tc.parts), tc.label+" full token count")
			for i := range tc.parts {
				require.True(t, hasExactPOS(gotFull[i], tc.tag),
					"%s full hybrid token[%d] want %s in %v", tc.label, i, tc.tag, gotFull[i])
			}

			gotJO := contentPOSTagsMW(javaOrder(fresh()))
			for i := range tc.parts {
				require.True(t, hasExactPOS(gotJO[i], tc.tag),
					"%s javaOrder token[%d] want %s in %v", tc.label, i, tc.tag, gotJO[i])
			}

			// Leave multiword out: pre+XML has no multiword wrap.
			gotNoM := contentPOSTagsMW(noMulti.Disambiguate(fresh()))
			for i, tags := range gotNoM {
				require.False(t, hasExactPOS(tags, tc.tag) || hasAnyAnglePOS(tags),
					"%s without multiword token[%d]: %v", tc.label, i, tags)
			}

			// Leave XML out: multiword wrap remains.
			gotNoX := contentPOSTagsMW(noXML.Disambiguate(fresh()))
			for i := range tc.parts {
				require.True(t, hasExactPOS(gotNoX[i], tc.tag),
					"%s without XML token[%d]: %v", tc.label, i, gotNoX[i])
			}

			// No setIgnoreSpelling on Ukrainian multiwords.
			for i, tr := range full.Disambiguate(fresh()).GetTokens() {
				if i == 0 || tr.IsWhitespace() {
					continue
				}
				require.False(t, tr.IsIgnoredBySpeller(),
					"%s full hybrid token %q must not ignore spelling via multiwords",
					tc.label, tr.GetToken())
			}
		}
	}

	// Casing flags: allowFirstCapitalized=true only first-caps lower entries;
	// all-upper / titlecase second token denied (MultiWordChunker2).
	{
		gotAllCap := contentPOSTagsMW(full.Disambiguate(mwSent("НА", "ЖАЛЬ")))
		for i, tags := range gotAllCap {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "<insert>"),
				"НА ЖАЛЬ all-caps denied full hybrid token[%d]: %v", i, tags)
		}
		gotTitle := contentPOSTagsMW(full.Disambiguate(mwSent("На", "Жаль")))
		for i, tags := range gotTitle {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "<insert>"),
				"На Жаль titlecase denied full hybrid token[%d]: %v", i, tags)
		}
		// Listed lowercase still matches.
		gotListed := contentPOSTagsMW(full.Disambiguate(mwSent("на", "жаль")))
		require.True(t, hasExactPOS(gotListed[0], "<insert>"), "на жаль listed: %v", gotListed[0])
		require.True(t, hasExactPOS(gotListed[1], "<insert>"), "на жаль listed: %v", gotListed[1])
	}

	// Reverse multiword/XML after pre: stages largely independent on these surfaces
	// (no UK multiword-flatten rules). Order proof is composition + leave-one-out + call order.
	{
		label := "для годиться reverse"
		fresh := func() *languagetool.AnalyzedSentence { return mwSent("для", "годиться") }
		gotRev := contentPOSTagsMW(reverseMWXML(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<adv>"),
			"%s reverse still has multiword wrap: %v", label, gotRev[0])
		require.True(t, hasExactPOS(gotRev[1], "<adv>"),
			"%s reverse close: %v", label, gotRev[1])
	}

	// --- (2) XML-only effects (token-built; no uk.dict) ---
	// Chunker does not invent immunize / noninfl:bad / number POS; XML stage does.
	{
		// freq_infix: -ськ- → immunize
		freshInfx := func() *languagetool.AnalyzedSentence { return tokenSentence("-ськ-") }
		requireNotImmunized(t, onlyMulti.Disambiguate(freshInfx()), "-ськ-")
		requireImmunized(t, onlyXML.Disambiguate(freshInfx()), "-ськ-")
		requireImmunized(t, full.Disambiguate(freshInfx()), "-ськ-")
		requireImmunized(t, javaOrder(freshInfx()), "-ськ-")
		requireNotImmunized(t, noXML.Disambiguate(freshInfx()), "-ськ-")
		requireImmunized(t, noMulti.Disambiguate(freshInfx()), "-ськ-")

		// td_bad: і + тд → noninfl:bad
		freshTD := func() *languagetool.AnalyzedSentence { return tokenSentence("і", "тд") }
		gotM := contentPOSTagsMW(onlyMulti.Disambiguate(freshTD()))
		require.False(t, hasExactPOS(gotM[1], "noninfl:bad"),
			"і тд multi-only must not invent noninfl:bad: %v", gotM[1])
		gotX := contentPOSTagsMW(onlyXML.Disambiguate(freshTD()))
		require.True(t, hasExactPOS(gotX[1], "noninfl:bad"), "і тд xml-only: %v", gotX[1])
		gotFull := contentPOSTagsMW(full.Disambiguate(freshTD()))
		require.True(t, hasExactPOS(gotFull[1], "noninfl:bad"), "і тд full: %v", gotFull[1])
		gotJO := contentPOSTagsMW(javaOrder(freshTD()))
		require.True(t, hasExactPOS(gotJO[1], "noninfl:bad"), "і тд javaOrder: %v", gotJO[1])
		gotNoX := contentPOSTagsMW(noXML.Disambiguate(freshTD()))
		require.False(t, hasExactPOS(gotNoX[1], "noninfl:bad"), "і тд without XML: %v", gotNoX[1])

		// sviatogo_yura: святого + Юра → noun:anim:m:v_rod:prop:fname
		freshSY := func() *languagetool.AnalyzedSentence { return tokenSentence("святого", "Юра") }
		require.False(t, hasExactPOS(contentPOSTagsMW(onlyMulti.Disambiguate(freshSY()))[1],
			"noun:anim:m:v_rod:prop:fname"), "святого Юра multi-only invent")
		require.True(t, hasExactPOS(contentPOSTagsMW(onlyXML.Disambiguate(freshSY()))[1],
			"noun:anim:m:v_rod:prop:fname"), "святого Юра xml-only")
		require.True(t, hasExactPOS(contentPOSTagsMW(full.Disambiguate(freshSY()))[1],
			"noun:anim:m:v_rod:prop:fname"), "святого Юра full")
		require.True(t, hasExactPOS(contentPOSTagsMW(javaOrder(freshSY()))[1],
			"noun:anim:m:v_rod:prop:fname"), "святого Юра javaOrder")

		// c-r_1 / c-r_2: ц. р.
		freshCR := func() *languagetool.AnalyzedSentence { return tokenSentence("ц.", "р.") }
		outFull := full.Disambiguate(freshCR())
		require.Contains(t, posTagsOn(tokenBySurface(outFull, "ц.")), "adj:m:v_rod:pron:dem")
		require.Contains(t, posTagsOn(tokenBySurface(outFull, "р.")), "noun:inanim:m:v_rod")
		outM := onlyMulti.Disambiguate(freshCR())
		require.NotContains(t, posTagsOn(tokenBySurface(outM, "ц.")), "adj:m:v_rod:pron:dem")

		// POINT_NUMBER: номер + 17-а → noninfl
		freshPN := func() *languagetool.AnalyzedSentence { return tokenSentence("номер", "17-а") }
		require.Contains(t, posTagsOn(tokenBySurface(full.Disambiguate(freshPN()), "17-а")), "noninfl")
		require.NotContains(t, posTagsOn(tokenBySurface(onlyMulti.Disambiguate(freshPN()), "17-а")), "noninfl")
		require.NotContains(t, posTagsOn(tokenBySurface(noXML.Disambiguate(freshPN()), "17-а")), "noninfl")

		// DIS_KOT_D_IVUAR: Кот + д'Івуара
		freshKot := func() *languagetool.AnalyzedSentence { return tokenSentence("Кот", "д'Івуара") }
		require.Contains(t, posTagsOn(tokenBySurface(full.Disambiguate(freshKot()), "Кот")),
			"noninfl:foreign:prop:geo:bad")
		require.NotContains(t, posTagsOn(tokenBySurface(onlyMulti.Disambiguate(freshKot()), "Кот")),
			"noninfl:foreign:prop:geo:bad")
	}

	// --- (3) Pre-only effect: removeVmis (Java preDisambiguate) ---
	// Multiword alone / XML alone do not strip v_mis; full hybrid pre does.
	{
		freshVM := func() *languagetool.AnalyzedSentence {
			p, l1 := "noun:inanim:m:v_mis", "зв'язок"
			p2, l2 := "noun:inanim:m:v_rod", "зв'язок"
			atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
				languagetool.NewAnalyzedToken("Зв'язку", &p, &l1),
				languagetool.NewAnalyzedToken("Зв'язку", &p2, &l2),
			}, 0)
			start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
				languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil),
			}, 0)
			return languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, atr})
		}

		// multi-only: keeps v_mis
		outM := onlyMulti.Disambiguate(freshVM())
		require.True(t, outM.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"),
			"multi-only must not run removeVmis")
		// xml-only: keeps v_mis (no invent strip)
		outX := onlyXML.Disambiguate(freshVM())
		require.True(t, outX.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"),
			"xml-only must not run removeVmis")
		// pre-only hybrid: strips v_mis
		outP := preOnly.Disambiguate(freshVM())
		require.False(t, outP.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"),
			"pre-only removeVmis")
		require.True(t, outP.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_rod"))
		// full hybrid: strips v_mis
		outF := full.Disambiguate(freshVM())
		require.False(t, outF.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"),
			"full hybrid pre removeVmis")
		require.True(t, outF.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_rod"))
		// javaOrder composition
		outJO := javaOrder(freshVM())
		require.False(t, outJO.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"))
		// noSimple still runs retags (RemoveVmis is not SimpleDisambiguator)
		outNS := noSimple.Disambiguate(freshVM())
		require.False(t, outNS.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"),
			"retags run even when Simple is nil")
	}

	// Dual-stage: multiword + XML on separate tokens in one sentence.
	// "для годиться -ськ-" is awkward; use sequential surfaces "для" "годиться" and separate
	// immunize check already done. Dual POS: multiword wrap + later XML on different sentence ok.
	// Combined: "для годиться" multiword wrap still present after full; immunize on separate call.
	{
		label := "для годиться + immunize dual surfaces"
		got := contentPOSTagsMW(full.Disambiguate(mwSent("для", "годиться")))
		require.True(t, hasExactPOS(got[0], "<adv>"), "%s multi after full: %v", label, got[0])
		requireImmunized(t, full.Disambiguate(tokenSentence("-ськ-")), "-ськ-")
	}

	// Dual-stage single sentence via tokenSentence (spaces): multiword still matches.
	{
		label := "tokenSentence multiword + XML later token"
		// "для годиться" as spaced tokens still chunked by MultiWordChunker2
		fresh := func() *languagetool.AnalyzedSentence {
			return tokenSentence("для", "годиться")
		}
		gotFull := contentPOSTagsMW(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "<adv>"), "%s full: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "<adv>"), "%s full: %v", label, gotFull[1])
		// Without multiword: no wrap
		gotNoM := contentPOSTagsMW(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasAnyAnglePOS(tags),
				"%s without multi token[%d]: %v", label, i, tags)
		}
	}
}

// TestUkrainianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS proves
// multiword (after pre, before XML) attaches official multiword wrap tags and XML still fires.
func TestUkrainianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS(t *testing.T) {
	requireUKHybridResources(t)
	full := NewUkrainianHybridDisambiguator()

	out := full.Disambiguate(mwSent("для", "годиться"))
	got := contentPOSTagsMW(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<adv>"), "для after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "<adv>"), "годиться after full order: %v", got[1])

	out = full.Disambiguate(mwSent("на", "жаль"))
	got = contentPOSTagsMW(out)
	require.True(t, hasExactPOS(got[0], "<insert>"), "на after full: %v", got[0])
	require.True(t, hasExactPOS(got[1], "<insert>"), "жаль after full: %v", got[1])

	out = full.Disambiguate(mwSent("а", "капела"))
	got = contentPOSTagsMW(out)
	require.True(t, hasExactPOS(got[0], "<adv>"), "а after full: %v", got[0])

	out = full.Disambiguate(mwSent("від", "і", "до"))
	got = contentPOSTagsMW(out)
	require.True(t, hasExactPOS(got[0], "<adv>"), "від after full: %v", got[0])
	require.True(t, hasExactPOS(got[2], "<adv>"), "до after full: %v", got[2])

	// First-cap still works under allowFirstCapitalized=true after full order.
	out = full.Disambiguate(mwSent("Для", "годиться"))
	got = contentPOSTagsMW(out)
	require.True(t, hasExactPOS(got[0], "<adv>"), "Для after full: %v", got[0])

	// XML effects still fire after multiword (multiword no-op on these surfaces).
	requireImmunized(t, full.Disambiguate(tokenSentence("-ськ-")), "-ськ-")
	require.True(t, hasExactPOS(contentPOSTagsMW(full.Disambiguate(tokenSentence("і", "тд")))[1], "noninfl:bad"))
	require.Contains(t, posTagsOn(tokenBySurface(
		full.Disambiguate(tokenSentence("святого", "Юра")), "Юра")),
		"noun:anim:m:v_rod:prop:fname")
}

// TestUkrainianHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == XML(multiword(pre(input))) for official isolation surfaces.
func TestUkrainianHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireUKHybridResources(t)
	mw := UkrainianMultiwordChunkerDefault()
	xml := UkrainianXmlRuleDisambiguator()
	full := NewUkrainianHybridDisambiguator()
	preOnly := &UkrainianHybridDisambiguator{Simple: NewSimpleDisambiguator()}
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantPOS0 string
		wantPOS1 string
		wantPOS  map[string]string
		wantImm  map[string]bool
		noAngles bool
		// optional pre-outcome check on first content token partial POS
		noPartial  string
		yesPartial string
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("для", "годиться") },
			label:    "для годиться",
			wantPOS0: "<adv>",
			wantPOS1: "<adv>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("на", "жаль") },
			label:    "на жаль",
			wantPOS0: "<insert>",
			wantPOS1: "<insert>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("а", "капела") },
			label:    "а капела",
			wantPOS0: "<adv>",
			wantPOS1: "<adv>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("як", "правило") },
			label:    "як правило",
			wantPOS0: "<insert>",
			wantPOS1: "<insert>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("від", "і", "до") },
			label:    "від і до",
			wantPOS0: "<adv>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("Для", "годиться") },
			label:    "Для годиться first-cap",
			wantPOS0: "<adv>",
			wantPOS1: "<adv>",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("НА", "ЖАЛЬ") },
			label:    "НА ЖАЛЬ all-caps denied",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("На", "Жаль") },
			label:    "На Жаль titlecase denied",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return mwSent("Zxqwv", "Plmnb") },
			label:    "random non-listed",
			noAngles: true,
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("-ськ-") },
			label:   "-ськ-",
			wantImm: map[string]bool{"-ськ-": true},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("і", "тд") },
			label:   "і тд",
			wantPOS: map[string]string{"тд": "noninfl:bad"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("святого", "Юра") },
			label:   "святого Юра",
			wantPOS: map[string]string{"Юра": "noun:anim:m:v_rod:prop:fname"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("ц.", "р.") },
			label:   "ц. р.",
			wantPOS: map[string]string{"ц.": "adj:m:v_rod:pron:dem", "р.": "noun:inanim:m:v_rod"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("номер", "17-а") },
			label:   "номер 17-а",
			wantPOS: map[string]string{"17-а": "noninfl"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("Кот", "д'Івуара") },
			label:   "Кот д'Івуара",
			wantPOS: map[string]string{"Кот": "noninfl:foreign:prop:geo:bad"},
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				p, l1 := "noun:inanim:m:v_mis", "зв'язок"
				p2, l2 := "noun:inanim:m:v_rod", "зв'язок"
				atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
					languagetool.NewAnalyzedToken("Зв'язку", &p, &l1),
					languagetool.NewAnalyzedToken("Зв'язку", &p2, &l2),
				}, 0)
				start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
					languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil),
				}, 0)
				return languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, atr})
			},
			label:      "removeVmis pre",
			noPartial:  "v_mis",
			yesPartial: "v_rod",
		},
	}
	for _, tc := range cases {
		outFull := full.Disambiguate(tc.fresh())
		// Java: preDisambiguate; disambiguator.disambiguate(chunker.disambiguate(input))
		outManual := ukJavaOrderComposition(preOnly, mw, xml, tc.fresh())

		gotFull := contentPOSTagsMW(outFull)
		gotManual := contentPOSTagsMW(outManual)
		require.Equal(t, len(gotFull), len(gotManual), tc.label+" content POS count")
		if tc.wantPOS0 != "" {
			require.True(t, hasExactPOS(gotFull[0], tc.wantPOS0), "%s full POS0: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotManual[0], tc.wantPOS0), "%s manual POS0: %v", tc.label, gotManual[0])
			if tc.wantPOS1 != "" {
				require.True(t, hasExactPOS(gotFull[1], tc.wantPOS1), "%s full POS1: %v", tc.label, gotFull[1])
				require.True(t, hasExactPOS(gotManual[1], tc.wantPOS1), "%s manual POS1: %v", tc.label, gotManual[1])
			}
		} else if tc.noAngles || (len(tc.wantPOS) == 0 && len(tc.wantImm) == 0 && tc.noPartial == "") {
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
			require.True(t, hasExactPOS(posTagsOn(trF), wantPOS),
				"%s full %q want %s in %v", tc.label, surface, wantPOS, posTagsOn(trF))
			require.True(t, hasExactPOS(posTagsOn(trM), wantPOS),
				"%s manual %q want %s in %v", tc.label, surface, wantPOS, posTagsOn(trM))
		}
		if tc.noPartial != "" {
			require.False(t, outFull.GetTokensWithoutWhitespace()[1].HasPartialPosTag(tc.noPartial),
				"%s full must not have %s", tc.label, tc.noPartial)
			require.False(t, outManual.GetTokensWithoutWhitespace()[1].HasPartialPosTag(tc.noPartial),
				"%s manual must not have %s", tc.label, tc.noPartial)
		}
		if tc.yesPartial != "" {
			require.True(t, outFull.GetTokensWithoutWhitespace()[1].HasPartialPosTag(tc.yesPartial),
				"%s full must have %s", tc.label, tc.yesPartial)
			require.True(t, outManual.GetTokensWithoutWhitespace()[1].HasPartialPosTag(tc.yesPartial),
				"%s manual must have %s", tc.label, tc.yesPartial)
		}
	}
}

// TestUkrainianHybridDisambiguator_StageOrderIsPreThenMultiwordThenXML proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestUkrainianHybridDisambiguator_StageOrderIsPreThenMultiwordThenXML(t *testing.T) {
	requireUKHybridResources(t)
	mw := UkrainianMultiwordChunkerDefault()
	xml := UkrainianXmlRuleDisambiguator()
	full := NewUkrainianHybridDisambiguator()

	// Multiword surface: only Chunker produces multiword wrap; full keeps wrap after pre+XML.
	{
		fresh := func() *languagetool.AnalyzedSentence { return mwSent("для", "годиться") }
		gotFull := contentPOSTagsMW(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "<adv>"), "full: %v", gotFull[0])

		// Without Chunker → no multiword wrap
		noChunk := &UkrainianHybridDisambiguator{Simple: NewSimpleDisambiguator(), Inner: xml}
		gotNoC := contentPOSTagsMW(noChunk.Disambiguate(fresh()))
		for i, tags := range gotNoC {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "<adv>"),
				"without chunker token[%d]: %v", i, tags)
		}
		// Only Chunker field on hybrid (pre retags still run; no XML) → multiword wrap
		onlyC := &UkrainianHybridDisambiguator{Simple: NewSimpleDisambiguator(), Chunker: mw}
		gotOnlyC := contentPOSTagsMW(onlyC.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyC[0], "<adv>"), "only chunker hybrid: %v", gotOnlyC[0])
		require.True(t, hasExactPOS(gotOnlyC[1], "<adv>"), "only chunker hybrid close: %v", gotOnlyC[1])
	}

	// XML-only surface: only Inner immunizes.
	{
		freshInfx := func() *languagetool.AnalyzedSentence { return tokenSentence("-ськ-") }
		requireImmunized(t, full.Disambiguate(freshInfx()), "-ськ-")
		requireNotImmunized(t, (&UkrainianHybridDisambiguator{
			Simple: NewSimpleDisambiguator(), Chunker: mw,
		}).Disambiguate(freshInfx()), "-ськ-")
		requireImmunized(t, (&UkrainianHybridDisambiguator{
			Simple: NewSimpleDisambiguator(), Inner: xml,
		}).Disambiguate(freshInfx()), "-ськ-")
	}

	// Pre surface: only pre (retags) strips v_mis; multiword/XML alone keep it.
	{
		freshVM := func() *languagetool.AnalyzedSentence {
			p, l1 := "noun:inanim:m:v_mis", "зв'язок"
			p2, l2 := "noun:inanim:m:v_rod", "зв'язок"
			atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
				languagetool.NewAnalyzedToken("Зв'язку", &p, &l1),
				languagetool.NewAnalyzedToken("Зв'язку", &p2, &l2),
			}, 0)
			start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
				languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil),
			}, 0)
			return languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, atr})
		}
		require.False(t, full.Disambiguate(freshVM()).GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"))
		require.True(t, mw.Disambiguate(freshVM()).GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"))
		require.True(t, xml.Disambiguate(freshVM()).GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"))
		// Hybrid with nil stages still runs retags (pre body)
		require.False(t, (&UkrainianHybridDisambiguator{}).Disambiguate(freshVM()).
			GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_mis"))
	}

	// Call-order: Chunker then Rules after pre (Java nested call after preDisambiguate).
	{
		var order []string
		rulesStub := &ukOrderStage{name: "rules", order: &order}
		chunkStub := &ukOrderStage{name: "chunker", order: &order}
		d := &UkrainianHybridDisambiguator{
			Simple:  NewSimpleDisambiguator(),
			Chunker: chunkStub,
			Inner:   rulesStub,
		}
		d.Disambiguate(tokenSentence("x"))
		require.Equal(t, []string{"chunker", "rules"}, order,
			"Java: pre then disambiguator.disambiguate(chunker.disambiguate(input)) → multiword then XML")
	}
}

// ukOrderStage records Disambiguate call order for stage-order proof.
type ukOrderStage struct {
	disambiguation.AbstractDisambiguator
	name  string
	order *[]string
}

func (s *ukOrderStage) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s.order != nil {
		*s.order = append(*s.order, s.name)
	}
	return input
}

var _ disambiguation.Disambiguator = (*ukOrderStage)(nil)
