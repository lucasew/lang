package ru

// Outcome twins for RussianHybridDisambiguator full stage order:
// Java RussianHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(input))
// i.e. MultiWordChunker.getInstance("/ru/multiwords.txt") defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling) FIRST, then
// XmlRuleDisambiguator(Russian, useGlobal=false).
//
// CRITICAL: multiword→XML (same as Romance hybrids; opposite of Polish XML→multiword).
// Official russian.dict is not required: token-built AnalyzedSentence patterns
// (same helpers as ACCEPTed multiword/XML stage leaves).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireRUHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverRussianMultiwords() == "" {
		t.Skip("official ru/multiwords.txt not discoverable")
	}
	if DiscoverRussianDisambiguationXML() == "" {
		t.Skip("official ru/disambiguation.xml not discoverable")
	}
}

// tokenSentence builds SENT_START + tokens with spaces between word tokens.
func tokenSentence(words ...string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(multiwordTokens(words...))
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

// TestNewRussianHybridDisambiguator_WiresBothStages proves Java constructor
// eagerly builds multiwords Chunker and XmlRuleDisambiguator when the same
// official resources Java loads are present — with Russian flags.
func TestNewRussianHybridDisambiguator_WiresBothStages(t *testing.T) {
	requireRUHybridResources(t)

	mw := RussianMultiWordChunker()
	xml := RussianXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	d := NewRussianHybridDisambiguator()
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/ru/multiwords.txt) defaults F,F,F")
	require.NotNil(t, d.Rules,
		"disambiguator = new XmlRuleDisambiguator(Russian.getInstance()) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Russian multiwords defaults (no invent):
	// NO setRemovePreviousTags, NO setIgnoreSpelling
	// allowFirstCapitalized/allowAllUppercase/allowTitlecase false (outcome-tested)
	require.False(t, mw.RemovePreviousTags, "Russian multiwords does NOT setRemovePreviousTags")
	require.False(t, mw.AddIgnoreSpelling, "Russian multiwords does NOT setIgnoreSpelling")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official RU pack has numeric, VERB-KA, case, ADV_OB* rules.
	require.GreaterOrEqual(t, len(xml.Rules), 50,
		"Russian XmlRuleDisambiguator must load official ru/disambiguation.xml rules")
}

// TestRussianHybridDisambiguator_OrderMultiwordThenXML proves stage isolation vs
// full Java order with Java-visible POS outcomes.
// Order is multiword first → XML second (NOT XML→multiword like Polish).
func TestRussianHybridDisambiguator_OrderMultiwordThenXML(t *testing.T) {
	requireRUHybridResources(t)

	mw := RussianMultiWordChunker()
	xml := RussianXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyMulti := &RussianHybridDisambiguator{Chunker: mw}
	onlyXML := &RussianHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(sent))
	}
	// Reverse of Java (Polish-style XML→multiword) — used for leave-one-out / order contrast.
	reverseOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return mw.Disambiguate(xml.Disambiguate(sent))
	}
	full := NewRussianHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Multiword-only phrase: "до мажор" in ru/multiwords, not an XML invent ---
	// Multiword alone → <NN:Masc> / </NN:Masc> (no removePreviousTags)
	// XML alone → no multiword POS
	// Full hybrid (multiword then XML) → multiword POS survives
	// Without multiword → no multiword POS
	{
		label := "до мажор"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("до", "мажор"))
		}

		gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
		require.Len(t, gotM, 2, label)
		require.True(t, hasExactPOS(gotM[0], "<NN:Masc>"), "%s multiword-only до: %v", label, gotM[0])
		require.True(t, hasExactPOS(gotM[1], "</NN:Masc>"), "%s multiword-only мажор: %v", label, gotM[1])

		gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
		require.Len(t, gotX, 2, label)
		for i, tags := range gotX {
			require.False(t, hasExactPOS(tags, "<NN:Masc>") || hasExactPOS(tags, "</NN:Masc>") || hasAnyAnglePOS(tags),
				"%s xml-only token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 2, label)
		require.True(t, hasExactPOS(gotFull[0], "<NN:Masc>"),
			"%s full hybrid до (multiword then XML): %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[1], "</NN:Masc>"),
			"%s full hybrid мажор: %v", label, gotFull[1])

		gotJO := contentPOSTags(javaOrder(fresh()))
		require.True(t, hasExactPOS(gotJO[0], "<NN:Masc>"), "%s javaOrder до: %v", label, gotJO[0])
		require.True(t, hasExactPOS(gotJO[1], "</NN:Masc>"), "%s javaOrder мажор: %v", label, gotJO[1])

		// Leave multiword out: XML-only hybrid has no multiword POS.
		noMulti := &RussianHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasAnyAnglePOS(tags),
				"%s without multiword token[%d] must have no multiword POS, got %v", label, i, tags)
		}

		// No setIgnoreSpelling on Russian multiwords.
		for i, tr := range full.Disambiguate(fresh()).GetTokens() {
			if i == 0 || tr.IsWhitespace() {
				continue
			}
			require.False(t, tr.IsIgnoredBySpeller(),
				"%s full hybrid token %q must not ignore spelling via multiwords", label, tr.GetToken())
		}

		// Reverse order still keeps multiword POS here (XML does not wipe NN:Masc on this surface).
		// Order proof for independent stages is composition + leave-one-out + stage call order.
		gotRev := contentPOSTags(reverseOrder(fresh()))
		require.True(t, hasExactPOS(gotRev[0], "<NN:Masc>") || hasExactPOS(gotRev[0], "NN:Masc"),
			"%s reverse still has multiword tag (stages independent on this surface): %v", label, gotRev[0])
	}

	// Multiword 3-token: "откуда ни возьмись" → FR open/close
	{
		label := "откуда ни возьмись"
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("откуда", "ни", "возьмись"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.Len(t, gotFull, 3, label)
		require.True(t, hasExactPOS(gotFull[0], "<FR>"), "%s full откуда: %v", label, gotFull[0])
		require.True(t, hasExactPOS(gotFull[2], "</FR>"), "%s full возьмись: %v", label, gotFull[2])
		require.False(t, hasAnyAnglePOS(gotFull[1]), "%s interior ни: %v", label, gotFull[1])

		// Without multiword → no FR angles
		noMulti := &RussianHybridDisambiguator{Rules: xml}
		gotNoM := contentPOSTags(noMulti.Disambiguate(fresh()))
		for i, tags := range gotNoM {
			require.False(t, hasExactPOS(tags, "<FR>") || hasExactPOS(tags, "</FR>"),
				"%s without multi token[%d]: %v", label, i, tags)
		}
	}

	// Multiword 2-token ADV / CONJ / FR (XML does not rewrite close tag on these surfaces)
	{
		for _, tc := range []struct {
			parts []string
			open  string
			close string
			label string
		}{
			{[]string{"пиши", "пропал"}, "<FR>", "</FR>", "пиши пропал"},
			{[]string{"черт", "возьми"}, "<CONJ>", "</CONJ>", "черт возьми"},
			{[]string{"будь", "здоров"}, "<ADV>", "</ADV>", "будь здоров"},
			{[]string{"в", "будущем"}, "<ADV>", "</ADV>", "в будущем"},
			{[]string{"в", "целом"}, "<ADV>", "</ADV>", "в целом"},
		} {
			fresh := func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...))
			}
			gotFull := contentPOSTags(full.Disambiguate(fresh()))
			require.True(t, hasExactPOS(gotFull[0], tc.open), "%s full open: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotFull[len(gotFull)-1], tc.close),
				"%s full close: %v", tc.label, gotFull[len(gotFull)-1])
			// multiword alone same; XML alone no multiword angles
			gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
			require.True(t, hasExactPOS(gotM[0], tc.open), "%s multi-only: %v", tc.label, gotM[0])
			gotX := contentPOSTags(onlyXML.Disambiguate(fresh()))
			for i, tags := range gotX {
				require.False(t, hasAnyAnglePOS(tags),
					"%s xml-only token[%d]: %v", tc.label, i, tags)
			}
		}
	}

	// Order-sensitive multiword+XML: official multiwords + ADV_OB* rewrite second token.
	// Java multiword→XML: open angle survives, close becomes plain ADV (ADV_OB match).
	// Reverse XML→multiword: multiword last re-adds </ADV> close angle.
	// Proves stages are not commutative on these official surfaces.
	{
		for _, tc := range []struct {
			parts []string
			label string
		}{
			{[]string{"до", "свидания"}, "до свидания"},   // multiwords ADV + ADV_OB4
			{[]string{"в", "дальнейшем"}, "в дальнейшем"}, // multiwords ADV + ADV_OB1
		} {
			fresh := func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...))
			}
			gotM := contentPOSTags(onlyMulti.Disambiguate(fresh()))
			require.True(t, hasExactPOS(gotM[0], "<ADV>"), "%s multi open: %v", tc.label, gotM[0])
			require.True(t, hasExactPOS(gotM[1], "</ADV>"), "%s multi close: %v", tc.label, gotM[1])

			gotFull := contentPOSTags(full.Disambiguate(fresh()))
			require.True(t, hasExactPOS(gotFull[0], "<ADV>"),
				"%s full hybrid open survives multiword-then-XML: %v", tc.label, gotFull[0])
			require.True(t, hasExactPOS(gotFull[1], "ADV"),
				"%s full hybrid close flattened by ADV_OB* after multiword: %v", tc.label, gotFull[1])
			require.False(t, hasExactPOS(gotFull[1], "</ADV>"),
				"%s full must not keep multiword close after XML: %v", tc.label, gotFull[1])

			gotJO := contentPOSTags(javaOrder(fresh()))
			require.True(t, hasExactPOS(gotJO[0], "<ADV>"), "%s javaOrder open: %v", tc.label, gotJO[0])
			require.True(t, hasExactPOS(gotJO[1], "ADV"), "%s javaOrder close: %v", tc.label, gotJO[1])

			// Reverse differs: multiword last restores </ADV>
			gotRev := contentPOSTags(reverseOrder(fresh()))
			require.True(t, hasExactPOS(gotRev[1], "</ADV>"),
				"%s reverse XML→multiword keeps multiword close (proves multiword-before-XML): %v",
				tc.label, gotRev[1])
			require.False(t, hasExactPOS(gotRev[1], "ADV") && !hasExactPOS(gotRev[1], "</ADV>"),
				"%s reverse must differ from Java order on close token: %v", tc.label, gotRev[1])

			// Leave multiword out: ADV_OB may not invent the same multiword-open path alone
			// (token-built without prior readings — ADV_OB match alone may no-op).
			// Leave XML out: multiword close remains </ADV>
			noXML := &RussianHybridDisambiguator{Chunker: mw}
			gotNoX := contentPOSTags(noXML.Disambiguate(fresh()))
			require.True(t, hasExactPOS(gotNoX[1], "</ADV>"),
				"%s without XML keeps multiword close: %v", tc.label, gotNoX[1])
		}
	}

	// Casing flags: allowFirstCapitalized=false on multiwords for unlisted capital
	// (будь здоров has no "Будь здоров" line; До мажор IS listed).
	{
		freshCap := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("Будь", "здоров"))
		}
		got := contentPOSTags(full.Disambiguate(freshCap()))
		for i, tags := range got {
			require.False(t, hasAnyAnglePOS(tags),
				"Будь здоров first-cap denied full hybrid token[%d]: %v", i, tags)
		}
		// Listed capital form still matches (explicit multiwords entry, not flag invent).
		gotListed := contentPOSTags(full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("До", "мажор"))))
		require.True(t, hasExactPOS(gotListed[0], "<NN:Masc>"), "До мажор listed: %v", gotListed[0])
	}

	// --- (2) XML-only effects: ru NumD_*_tag (token-built; no russian.dict) ---
	// Chunker does not invent NumD_*; XML stage does.
	{
		// NumD_D: 73 / 2 / 4 match \d*[234] (exceptions for *12/*13/*14)
		sent73 := tokenSentence("73")
		gotM := contentPOSTags(onlyMulti.Disambiguate(sent73))
		require.False(t, hasExactPOS(gotM[0], "NumD_D"), "73 multiword-only must not invent NumD_D: %v", gotM[0])
		gotX := contentPOSTags(onlyXML.Disambiguate(tokenSentence("73")))
		require.True(t, hasExactPOS(gotX[0], "NumD_D"), "73 xml-only NumD_D: %v", gotX[0])
		gotFull := contentPOSTags(full.Disambiguate(tokenSentence("73")))
		require.True(t, hasExactPOS(gotFull[0], "NumD_D"), "73 full hybrid NumD_D: %v", gotFull[0])
		gotJO := contentPOSTags(javaOrder(tokenSentence("73")))
		require.True(t, hasExactPOS(gotJO[0], "NumD_D"), "73 javaOrder NumD_D: %v", gotJO[0])

		// leave-one-out: without Rules, multiword alone must not invent NumD
		noXML := &RussianHybridDisambiguator{Chunker: mw}
		gotNoX := contentPOSTags(noXML.Disambiguate(tokenSentence("73")))
		require.False(t, hasExactPOS(gotNoX[0], "NumD_D"), "73 without XML: %v", gotNoX[0])

		// NumD_S: 71
		got71 := contentPOSTags(full.Disambiguate(tokenSentence("71")))
		require.True(t, hasExactPOS(got71[0], "NumD_S"), "71 full NumD_S: %v", got71[0])

		// NumD_P: 75
		got75 := contentPOSTags(full.Disambiguate(tokenSentence("75")))
		require.True(t, hasExactPOS(got75[0], "NumD_P"), "75 full NumD_P: %v", got75[0])

		// ABR_with_dot: проф.
		sentProf := tokenSentence("проф", ".")
		gotProfX := contentPOSTags(onlyXML.Disambiguate(sentProf))
		// first content token should gain ABR
		require.True(t, hasExactPOS(gotProfX[0], "ABR"), "проф. xml-only ABR: %v", gotProfX[0])
		gotProfFull := contentPOSTags(full.Disambiguate(tokenSentence("проф", ".")))
		require.True(t, hasExactPOS(gotProfFull[0], "ABR"), "проф. full ABR: %v", gotProfFull[0])
		// multiword alone no ABR
		gotProfM := contentPOSTags(onlyMulti.Disambiguate(tokenSentence("проф", ".")))
		require.False(t, hasExactPOS(gotProfM[0], "ABR"), "проф. multi-only no ABR: %v", gotProfM[0])
	}
}

// TestRussianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS proves
// multiword first (before XML) still attaches official multiword open/close POS
// and XML still fires after.
func TestRussianHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockMultiwordPOS(t *testing.T) {
	requireRUHybridResources(t)
	full := NewRussianHybridDisambiguator()

	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("до", "мажор")))
	got := contentPOSTags(out)
	require.Len(t, got, 2)
	require.True(t, hasExactPOS(got[0], "<NN:Masc>"), "до after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</NN:Masc>"), "мажор after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("пиши", "пропал")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<FR>"), "пиши after full order: %v", got[0])
	require.True(t, hasExactPOS(got[1], "</FR>"), "пропал after full order: %v", got[1])

	out = full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("откуда", "ни", "возьмись")))
	got = contentPOSTags(out)
	require.True(t, hasExactPOS(got[0], "<FR>"), "откуда after full: %v", got[0])
	require.True(t, hasExactPOS(got[2], "</FR>"), "возьмись after full: %v", got[2])

	// XML effects still fire after multiword (multiword no-op on this surface).
	got73 := contentPOSTags(full.Disambiguate(tokenSentence("73")))
	require.True(t, hasExactPOS(got73[0], "NumD_D"), "73 after full order: %v", got73[0])
	gotProf := contentPOSTags(full.Disambiguate(tokenSentence("проф", ".")))
	require.True(t, hasExactPOS(gotProf[0], "ABR"), "проф. after full order: %v", gotProf[0])
}

// TestRussianHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == XML(multiword(input)) for official isolation surfaces.
func TestRussianHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireRUHybridResources(t)
	mw := RussianMultiWordChunker()
	xml := RussianXmlRuleDisambiguator()
	full := NewRussianHybridDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantPOS0 string // empty = no multiword angle POS required
		wantPOS1 string
		wantPOS  map[string]string // surface → required POS (for XML-only)
	}
	cases := []caseT{
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("до", "мажор"))
			},
			label:    "до мажор",
			wantPOS0: "<NN:Masc>",
			wantPOS1: "</NN:Masc>",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("до", "минор"))
			},
			label:    "до минор",
			wantPOS0: "<NN:Masc>",
			wantPOS1: "</NN:Masc>",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("пиши", "пропал"))
			},
			label:    "пиши пропал",
			wantPOS0: "<FR>",
			wantPOS1: "</FR>",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("откуда", "ни", "возьмись"))
			},
			label:    "откуда ни возьмись",
			wantPOS0: "<FR>",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("в", "целом"))
			},
			label:    "в целом",
			wantPOS0: "<ADV>",
			wantPOS1: "</ADV>",
		},
		{
			// multiword ADV then ADV_OB4 → close becomes plain ADV
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("до", "свидания"))
			},
			label:    "до свидания",
			wantPOS0: "<ADV>",
			wantPOS1: "ADV",
		},
		{
			// multiword ADV then ADV_OB1 → close becomes plain ADV
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("в", "дальнейшем"))
			},
			label:    "в дальнейшем",
			wantPOS0: "<ADV>",
			wantPOS1: "ADV",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("Будь", "здоров"))
			},
			label: "Будь здоров first-cap denied",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens("Zxqwv", "Plmnb"))
			},
			label: "random non-listed",
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("73") },
			label:   "73",
			wantPOS: map[string]string{"73": "NumD_D"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("71") },
			label:   "71",
			wantPOS: map[string]string{"71": "NumD_S"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("75") },
			label:   "75",
			wantPOS: map[string]string{"75": "NumD_P"},
		},
		{
			fresh:   func() *languagetool.AnalyzedSentence { return tokenSentence("проф", ".") },
			label:   "проф.",
			wantPOS: map[string]string{"проф": "ABR"},
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
		} else if len(tc.wantPOS) == 0 {
			// Non-listed / first-cap: no multiword invent angles required
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
		// Ignore parity.
		ft, mt := outFull.GetTokens(), outManual.GetTokens()
		require.Equal(t, len(ft), len(mt), tc.label+" token count")
		for i := range ft {
			if i == 0 || ft[i].IsWhitespace() {
				continue
			}
			require.Equal(t, ft[i].IsIgnoredBySpeller(), mt[i].IsIgnoredBySpeller(),
				"%s token[%d]=%q ignore parity full vs javaOrder", tc.label, i, ft[i].GetToken())
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

// TestRussianHybridDisambiguator_StageOrderIsMultiwordThenXML proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestRussianHybridDisambiguator_StageOrderIsMultiwordThenXML(t *testing.T) {
	requireRUHybridResources(t)
	mw := RussianMultiWordChunker()
	xml := RussianXmlRuleDisambiguator()
	full := NewRussianHybridDisambiguator()

	// Multiword-only surface: only Chunker produces multiword POS.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("до", "мажор"))
		}
		gotFull := contentPOSTags(full.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotFull[0], "<NN:Masc>"), "full: %v", gotFull[0])

		// Without Chunker → no multiword POS
		noChunk := &RussianHybridDisambiguator{Rules: xml}
		gotNoC := contentPOSTags(noChunk.Disambiguate(fresh()))
		for i, tags := range gotNoC {
			require.False(t, hasAnyAnglePOS(tags), "without chunker token[%d]: %v", i, tags)
		}
		// Only Chunker → multiword POS
		onlyC := &RussianHybridDisambiguator{Chunker: mw}
		gotOnlyC := contentPOSTags(onlyC.Disambiguate(fresh()))
		require.True(t, hasExactPOS(gotOnlyC[0], "<NN:Masc>"), "only chunker: %v", gotOnlyC[0])
	}

	// XML-only surface: only Rules sets NumD_D.
	{
		sent := tokenSentence("73")
		gotFull := contentPOSTags(full.Disambiguate(sent))
		require.True(t, hasExactPOS(gotFull[0], "NumD_D"), "full 73: %v", gotFull[0])
		gotNoXML := contentPOSTags((&RussianHybridDisambiguator{Chunker: mw}).Disambiguate(tokenSentence("73")))
		require.False(t, hasExactPOS(gotNoXML[0], "NumD_D"), "chunker-only 73: %v", gotNoXML[0])
		gotOnlyXML := contentPOSTags((&RussianHybridDisambiguator{Rules: xml}).Disambiguate(tokenSentence("73")))
		require.True(t, hasExactPOS(gotOnlyXML[0], "NumD_D"), "xml-only 73: %v", gotOnlyXML[0])
	}

	// Call-order: Chunker then Rules (Java nested call: outer=disambiguator, inner=chunker).
	{
		var order []string
		rulesStub := &orderStage{name: "rules", order: &order}
		chunkStub := &orderStage{name: "chunker", order: &order}
		d := &RussianHybridDisambiguator{Rules: rulesStub, Chunker: chunkStub}
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
