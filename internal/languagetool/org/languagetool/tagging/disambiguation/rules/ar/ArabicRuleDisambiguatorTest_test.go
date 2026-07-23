package ar

// Outcome twins for Arabic XmlRuleDisambiguator as used by ArabicHybridDisambiguator:
// Java new XmlRuleDisambiguator(new Arabic()) with useGlobalDisambiguation=false.
// Cases derived from official resource/ar/disambiguation.xml <example type="ambiguous">
// (Keep_Only_verbs_*, Keep_Only_Nouns_after_Jar, Numeric_phrase_tags*) + real ArabicTagger
// readings — same bar as RomanianRuleDisambiguatorTest / ItalianRuleDisambiguatorTest.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	tagar "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadARXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Arabic())
// (useGlobalDisambiguation default false) over official resource/ar/disambiguation.xml.
func loadARXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := discoverARDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ar", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverARDisambiguationXML() string {
	if p := os.Getenv("LANG_AR_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ar",
		"src", "main", "resources", "org", "languagetool", "resource", "ar", "disambiguation.xml")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func setupARDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagar.DiscoverArabicPOSDict() == "" {
		t.Skip("arabic.dict not in tree")
	}
	tagar.EnsureDefaultArabicTagger()
	require.NotNil(t, tagar.DefaultArabicTagger)
	require.NotNil(t, tagar.DefaultArabicTagger.GetWordTagger())

	xml = loadARXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("ar/disambiguation.xml not in tree or failed to load")
	}
	// Official AR pack: 5 rules (Keep_Only_verbs_*, Keep_Only_Nouns_after_Jar, Numeric_phrase_tags*)
	require.GreaterOrEqual(t, len(xml.Rules), 5)
	require.NotNil(t, xml.UnifierConfig, "unification tables from ar/disambiguation.xml")
	return disambigxx.NewDemoDisambiguator(), xml
}

// Keep_Only_verbs_after_some_tools: قد + عامل → drop noun readings on عامل.
// XML example grounded: inputform multi (N+V) → outputform V only.
func TestArabicRuleDisambiguator_KeepOnlyVerbsQadAamil(t *testing.T) {
	demo, xmlDisam := setupARDisambiguation(t)
	const input = "قد عامل"
	require.Equal(t,
		"/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---|عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START قد/[قد]NJ-;M1--;---|قد/[قد]NJ-;M1A-;---|قد/[قد]NJ-;M1I-;---|قد/[قد]NJ-;M1U-;---|قد/[قد]NM-;M1--;---|قد/[قد]NM-;M1A-;---|قد/[قد]NM-;M1I-;---|قد/[قد]NM-;M1U-;---|قد/[قَادَ]VW1;M1Y-i--;---|قد/[قَدَّ]V31;M1H-pa-;---|قد/[قَدَّ]V31;M1H-pp-;---  /[null]null عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---",
		myAssertDisambiguate(input, xmlDisam),
		"xml Keep_Only_verbs_after_some_tools")
}

// Keep_Only_Nouns_after_Jar: في + عامل → drop verb readings on عامل.
// XML example: outputform noun-only NA- readings.
func TestArabicRuleDisambiguator_KeepOnlyNounsAfterJarFiAamil(t *testing.T) {
	demo, xmlDisam := setupARDisambiguation(t)
	const input = "في عامل"
	require.Equal(t,
		"/[null]SENT_START في/[في]PR-;---;---|في/[في]PRD;---;---|في/[وَفَى]VW1;F1Y-i--;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---|عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START في/[في]PR-;---;---|في/[في]PRD;---;---|في/[وَفَى]VW1;F1Y-i--;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---",
		myAssertDisambiguate(input, xmlDisam),
		"xml Keep_Only_Nouns_after_Jar")
}

// Keep_Only_verbs_after_some_tools_1: بعد أن + عامل → verb-only on عامل.
func TestArabicRuleDisambiguator_KeepOnlyVerbsAfterBaadAnAamil(t *testing.T) {
	demo, xmlDisam := setupARDisambiguation(t)
	const input = "بعد أن عامل"
	require.Equal(t,
		"/[null]SENT_START بعد/[بعد]NJ-;M1--;---|بعد/[بعد]NJ-;M1A-;---|بعد/[بعد]NJ-;M1I-;---|بعد/[بعد]NJ-;M1U-;---|بعد/[بَعُدَ]V30;M1H-pa-;---|بعد/[بَعُدَ]V30;M1H-pp-;---|بعد/[بَعِدَ]V30;M1H-pa-;---|بعد/[بَعِدَ]V30;M1H-pp-;---|بعد/[بَعَّدَ]V41;M1H-pa-;---|بعد/[بَعَّدَ]V41;M1H-pp-;---|بعد/[بَعَّدَ]V41;M1Y-i--;---|بعد/[عد]NJ-;M1--;-B-|بعد/[عد]NJ-;M1I-;-B-|بعد/[عد]NM-;M1--;-B-|بعد/[عد]NM-;M1I-;-B-  /[null]null أن/[آنَ]V-0;F3H-pa-;---|أن/[آنَ]V-0;F3H-pp-;---|أن/[آنَ]V-0;F3Y-i--;---|أن/[آنَ]V-0;M1Y-i--;---|أن/[أَنَّ]P--;----;---|أن/[وَأَى]VW1;M3Y-i--;---|أن/[وَنَى]VW1;M1I-fa0;---  /[null]null عامل/[عامل]NA-;M1--;---|عامل/[عامل]NA-;M1A-;---|عامل/[عامل]NA-;M1I-;---|عامل/[عامل]NA-;M1U-;---|عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START بعد/[بعد]NJ-;M1--;---|بعد/[بعد]NJ-;M1A-;---|بعد/[بعد]NJ-;M1I-;---|بعد/[بعد]NJ-;M1U-;---|بعد/[بَعُدَ]V30;M1H-pa-;---|بعد/[بَعُدَ]V30;M1H-pp-;---|بعد/[بَعِدَ]V30;M1H-pa-;---|بعد/[بَعِدَ]V30;M1H-pp-;---|بعد/[بَعَّدَ]V41;M1H-pa-;---|بعد/[بَعَّدَ]V41;M1H-pp-;---|بعد/[بَعَّدَ]V41;M1Y-i--;---|بعد/[عد]NJ-;M1--;-B-|بعد/[عد]NJ-;M1I-;-B-|بعد/[عد]NM-;M1--;-B-|بعد/[عد]NM-;M1I-;-B-  /[null]null أن/[آنَ]V-0;F3H-pa-;---|أن/[آنَ]V-0;F3H-pp-;---|أن/[آنَ]V-0;F3Y-i--;---|أن/[آنَ]V-0;M1Y-i--;---|أن/[أَنَّ]P--;----;---|أن/[وَأَى]VW1;M3Y-i--;---|أن/[وَنَى]VW1;M1I-fa0;---  /[null]null عامل/[عَامَلَ]V41;M1H-pa-;---|عامل/[عَامَلَ]V41;M1Y-i--;---",
		myAssertDisambiguate(input, xmlDisam),
		"xml Keep_Only_verbs_after_some_tools_1")
}

// Numeric_phrase_tags (+ Numeric_phrase_tags2 cascade): ثلاثة وثلاثون → NN.* only.
// XML example: وثلاثون NND only.
func TestArabicRuleDisambiguator_NumericPhraseThalathaWathalathun(t *testing.T) {
	demo, xmlDisam := setupARDisambiguation(t)
	const input = "ثلاثة وثلاثون"
	require.Equal(t,
		"/[null]SENT_START ثلاثة/[ثلاث]NJ-;F1--;---|ثلاثة/[ثلاث]NJ-;F1A-;---|ثلاثة/[ثلاث]NJ-;F1I-;---|ثلاثة/[ثلاث]NJ-;F1U-;---|ثلاثة/[ثلاثة]NJ-;-1--;---|ثلاثة/[ثلاثة]NJ-;-1A-;---|ثلاثة/[ثلاثة]NJ-;-1I-;---|ثلاثة/[ثلاثة]NJ-;-1U-;---|ثلاثة/[ثلاثة]NNU;M3--;---  /[null]null وثلاثون/[ثلاث]NJ-;-3--;W--|وثلاثون/[ثلاث]NJ-;M3--;W--|وثلاثون/[ثلاثون]NND;-3U-;W--",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ثلاثة/[ثلاثة]NNU;M3--;---  /[null]null وثلاثون/[ثلاثون]NND;-3U-;W--",
		myAssertDisambiguate(input, xmlDisam),
		"xml Numeric_phrase_tags")
}

// Numeric_phrase_tags2: ثلاثون ألف → NN.* on first token (and second via Numeric_phrase_tags).
// XML example: ثلاثون NND only.
func TestArabicRuleDisambiguator_NumericPhraseThalathunAlf(t *testing.T) {
	demo, xmlDisam := setupARDisambiguation(t)
	const input = "ثلاثون ألف"
	require.Equal(t,
		"/[null]SENT_START ثلاثون/[ثلاث]NJ-;-3--;---|ثلاثون/[ثلاث]NJ-;M3--;---|ثلاثون/[ثلاثون]NND;-3U-;---  /[null]null ألف/[ألف]NA-;M3--;---|ألف/[ألف]NA-;M3A-;---|ألف/[ألف]NA-;M3I-;---|ألف/[ألف]NA-;M3U-;---|ألف/[ألف]NJ-;M1--;---|ألف/[ألف]NJ-;M1A-;---|ألف/[ألف]NJ-;M1I-;---|ألف/[ألف]NJ-;M1U-;---|ألف/[ألف]NM-;M1--;---|ألف/[ألف]NM-;M1A-;---|ألف/[ألف]NM-;M1I-;---|ألف/[ألف]NM-;M1U-;---|ألف/[ألف]NNH;-1--;---|ألف/[أَلَفَ]V31;M1H-pa-;---|ألف/[أَلَفَ]V31;M1H-pp-;---|ألف/[أَلَفَّ]V41;M1H-pa-;---|ألف/[أَلَفَّ]V41;M1H-pp-;---|ألف/[أَلَفَّ]V41;M1I-faA;---|ألف/[أَلَفَّ]V41;M1I-faU;---|ألف/[أَلَفَّ]V41;M1I-fpA;---|ألف/[أَلَفَّ]V41;M1I-fpU;---|ألف/[أَلِفَ]V31;M1H-pa-;---|ألف/[أَلِفَ]V31;M1H-pp-;---|ألف/[أَلَّفَ]V41;M1H-pa-;---|ألف/[أَلَّفَ]V41;M1H-pp-;---|ألف/[أَلَّفَ]V41;M1Y-i--;---|ألف/[أَلْفَى]VW1;M1I-fa0;---|ألف/[أَلْفَى]VW1;M1I-fp0;---|ألف/[أَلْفَى]VW1;M1Y-i--;---|ألف/[لَافَ]VW1;M1I-fa0;---|ألف/[لَافَ]VW1;M1I-fp0;---|ألف/[لَفَا]VW1;M1I-fa0;---|ألف/[لَفَا]VW1;M1I-fp0;---|ألف/[لَفَّ]V31;M1I-faA;---|ألف/[لَفَّ]V31;M1I-faU;---|ألف/[لَفَّ]V31;M1I-fpA;---|ألف/[لَفَّ]V31;M1I-fpU;---|ألف/[وَلَفَ]VW0;M1I-fa0;---|ألف/[وَلَفَ]VW0;M1I-faA;---|ألف/[وَلَفَ]VW0;M1I-faU;---",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ثلاثون/[ثلاثون]NND;-3U-;---  /[null]null ألف/[ألف]NNH;-1--;---",
		myAssertDisambiguate(input, xmlDisam),
		"xml Numeric_phrase_tags2")
}

// Hybrid Rules stage uses the same official XML (Java eager XmlRuleDisambiguator field).
func TestArabicHybridDisambiguator_RulesStageMatchesXml(t *testing.T) {
	_, xmlDisam := setupARDisambiguation(t)
	hybrid := tagar.NewArabicHybridDisambiguator()
	require.NotNil(t, hybrid.Rules, "Java constructs XmlRuleDisambiguator eagerly")
	const input = "قد عامل"
	require.Equal(t,
		myAssertDisambiguate(input, xmlDisam),
		myAssertDisambiguate(input, hybrid),
		"hybrid Rules stage == standalone XmlRuleDisambiguator")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Arabic), ArabicTagger, disambiguator).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagar.EnsureDefaultArabicTagger()
	tagger := tagar.DefaultArabicTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("ar")
	var out strings.Builder
	for _, sentence := range st.Tokenize(input) {
		tokens := wt.Tokenize(sentence)
		var noWS []string
		for _, tok := range tokens {
			if testToolsIsWord(tok) {
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
			if testToolsIsWord(tokenStr) {
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
		finalSentence := languagetool.NewAnalyzedSentence(tokenArray)
		if dis != nil {
			finalSentence = dis.Disambiguate(finalSentence)
		}
		out.WriteString(formatMyAssertSentence(finalSentence))
	}
	return out.String()
}

func testToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func formatMyAssertSentence(sent *languagetool.AnalyzedSentence) string {
	if sent == nil {
		return ""
	}
	var parts []string
	for _, tr := range sent.GetTokens() {
		var readings []string
		for _, r := range tr.GetReadings() {
			if r != nil {
				readings = append(readings, testToolsGetAsString(r))
			}
		}
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}

func testToolsGetAsString(tok *languagetool.AnalyzedToken) string {
	lemma, pos := "null", "null"
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	return tok.GetToken() + "/[" + lemma + "]" + pos
}
