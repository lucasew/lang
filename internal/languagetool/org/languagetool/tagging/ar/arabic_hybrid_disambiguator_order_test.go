package ar

// Outcome twins for ArabicHybridDisambiguator full stage order:
// Java ArabicHybridDisambiguator.disambiguate:
//   disambiguator.disambiguate(chunker.disambiguate(input))
// i.e. MultiWordChunker.getInstance("/ar/multiwords.txt") defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling) FIRST, then
// XmlRuleDisambiguator(new Arabic(), useGlobal=false).
//
// CRITICAL: multiword→XML (same Romance order as Irish/Galician/Russian/ES;
// opposite of Polish/Swedish XML→multiword).
// Official ar/multiwords.txt is comment-only → multiword stage is a no-op
// (empty maps; still eagerly wired). Do not invent multiword entries.
// XML stage still applies real official ar/disambiguation.xml outcomes
// (Keep_Only_verbs_*, Keep_Only_Nouns_after_Jar, Numeric_phrase_tags*).

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func requireARHybridResources(t *testing.T) {
	t.Helper()
	if DiscoverArabicMultiwords() == "" {
		t.Skip("official ar/multiwords.txt not discoverable")
	}
	if DiscoverArabicDisambiguationXML() == "" {
		t.Skip("official ar/disambiguation.xml not discoverable")
	}
}

func requireARHybridResourcesWithDict(t *testing.T) {
	t.Helper()
	requireARHybridResources(t)
	if DiscoverArabicPOSDict() == "" {
		t.Skip("arabic.dict not in tree")
	}
	EnsureDefaultArabicTagger()
	require.NotNil(t, DefaultArabicTagger)
	require.NotNil(t, DefaultArabicTagger.GetWordTagger())
}

// taggedARSentence ports the tagging half of Java TestTools.myAssert for AR:
// WordTokenizer + SRXSentenceTokenizer("ar") + ArabicTagger → AnalyzedSentence.
func taggedARSentence(input string) *languagetool.AnalyzedSentence {
	EnsureDefaultArabicTagger()
	tagger := DefaultArabicTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("ar")
	// Single sentence expected for isolation surfaces.
	var sentence string
	for _, s := range st.Tokenize(input) {
		sentence = s
		break
	}
	tokens := wt.Tokenize(sentence)
	var noWS []string
	for _, tok := range tokens {
		if arOrderIsWord(tok) {
			noWS = append(noWS, tok)
		}
	}
	aTokens := tagger.Tag(noWS)
	tokenArray := make([]*languagetool.AnalyzedTokenReadings, 0, len(tokens)+1)
	ss := languagetool.SentenceStartTagName
	tokenArray = append(tokenArray, languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", &ss, nil), 0))
	startPos := 0
	noWSCount := 0
	for _, tokenStr := range tokens {
		var posTag *languagetool.AnalyzedTokenReadings
		if arOrderIsWord(tokenStr) {
			posTag = aTokens[noWSCount]
			posTag.SetStartPos(startPos)
			noWSCount++
		} else {
			posTag = languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(tokenStr, nil, nil), startPos)
		}
		tokenArray = append(tokenArray, posTag)
		startPos += tokenizers.UTF16Len(tokenStr)
	}
	return languagetool.NewAnalyzedSentence(tokenArray)
}

func arOrderIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func formatAROrderSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				lemma, pos := "null", "null"
				if r.GetLemma() != nil {
					lemma = *r.GetLemma()
				}
				if r.GetPOSTag() != nil {
					pos = *r.GetPOSTag()
				}
				readings = append(readings, r.GetToken()+"/["+lemma+"]"+pos)
			}
		}
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}


// TestNewArabicHybridDisambiguator_WiresBothStages proves Java constructor
// eagerly builds multiwords Chunker and XmlRuleDisambiguator when the same
// official resources Java loads are present — with Arabic flags.
// Official multiwords may be empty; Chunker is still non-nil (Java constructs).
func TestNewArabicHybridDisambiguator_WiresBothStages(t *testing.T) {
	requireARHybridResources(t)

	mw := ArabicMultiWordChunker()
	xml := ArabicXmlRuleDisambiguator()
	require.NotNil(t, mw, "Java MultiWordChunker.getInstance even when multiwords comment-only")
	require.NotNil(t, xml)

	d := NewArabicHybridDisambiguator()
	require.NotNil(t, d.Chunker,
		"chunker = MultiWordChunker.getInstance(/ar/multiwords.txt) defaults F,F,F")
	require.NotNil(t, d.Rules,
		"disambiguator = new XmlRuleDisambiguator(new Arabic()) // useGlobal=false")
	require.Same(t, mw, d.Chunker)
	require.Same(t, xml, d.Rules)

	// Arabic multiwords defaults (no invent):
	// NO setRemovePreviousTags, NO setIgnoreSpelling
	require.False(t, mw.RemovePreviousTags, "Arabic multiwords does NOT setRemovePreviousTags")
	require.False(t, mw.AddIgnoreSpelling, "Arabic multiwords does NOT setIgnoreSpelling")
	// Official file comment-only → empty phrase list (still loaded).
	require.Empty(t, mw.Lines, "official ar/multiwords.txt comment-only; no invent phrases")

	// useGlobal=false: language XML only (no disambiguation-global pack appended).
	// Official AR pack: 5 rules (Keep_Only_verbs_*, Keep_Only_Nouns_after_Jar, Numeric_phrase_tags*).
	require.GreaterOrEqual(t, len(xml.Rules), 5,
		"Arabic XmlRuleDisambiguator must load official ar/disambiguation.xml rules")
	require.LessOrEqual(t, len(xml.Rules), 20,
		"Arabic XmlRuleDisambiguator useGlobal=false should not append global pack")
}

// TestArabicHybridDisambiguator_OrderMultiwordThenXML proves stage isolation vs
// full Java order. Multiword stage is no-op (empty official multiwords); XML
// still applies real disambiguation.xml outcomes. Order is multiword first → XML.
func TestArabicHybridDisambiguator_OrderMultiwordThenXML(t *testing.T) {
	requireARHybridResourcesWithDict(t)

	mw := ArabicMultiWordChunker()
	xml := ArabicXmlRuleDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	onlyMulti := &ArabicHybridDisambiguator{Chunker: mw}
	onlyXML := &ArabicHybridDisambiguator{Rules: xml}
	// Manual Java order composition (must match full hybrid).
	javaOrder := func(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		return xml.Disambiguate(mw.Disambiguate(sent))
	}
	full := NewArabicHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// --- (1) Multiword no-op: empty official multiwords invent no angle POS ---
	{
		for _, parts := range [][]string{
			{"قد", "عامل"},
			{"foo", "bar"},
			{"random", "phrase"},
		} {
			label := strings.Join(parts, " ")
			fresh := func() *languagetool.AnalyzedSentence {
				return languagetool.NewAnalyzedSentence(multiwordTokens(parts...))
			}
			for _, got := range [][][]string{
				contentPOSTags(onlyMulti.Disambiguate(fresh())),
				contentPOSTags(full.Disambiguate(fresh())),
				contentPOSTags(javaOrder(fresh())),
			} {
				for i, tags := range got {
					require.False(t, hasAnyAnglePOS(tags),
						"%s multiword no invent token[%d]: %v", label, i, tags)
				}
			}
			// No setIgnoreSpelling on multiwords.
			for i, tr := range full.Disambiguate(fresh()).GetTokens() {
				if i == 0 || tr.IsWhitespace() {
					continue
				}
				require.False(t, tr.IsIgnoredBySpeller(),
					"%s full hybrid token %q must not ignore spelling via multiwords", label, tr.GetToken())
			}
		}
	}

	// --- (2) XML Keep_Only_verbs_after_some_tools: قد + عامل → verb-only on عامل ---
	// Multiword alone → no filter; XML alone / full / javaOrder → same XML outcome.
	{
		label := "قد عامل"
		const wantXML = "/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---"
		const wantDemo = "/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---|عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---"

		fresh := func() *languagetool.AnalyzedSentence { return taggedARSentence("قد عامل") }

		// Multiword-only leaves tagger readings untouched (empty multiwords no-op).
		require.Equal(t, wantDemo, formatAROrderSentence(onlyMulti.Disambiguate(fresh())),
			"%s multiword-only must be no-op on tagged input", label)

		// XML alone applies Keep_Only_verbs.
		require.Equal(t, wantXML, formatAROrderSentence(onlyXML.Disambiguate(fresh())),
			"%s xml-only", label)

		// Full hybrid (multiword then XML) matches XML (multiword empty no-op).
		require.Equal(t, wantXML, formatAROrderSentence(full.Disambiguate(fresh())),
			"%s full hybrid", label)
		require.Equal(t, wantXML, formatAROrderSentence(javaOrder(fresh())),
			"%s javaOrder composition", label)

		// Leave multiword out: XML-only hybrid still has verb-only outcome.
		noMulti := &ArabicHybridDisambiguator{Rules: xml}
		require.Equal(t, wantXML, formatAROrderSentence(noMulti.Disambiguate(fresh())),
			"%s without multiword still XML", label)

		// Leave XML out: multiword-only keeps ambiguous noun+verb.
		noXML := &ArabicHybridDisambiguator{Chunker: mw}
		require.Equal(t, wantDemo, formatAROrderSentence(noXML.Disambiguate(fresh())),
			"%s without XML keeps ambiguous readings", label)
	}

	// --- (3) XML Keep_Only_Nouns_after_Jar: في + عامل → noun-only on عامل ---
	{
		label := "في عامل"
		const wantXML = "/[null]SENT_START في/[في]PR-;---;---|في/[في]PRD;---;---|في/[وَفَى]VW1;F1Y-i--;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---"
		fresh := func() *languagetool.AnalyzedSentence { return taggedARSentence("في عامل") }
		require.Equal(t, wantXML, formatAROrderSentence(full.Disambiguate(fresh())), label+" full")
		require.Equal(t, wantXML, formatAROrderSentence(javaOrder(fresh())), label+" javaOrder")
		require.Equal(t, wantXML, formatAROrderSentence(onlyXML.Disambiguate(fresh())), label+" xml-only")
		// multiword-only still has verb readings on عامل
		demo := formatAROrderSentence(onlyMulti.Disambiguate(fresh()))
		require.Contains(t, demo, "V41", "%s multiword-only must keep verb reading", label)
		require.NotEqual(t, wantXML, demo, "%s multiword-only must not apply XML", label)
	}

	// --- (4) XML Numeric_phrase_tags: ثلاثة وثلاثون ---
	{
		label := "ثلاثة وثلاثون"
		const wantXML = "/[null]SENT_START ثلاثة/[ثلاثة]NNU;M3--;---  /[null]null وثلاثون/[ثلاثون]NND;-3U-;W--"
		fresh := func() *languagetool.AnalyzedSentence { return taggedARSentence("ثلاثة وثلاثون") }
		require.Equal(t, wantXML, formatAROrderSentence(full.Disambiguate(fresh())), label+" full")
		require.Equal(t, wantXML, formatAROrderSentence(javaOrder(fresh())), label+" javaOrder")
		// multiword alone does not reduce to NN.*
		demo := formatAROrderSentence(onlyMulti.Disambiguate(fresh()))
		require.NotEqual(t, wantXML, demo, "%s multiword-only must not apply numeric XML", label)
	}
}

// TestArabicHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockXML proves
// empty multiword stage first does not block subsequent XML outcomes.
func TestArabicHybridDisambiguator_MultiwordBeforeXML_DoesNotBlockXML(t *testing.T) {
	requireARHybridResourcesWithDict(t)
	full := NewArabicHybridDisambiguator()
	require.NotNil(t, full.Chunker)
	require.NotNil(t, full.Rules)

	// Multiword angles never appear (empty official multiwords; no invent).
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("قد", "عامل")))
	for i, tags := range contentPOSTags(out) {
		require.False(t, hasAnyAnglePOS(tags), "no invent multiword POS token[%d]: %v", i, tags)
	}

	// XML effects still fire after multiword (multiword no-op on these surfaces).
	const wantQad = "/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---"
	require.Equal(t, wantQad, formatAROrderSentence(full.Disambiguate(taggedARSentence("قد عامل"))))

	const wantFi = "/[null]SENT_START في/[في]PR-;---;---|في/[في]PRD;---;---|في/[وَفَى]VW1;F1Y-i--;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---"
	require.Equal(t, wantFi, formatAROrderSentence(full.Disambiguate(taggedARSentence("في عامل"))))

	const wantNum = "/[null]SENT_START ثلاثة/[ثلاثة]NNU;M3--;---  /[null]null وثلاثون/[ثلاثون]NND;-3U-;W--"
	require.Equal(t, wantNum, formatAROrderSentence(full.Disambiguate(taggedARSentence("ثلاثة وثلاثون"))))

	const wantNum2 = "/[null]SENT_START ثلاثون/[ثلاثون]NND;-3U-;---  /[null]null ألف/[ألف]NNH;-1--;---"
	require.Equal(t, wantNum2, formatAROrderSentence(full.Disambiguate(taggedARSentence("ثلاثون ألف"))))
}

// TestArabicHybridDisambiguator_JavaOrderCompositionEqualsFull proves
// full.Disambiguate == XML(multiword(input)) for official isolation surfaces.
func TestArabicHybridDisambiguator_JavaOrderCompositionEqualsFull(t *testing.T) {
	requireARHybridResourcesWithDict(t)
	mw := ArabicMultiWordChunker()
	xml := ArabicXmlRuleDisambiguator()
	full := NewArabicHybridDisambiguator()
	require.NotNil(t, mw)
	require.NotNil(t, xml)

	type caseT struct {
		fresh    func() *languagetool.AnalyzedSentence
		label    string
		wantFmt  string // full formatted sentence if non-empty
		noAngles bool
	}
	cases := []caseT{
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("قد", "عامل")) },
			label:    "قد عامل untagged multiword tokens",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("foo", "bar")) },
			label:    "foo bar random",
			noAngles: true,
		},
		{
			fresh:    func() *languagetool.AnalyzedSentence { return languagetool.NewAnalyzedSentence(multiwordTokens("random", "phrase")) },
			label:    "random phrase",
			noAngles: true,
		},
		{
			fresh: func() *languagetool.AnalyzedSentence { return taggedARSentence("قد عامل") },
			label: "قد عامل tagged",
			wantFmt: "/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence { return taggedARSentence("في عامل") },
			label: "في عامل tagged",
			wantFmt: "/[null]SENT_START في/[في]PR-;---;---|في/[في]PRD;---;---|في/[وَفَى]VW1;F1Y-i--;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence { return taggedARSentence("بعد أن عامل") },
			label: "بعد أن عامل tagged",
			wantFmt: "/[null]SENT_START بعد/[بعد]NJ-;M1--;---|بعد/[بعد]NJ-;M1A-;---|بعد/[بعد]NJ-;M1I-;---|بعد/[بعد]NJ-;M1U-;---|بعد/[بَعُدَ]V30;M1H-pa-;---|بعد/[بَعُدَ]V30;M1H-pp-;---|بعد/[بَعِدَ]V30;M1H-pa-;---|بعد/[بَعِدَ]V30;M1H-pp-;---|بعد/[بَعَّدَ]V41;M1H-pa-;---|بعد/[بَعَّدَ]V41;M1H-pp-;---|بعد/[بَعَّدَ]V41;M1Y-i--;---|بعد/[عد]NJ-;M1--;-B-|بعد/[عد]NJ-;M1I-;-B-|بعد/[عد]NM-;M1--;-B-|بعد/[عد]NM-;M1I-;-B-  /[null]null أن/[آنَ]V-0;F3H-pa-;---|أن/[آنَ]V-0;F3H-pp-;---|أن/[آنَ]V-0;F3Y-i--;---|أن/[آنَ]V-0;M1Y-i--;---|أن/[أَنَّ]P--;----;---|أن/[وَأَى]VW1;M3Y-i--;---|أن/[وَنَى]VW1;M1I-fa0;---  /[null]null عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence { return taggedARSentence("ثلاثة وثلاثون") },
			label: "ثلاثة وثلاثون tagged",
			wantFmt: "/[null]SENT_START ثلاثة/[ثلاثة]NNU;M3--;---  /[null]null وثلاثون/[ثلاثون]NND;-3U-;W--",
		},
		{
			fresh: func() *languagetool.AnalyzedSentence { return taggedARSentence("ثلاثون ألف") },
			label: "ثلاثون ألف tagged",
			wantFmt: "/[null]SENT_START ثلاثون/[ثلاثون]NND;-3U-;---  /[null]null ألف/[ألف]NNH;-1--;---",
		},
	}
	for _, tc := range cases {
		outFull := full.Disambiguate(tc.fresh())
		// Java: disambiguator.disambiguate(chunker.disambiguate(input))
		outManual := xml.Disambiguate(mw.Disambiguate(tc.fresh()))

		gotFull := contentPOSTags(outFull)
		gotManual := contentPOSTags(outManual)
		require.Equal(t, len(gotFull), len(gotManual), tc.label+" content POS count")

		if tc.noAngles {
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
		// Full format parity (readings + surfaces).
		require.Equal(t, formatAROrderSentence(outFull), formatAROrderSentence(outManual),
			"%s format parity full vs javaOrder", tc.label)
		if tc.wantFmt != "" {
			require.Equal(t, tc.wantFmt, formatAROrderSentence(outFull), "%s full wantFmt", tc.label)
			require.Equal(t, tc.wantFmt, formatAROrderSentence(outManual), "%s manual wantFmt", tc.label)
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
	}
}

// TestArabicHybridDisambiguator_StageOrderIsMultiwordThenXML proves each stage
// occupies its Java slot via leave-one-out isolation and call-order recording.
func TestArabicHybridDisambiguator_StageOrderIsMultiwordThenXML(t *testing.T) {
	requireARHybridResourcesWithDict(t)
	mw := ArabicMultiWordChunker()
	xml := ArabicXmlRuleDisambiguator()
	full := NewArabicHybridDisambiguator()

	// Multiword surface: empty official multiwords → no invent POS from any stage combo.
	{
		fresh := func() *languagetool.AnalyzedSentence {
			return languagetool.NewAnalyzedSentence(multiwordTokens("foo", "bar"))
		}
		for _, got := range [][][]string{
			contentPOSTags(full.Disambiguate(fresh())),
			contentPOSTags((&ArabicHybridDisambiguator{Chunker: mw}).Disambiguate(fresh())),
			contentPOSTags((&ArabicHybridDisambiguator{Rules: xml}).Disambiguate(fresh())),
		} {
			for i, tags := range got {
				require.False(t, hasAnyAnglePOS(tags), "no invent angle POS token[%d]: %v", i, tags)
			}
		}
	}

	// XML-only surface: only Rules filters readings (Keep_Only_verbs on قد عامل).
	{
		const wantXML = "/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---"
		const wantDemo = "/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---|عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---"
		fresh := func() *languagetool.AnalyzedSentence { return taggedARSentence("قد عامل") }

		require.Equal(t, wantXML, formatAROrderSentence(full.Disambiguate(fresh())), "full")
		require.Equal(t, wantDemo, formatAROrderSentence((&ArabicHybridDisambiguator{Chunker: mw}).Disambiguate(fresh())),
			"chunker-only leaves ambiguous readings")
		require.Equal(t, wantXML, formatAROrderSentence((&ArabicHybridDisambiguator{Rules: xml}).Disambiguate(fresh())),
			"xml-only")
		// Without Chunker → XML still applies
		require.Equal(t, wantXML, formatAROrderSentence((&ArabicHybridDisambiguator{Rules: xml}).Disambiguate(fresh())))
	}

	// Call-order: Chunker then Rules (Java nested call: outer=disambiguator, inner=chunker).
	{
		var order []string
		rulesStub := &arOrderStage{name: "rules", order: &order}
		chunkStub := &arOrderStage{name: "chunker", order: &order}
		d := &ArabicHybridDisambiguator{Rules: rulesStub, Chunker: chunkStub}
		d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("x")))
		require.Equal(t, []string{"chunker", "rules"}, order,
			"Java: disambiguator.disambiguate(chunker.disambiguate(input)) → multiword then XML")
	}
}

// arOrderStage records Disambiguate call order for stage-order proof.
type arOrderStage struct {
	name  string
	order *[]string
}

func (s *arOrderStage) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s.order != nil {
		*s.order = append(*s.order, s.name)
	}
	return input
}

var _ interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
} = (*arOrderStage)(nil)

