package gl

// Outcome twins for GalicianHybridDisambiguator full stage order:
// Java GalicianHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(input))
// i.e. MultiWordChunker.getInstance("/gl/multiwords.txt") defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling) FIRST, then
// XmlRuleDisambiguator(Galician, useGlobal=false).
//
// CRITICAL: multiword→XML (Romance order; same as Russian/ES; opposite of Polish/Swedish).
// Official galician.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/XML stage leaves).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireGLHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverGalicianMultiwords() == "" {
		t.Skip("official gl/multiwords.txt not discoverable")
	}
	if DiscoverGalicianDisambiguationXML() == "" {
		t.Skip("official gl/disambiguation.xml not discoverable")
	}
}

// multiwordTokens builds SENT_START + alternating content/space tokens for MultiWordChunker.
func multiwordTokens(parts ...string) []*languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for i, p := range parts {
		if i > 0 {
			toks = append(toks, languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)))
		}
		toks = append(toks, languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(p, nil, nil)))
	}
	return toks
}

// tokenSentence builds SENT_START + tokens with spaces between word tokens.
func tokenSentence(words ...string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(multiwordTokens(words...))
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

func hasExactPOS(tags []string, want string) bool {
	for _, p := range tags {
		if p == want {
			return true
		}
	}
	return false
}

func hasAnyAnglePOS(tags []string) bool {
	for _, p := range tags {
		if strings.Contains(p, "<") {
			return true
		}
	}
	return false
}

func tokenBySurface(sent *languagetool.AnalyzedSentence, surface string) *languagetool.AnalyzedTokenReadings {
	if sent == nil {
		return nil
	}
	for _, tr := range sent.GetTokensWithoutWhitespace() {
		if tr != nil && tr.GetToken() == surface {
			return tr
		}
	}
	return nil
}

func requireIgnored(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.True(t, tr.IsIgnoredBySpeller(), "%q must be ignore_spelling", s)
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

// TestNewGalicianHybridDisambiguator_WiresBothStages proves Java constructor
// eagerly builds multiwords Chunker and XmlRuleDisambiguator when the same
// official resources Java loads are present — with Galician flags.
func TestNewGalicianHybridDisambiguator_WiresBothStages(t *testing.T) {
	requireGLHybridResources(t)

	mw := GalicianMultiWordChunker()
	xml := GalicianXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewGalicianHybridDisambiguator()
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/gl/multiwords.txt) defaults F,F,F")
	require.NotNil(t, d.Rules,
		"disambiguator = new XmlRuleDisambiguator(new Galician()) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Galician multiwords defaults (no invent):
	// NO setRemovePreviousTags, NO setIgnoreSpelling
	// allowFirstCapitalized/allowAllUppercase/allowTitlecase false (outcome-tested)
	require.False(t, mw.RemovePreviousTags, "Galician multiwords does NOT setRemovePreviousTags")
	require.False(t, mw.AddIgnoreSpelling, "Galician multiwords does NOT setIgnoreSpelling")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official GL pack loads ~220 <rule> elements (NOMES_PROPRIOS, NUMBER, ignore_spelling, …).
	require.GreaterOrEqual(t, len(xml.Rules), 200,
		"Galician XmlRuleDisambiguator must load official gl/disambiguation.xml rules")
	require.LessOrEqual(t, len(xml.Rules), 250,
		"Galician XmlRuleDisambiguator useGlobal=false should not append global pack")
}

// TestGalicianHybridDisambiguator_OrderMultiwordThenXML proves stage isolation vs
// full Java order with Java-visible POS / ignore_spelling / immunize outcomes.
// Order is multiword first → XML second (NOT XML→multiword like Polish/Swedish).
func TestGalicianHybridDisambiguator_OrderMultiwordThenXML(t *testing.T) {
	requireGLHybridResources(t)

	mw := GalicianMultiWordChunker()
	xml := GalicianXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyMulti := &GalicianHybridDisambiguator{Chunker: mw}
	onlyXML := &GalicianHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(sent))
	}
	// Reverse of Java (Polish-style XML→multiword) — used for leave-one-out / order contrast.
	reverseOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return mw.Disambiguate(xml.Disambiguate(sent))
	}
	full := NewGalicianHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Multiword + XML flatten: "abaixo de" in gl/multiwords as SP000 ---
	// Multiword alone → <SP000> / </SP000> (no removePreviousTags)
	// XML alone → no multiword POS
	// Full hybrid (multiword then XML) → plain SP000 SP000 (XML multiword-flatten rules)
	// Reverse XML→multiword → angles remain (multiword last)
	{
		label := "abaixo de"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("abaixo", "de"))
		}

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "<SP000>"), "%s multiword-only abaixo: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</SP000>"), "%s multiword-only de: %v", label, gotM[1])

		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<SP000>") || hasExactPOS(tags, "</SP000>") ||
				hasExactPOS(tags, "SP000") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "SP000"),
			"%s full hybrid abaixo flattened (multiword then XML): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "SP000"),
			"%s full hybrid de flattened: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full hybrid must not keep angle tags: %v %v", label, gotFull[0], gotFull[1])

		gotJO := contentPOSTags(javaOrder(fresh()))
		require.True(t, hasExactPOS(gotJO[0], "SP000"), "%s javaOrder abaixo: %v", label, gotJO[0])
		require.True(t, hasExactPOS(gotJO[1], "SP000"), "%s javaOrder de: %v", label, gotJO[1])

		// Reverse differs: multiword last restores angles (proves multiword-before-XML).
		gotRev := contentPOSTags(reverseOrder(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<SP000>"),
			"%s reverse XML→multiword keeps multiword open (proves multiword-before-XML): %v",
			label, gotRev[0])
		require.True(t, hasExactPOS(gotRev[1], "</SP000>"),
			"%s reverse keeps multiword close: %v", label, gotRev[1])

		// Leave multiword out: XML-only hybrid has no multiword POS.
		noMulti := &GalicianHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "SP000"),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		// Leave XML out: multiword close remains angle.
		noXML := &GalicianHybridDisambiguator{Chunker: mw}
		gotNoX := contentPOSTags(noXML.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotNoX[0], "<SP000>"),
			"%s without XML keeps multiword open: %v", label, gotNoX[0])
		require.True(t, hasExactPOS(gotNoX[1], "</SP000>"),
			"%s without XML keeps multiword close: %v", label, gotNoX[1])

		// No setIgnoreSpelling on Galician multiwords.
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid token %q must not ignore spelling via multiwords", label, tr.GetToken())
		}
	}

	// Multiword RG 2-token: "a bordo" → multi angles; full flattens to RG RG
	{
		label := "a bordo"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("a", "bordo"))
		}
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "<RG>"), "%s multi open: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</RG>"), "%s multi close: %v", label, gotM[1])

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "RG"), "%s full open flattened: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "RG"), "%s full close flattened: %v", label, gotFull[1])

		gotRev := contentPOSTags(reverseOrder(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<RG>"),
			"%s reverse keeps multiword open: %v", label, gotRev[0])
	}

	// Multiword CS 2-token: "aínda que"
	{
		label := "aínda que"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("aínda", "que"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "CS"), "%s full open: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "CS"), "%s full close: %v", label, gotFull[1])
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotM[0], "<CS>"), "%s multi-only: %v", label, gotM[0])
		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		for i, tags := range gotX {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "CS"),
				"%s xml-only token[%d]: %v", label, i, tags)
		}
	}

	// Multiword 3-token SP000: "á beira de" — middle interior empty on multi-only;
	// full hybrid XML flattens all three to SP000.
	{
		label := "á beira de"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("á", "beira", "de"))
		}
		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 3, label)
		require.True(t, hasExactPOS(gotM[0], "<SP000>"), "%s multi open: %v", label, gotM[0])
		require.False(t, hasAnyAnglePOS(gotM[1]), "%s multi interior: %v", label, gotM[1])
		require.True(t, hasExactPOS(gotM[2], "</SP000>"), "%s multi close: %v", label, gotM[2])

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 3, label)
		require.True(t, hasExactPOS(gotFull[0], "SP000"), "%s full á: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "SP000"), "%s full beira: %v", label, gotFull[1])
		require.True(t, hasExactPOS(gotFull[2], "SP000"), "%s full de: %v", label, gotFull[2])
	}

	// Dual-stage: "Los Angeles" is multiword NPCNG00_ AND XML immunize (LOS_ANGELES).
	// Full multiword→XML: NOMES_PROPRIOS_MULTIWORD flattens to NPCN000_ + immunize.
	{
		label := "Los Angeles dual-stage"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Los", "Angeles"))
		}

		outM := onlyMulti.Disambiguate(fresh())
		gotM := contentPOSTags(outM)
		require.True(t, hasExactPOS(gotM[0], "<NPCNG00_>"), "%s multi open: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</NPCNG00_>"), "%s multi close: %v", label, gotM[1])
		requireNotImmunized(t, outM, "Los", "Angeles")
		requireNotIgnored(t, outM, "Los", "Angeles")

		outX := onlyXML.Disambiguate(fresh())
		// XML alone: FOREIGN_PROPER_NAMES (Los + UNKNOWN) and/or LOS_ANGELES immunize.
		requireImmunized(t, outX, "Los", "Angeles")
		gotX := contentPOSTags(outX)
		// Must not invent multiword angle tags alone.
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<NPCNG00_>") || hasExactPOS(tags, "</NPCNG00_>"),
				"%s xml-only token[%d] no multiword angles: %v", label, i, tags)
		}

		outFull := full.Disambiguate(fresh())
		gotFull := contentPOSTags(outFull)
		// multiword→XML: NOMES_PROPRIOS CN flatten of <NPCNG…> → NPCN000_
		require.True(t, hasExactPOS(gotFull[0], "NPCN000_"),
			"%s full Los after multiword then XML flatten: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "NPCN000_"),
			"%s full Angeles: %v", label, gotFull[1])
		require.False(t, hasAnyAnglePOS(gotFull[0]) || hasAnyAnglePOS(gotFull[1]),
			"%s full must flatten multiword angles: %v %v", label, gotFull[0], gotFull[1])
		requireImmunized(t, outFull, "Los", "Angeles")

		// javaOrder composition matches full POS + immunize
		outJO := javaOrder(fresh())
		gotJO := contentPOSTags(outJO)
		require.ElementsMatch(t, gotFull[0], gotJO[0], "%s POS0 full vs javaOrder", label)
		require.ElementsMatch(t, gotFull[1], gotJO[1], "%s POS1 full vs javaOrder", label)
		requireImmunized(t, outJO, "Los", "Angeles")
	}

	// Casing flags: allowAllUppercase=false, allowTitlecase=false
	// "Abaixo de" is explicitly listed (exact match, not first-cap invent).
	// "ABAIXO DE" and "Abaixo De" are not listed → no multiword POS under F,F,F.
	{
		gotListed := contentPOSTags(full.Disambiguate(
			languagetool.NewAnalyzedSentence(multiwordTokens("Abaixo", "de"))))
		require.True(t, hasExactPOS(gotListed[0], "SP000"), "Abaixo de listed: %v", gotListed[0])

		gotAllCap := contentPOSTags(full.Disambiguate(
			languagetool.NewAnalyzedSentence(multiwordTokens("ABAIXO", "DE"))))
		for i, tags := range gotAllCap {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "SP000"),
				"ABAIXO DE all-caps denied full hybrid token[%d]: %v", i, tags)
		}

		gotTitle := contentPOSTags(full.Disambiguate(
			languagetool.NewAnalyzedSentence(multiwordTokens("Abaixo", "De"))))
		for i, tags := range gotTitle {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "SP000"),
				"Abaixo De titlecase denied full hybrid token[%d]: %v", i, tags)
		}
	}

	// --- (2) XML-only effects: ignore_spelling / immunize / NUMBER (token-built; no dict) ---
	// Chunker does not set ignore_spelling or immunize; XML stage does.
	{
		// en vogue — ignore_spelling only in XML (EN_VOGUE)
		sentEV := tokenSentence("en", "vogue")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentEV), "en", "vogue")
		requireIgnored(t, onlyXML.Disambiguate(sentEV), "en", "vogue")
		requireIgnored(t, full.Disambiguate(sentEV), "en", "vogue")
		requireIgnored(t, javaOrder(sentEV), "en", "vogue")

		// leave-one-out: without Rules, multiword alone must not ignore
		noXML := &GalicianHybridDisambiguator{Chunker: mw}
		requireNotIgnored(t, noXML.Disambiguate(tokenSentence("en", "vogue")), "en", "vogue")

		// Las Vegas — immunize (XML only; not a multiword phrase that invents immunize)
		sentLV := tokenSentence("Las", "Vegas")
		requireNotImmunized(t, onlyMulti.Disambiguate(sentLV), "Las", "Vegas")
		requireImmunized(t, onlyXML.Disambiguate(sentLV), "Las", "Vegas")
		requireImmunized(t, full.Disambiguate(sentLV), "Las", "Vegas")
		requireNotImmunized(t, noXML.Disambiguate(tokenSentence("Las", "Vegas")), "Las", "Vegas")

		// Oak Ridge — immunize
		sentOR := tokenSentence("Oak", "Ridge")
		requireNotImmunized(t, onlyMulti.Disambiguate(sentOR), "Oak", "Ridge")
		requireImmunized(t, onlyXML.Disambiguate(sentOR), "Oak", "Ridge")
		requireImmunized(t, full.Disambiguate(sentOR), "Oak", "Ridge")

		// Rhythm and Blues — ignore_spelling
		sentRB := tokenSentence("Rhythm", "and", "Blues")
		requireNotIgnored(t, onlyMulti.Disambiguate(sentRB), "Rhythm", "and", "Blues")
		requireIgnored(t, onlyXML.Disambiguate(sentRB), "Rhythm", "and", "Blues")
		requireIgnored(t, full.Disambiguate(sentRB), "Rhythm", "and", "Blues")

		// NUMBER: 123 → Z0CN0 (XML only)
		sent123 := tokenSentence("123")
		gotM := contentPOSTags(onlyMulti.Disambiguate(sent123))
		require.False(t, hasExactPOS(gotM[0], "Z0CN0"), "123 multiword-only must not invent Z0CN0: %v", gotM[0])
		gotX := contentPOSTags(onlyXML.Disambiguate(tokenSentence("123")))
		require.True(t, hasExactPOS(gotX[0], "Z0CN0"), "123 xml-only Z0CN0: %v", gotX[0])
		gotFull := contentPOSTags(full.Disambiguate(tokenSentence("123")))
		require.True(t, hasExactPOS(gotFull[0], "Z0CN0"), "123 full hybrid Z0CN0: %v", gotFull[0])
		gotJO := contentPOSTags(javaOrder(tokenSentence("123")))
		require.True(t, hasExactPOS(gotJO[0], "Z0CN0"), "123 javaOrder Z0CN0: %v", gotJO[0])
		gotNoX := contentPOSTags(noXML.Disambiguate(tokenSentence("123")))
		require.False(t, hasExactPOS(gotNoX[0], "Z0CN0"), "123 without XML: %v", gotNoX[0])
	}
}

// TestGalicianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS proves
// multiword first (before XML) attaches official multiword POS and XML still fires after
// (flatten / ignore / immunize / NUMBER).
func TestGalicianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS(t *testing.T) {
	requireGLHybridResources(t)
	full := NewGalicianHybridDisambiguator()

	// Flattened multiword POS survives full order (XML multiword-flatten rules).
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("abaixo", "de")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "SP000"), "abaixo after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "SP000"), "de after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("a", "bordo")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "RG"), "a after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "RG"), "bordo after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("aínda", "que")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "CS"), "aínda after full: %v", got[0])
	require.True(t, hasExactPOS(got[1], "CS"), "que after full: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("á", "beira", "de")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "SP000"), "á after full: %v", got[0])
	require.True(t, hasExactPOS(got[2], "SP000"), "de after full: %v", got[2])

	// XML effects still fire after multiword (multiword no-op on these surfaces).
	requireIgnored(t, full.Disambiguate(tokenSentence("en", "vogue")), "en", "vogue")
	requireImmunized(t, full.Disambiguate(tokenSentence("Las", "Vegas")), "Las", "Vegas")
	got123 := contentPOSTags(full.Disambiguate(tokenSentence("123")))
	require.True(t, hasExactPOS(got123[0], "Z0CN0"), "123 after full order: %v", got123[0])
}

// TestGalicianHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == XML(multiword(input)) for official isolation surfaces.
func TestGalicianHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireGLHybridResources(t)
	mw := GalicianMultiWordChunker()
	xml := GalicianXmlRuleDisambiguator()
	full := NewGalicianHybridDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantPOS0 string // empty = no multiword/flatten POS required
		wantPOS1 string
		wantPOS  map[string]string // surface → required POS (for XML-only)
		wantIg   map[string]bool   // surface → ignore
		wantImm  map[string]bool   // surface → immunize
		noAngles bool              // require no multiword angle POS
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("abaixo", "de")) },
			label:    "abaixo de",
			wantPOS0: "SP000",
			wantPOS1: "SP000",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("a", "bordo")) },
			label:    "a bordo",
			wantPOS0: "RG",
			wantPOS1: "RG",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("aínda", "que")) },
			label:    "aínda que",
			wantPOS0: "CS",
			wantPOS1: "CS",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("acerca", "de")) },
			label:    "acerca de",
			wantPOS0: "SP000",
			wantPOS1: "SP000",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("menos", "mal")) },
			label:    "menos mal",
			wantPOS0: "RG",
			wantPOS1: "RG",
		},
		{
			// multiword NPCNG then NOMES_PROPRIOS flatten + LOS_ANGELES immunize
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Los", "Angeles")) },
			label:    "Los Angeles",
			wantPOS0: "NPCN000_",
			wantPOS1: "NPCN000_",
			wantImm:  map[string]bool{"Los": true, "Angeles": true},
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("á", "beira", "de")) },
			label:    "á beira de",
			wantPOS0: "SP000",
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("ABAIXO", "DE")) },
			label:    "ABAIXO DE all-caps denied",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Abaixo", "De")) },
			label:    "Abaixo De titlecase denied",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("Zxqwv", "Plmnb")) },
			label:    "random non-listed",
			noAngles: true,
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("en", "vogue") },
			label:  "en vogue",
			wantIg: map[string]bool{"en": true, "vogue": true},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("Las", "Vegas") },
			label:   "Las Vegas",
			wantImm: map[string]bool{"Las": true, "Vegas": true},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("Oak", "Ridge") },
			label:   "Oak Ridge",
			wantImm: map[string]bool{"Oak": true, "Ridge": true},
		},
		{
			fresh:  func() *languagetool.AnalyzedSentence { return tokenSentence("Rhythm", "and", "Blues") },
			label:  "Rhythm and Blues",
			wantIg: map[string]bool{"Rhythm": true, "and": true, "Blues": true},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("123") },
			label:   "123",
			wantPOS: map[string]string{"123": "Z0CN0"},
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
		} else if tc.noAngles || (len(tc.wantPOS) == 0 && len(tc.wantIg) == 0 && len(tc.wantImm) == 0) {
			// Non-listed / casing denied: no multiword invent angles required
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
		for surface, want := range tc.wantIg {
			trF := tokenBySurface(outFull, surface)
			trM := tokenBySurface(outManual, surface)
			require.NotNil(t, trF, "%s full missing %q", tc.label, surface)
			require.NotNil(t, trM, "%s manual missing %q", tc.label, surface)
			require.Equal(t, want, trF.IsIgnoredBySpeller(), "%s full %q ignore", tc.label, surface)
			require.Equal(t, want, trM.IsIgnoredBySpeller(), "%s manual %q ignore", tc.label, surface)
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

// TestGalicianHybridDisambiguator_StageOrderIsMultiwordThenXML proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestGalicianHybridDisambiguator_StageOrderIsMultiwordThenXML(t *testing.T) {
	requireGLHybridResources(t)
	mw := GalicianMultiWordChunker()
	xml := GalicianXmlRuleDisambiguator()
	full := NewGalicianHybridDisambiguator()

	// Multiword surface: only Chunker produces multiword POS; full order flattens via XML.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("abaixo", "de"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "SP000"), "full: %v", gotFull[0])

		// Without Chunker → no multiword POS / flatten target
		noChunk := &GalicianHybridDisambiguator{Rules: xml}
		gotNoC := contentPOSTags(noChunk.Disambiguate(fresh()))
		for i, tags := range gotNoC {
			require.False(t, hasAnyAnglePOS(tags) || hasExactPOS(tags, "SP000"),
				"without chunker token[%d]: %v", i, tags)
		}
		// Only Chunker → multiword angles (no XML flatten)
		onlyC := &GalicianHybridDisambiguator{Chunker: mw}
		gotOnlyC := contentPOSTags(onlyC.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyC[0], "<SP000>"), "only chunker: %v", gotOnlyC[0])
	}

	// XML-only surface: only Rules sets ignore_spelling.
	{
		sent := tokenSentence("en", "vogue")
		requireIgnored(t, full.Disambiguate(sent), "en", "vogue")
		requireNotIgnored(t, (&GalicianHybridDisambiguator{Chunker: mw}).Disambiguate(
			tokenSentence("en", "vogue")), "en", "vogue")
		requireIgnored(t, (&GalicianHybridDisambiguator{Rules: xml}).Disambiguate(
			tokenSentence("en", "vogue")), "en", "vogue")
	}

	// XML NUMBER: only Rules invents Z0CN0.
	{
		gotFull := contentPOSTags(full.Disambiguate(tokenSentence("123")))
		require.True(t, hasExactPOS(gotFull[0], "Z0CN0"), "full 123: %v", gotFull[0])
		gotNoXML := contentPOSTags((&GalicianHybridDisambiguator{Chunker: mw}).Disambiguate(tokenSentence("123")))
		require.False(t, hasExactPOS(gotNoXML[0], "Z0CN0"), "chunker-only 123: %v", gotNoXML[0])
		gotOnlyXML := contentPOSTags((&GalicianHybridDisambiguator{Rules: xml}).Disambiguate(tokenSentence("123")))
		require.True(t, hasExactPOS(gotOnlyXML[0], "Z0CN0"), "xml-only 123: %v", gotOnlyXML[0])
	}

	// Call-order: Chunker then Rules (Java nested call: outer=disambiguator, inner=chunker).
	{
		var order []string
		rulesStub := &orderStage{name: "rules", order: &order}
		chunkStub := &orderStage{name: "chunker", order: &order}
		d := &GalicianHybridDisambiguator{Rules: rulesStub, Chunker: chunkStub}
		d.Disambiguate(tokenSentence("x"))
		require.Equal(t, []string{"chunker", "rules"}, order,
			"Java: disambiguator.disambiguate(chunker.disambiguate(input)) → multiword then XML")
	}
}

// orderStage records Disambiguate call order for stage-order proof.
type orderStage struct {
	name  string
	order *[]string
}

func (s *orderStage) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s.order != nil {
		*s.order = append(*s.order, s.name)
	}
	return input
}

var _ interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
} = (*orderStage)(nil)
