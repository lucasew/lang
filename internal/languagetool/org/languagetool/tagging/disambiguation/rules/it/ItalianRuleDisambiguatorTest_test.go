package it

// Twin of ItalianRuleDisambiguator / XmlRuleDisambiguator(Italian) outcomes.
// Java has no dedicated ItalianRuleDisambiguatorTest with myAssert strings;
// cases are derived from official resource/it/disambiguation.xml (UNIFY_ADJ_NOUN,
// IO_VERB) + real ItalianTagger readings, same bar as RomanianRuleDisambiguatorTest.

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
	tagit "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/it"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadITXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Italian())
// (useGlobalDisambiguation default false) over official resource/it/disambiguation.xml.
func loadITXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := discoverITDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "it", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverITDisambiguationXML() string {
	if p := os.Getenv("LANG_IT_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "it",
		"src", "main", "resources", "org", "languagetool", "resource", "it", "disambiguation.xml")
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

func setupITDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagit.DiscoverItalianPOSDict() == "" {
		t.Skip("italian.dict not in tree")
	}
	tagit.EnsureDefaultItalianTagger()
	require.NotNil(t, tagit.DefaultItalianTagger)
	require.NotNil(t, tagit.DefaultItalianTagger.GetWordTagger())

	xml = loadITXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("it/disambiguation.xml not in tree or failed to load")
	}
	// Official IT pack: UNIFY_ADJ_NOUN + IO_VERB (2 rules in group) → 3 pattern rules.
	require.GreaterOrEqual(t, len(xml.Rules), 2)
	require.NotNil(t, xml.UnifierConfig, "unification tables from it/disambiguation.xml")
	return disambigxx.NewDemoDisambiguator(), xml
}

// IO_VERB: second-person subject filters non-2s readings on the verb.
func TestItalianRuleDisambiguator_IoVerbTuAmi(t *testing.T) {
	demo, xmlDisam := setupITDisambiguation(t)
	const input = "tu ami"
	require.Equal(t,
		"/[null]SENT_START tu/[tu]PRO-PERS-2-F-S|tu/[tu]PRO-PERS-2-M-S  /[null]null ami/[amare]VER:ind+pres+2+s|ami/[amare]VER:sub+pres+1+s|ami/[amare]VER:sub+pres+2+s|ami/[amare]VER:sub+pres+3+s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// XmlRuleDisambiguator IO_VERB: keep only 2nd-person singular verb readings
	require.Equal(t,
		"/[null]SENT_START tu/[tu]PRO-PERS-2-F-S|tu/[tu]PRO-PERS-2-M-S  /[null]null ami/[amare]VER:ind+pres+2+s|ami/[amare]VER:sub+pres+2+s",
		myAssertDisambiguate(input, xmlDisam),
		"xml disambiguator IO_VERB")
}

// IO_VERB: 3rd person singular filters imperative 2s reading on parla.
func TestItalianRuleDisambiguator_IoVerbLuiParla(t *testing.T) {
	demo, xmlDisam := setupITDisambiguation(t)
	const input = "lui parla"
	require.Equal(t,
		"/[null]SENT_START lui/[lui]PRO-PERS-3-M-S  /[null]null parla/[parlare]VER:impr+pres+2+s|parla/[parlare]VER:ind+pres+3+s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START lui/[lui]PRO-PERS-3-M-S  /[null]null parla/[parlare]VER:ind+pres+3+s",
		myAssertDisambiguate(input, xmlDisam),
		"xml disambiguator IO_VERB")
}

// UNIFY_ADJ_NOUN: gender/number agreement keeps only masculine singular ADJ.
func TestItalianRuleDisambiguator_UnifyAdjNounImpossibile(t *testing.T) {
	demo, xmlDisam := setupITDisambiguation(t)
	const input = "impossibile desiderio"
	require.Equal(t,
		"/[null]SENT_START impossibile/[impossibile]ADJ:pos+f+s|impossibile/[impossibile]ADJ:pos+m+s  /[null]null desiderio/[desiderio]NOUN-M:s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START impossibile/[impossibile]ADJ:pos+m+s  /[null]null desiderio/[desiderio]NOUN-M:s",
		myAssertDisambiguate(input, xmlDisam),
		"xml disambiguator UNIFY_ADJ_NOUN")
}

// UNIFY_ADJ_NOUN: drops non-ADJ readings on grande and feminine ADJ vs masculine noun.
func TestItalianRuleDisambiguator_UnifyAdjNounGrandeUomo(t *testing.T) {
	demo, xmlDisam := setupITDisambiguation(t)
	const input = "grande uomo"
	require.Equal(t,
		"/[null]SENT_START grande/[grande]ADJ:pos+f+s|grande/[grande]ADJ:pos+m+s|grande/[grande]NOUN-F:s|grande/[grande]NOUN-M:s  /[null]null uomo/[uomo]NOUN-M:s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START grande/[grande]ADJ:pos+m+s  /[null]null uomo/[uomo]NOUN-M:s",
		myAssertDisambiguate(input, xmlDisam),
		"xml disambiguator UNIFY_ADJ_NOUN")
}

// UNIFY_ADJ_NOUN: ADJ+NOUN both multi-reading — keep agreeing gender/number only.
func TestItalianRuleDisambiguator_UnifyAdjNounNuovoLibro(t *testing.T) {
	demo, xmlDisam := setupITDisambiguation(t)
	const input = "nuovo libro"
	require.Equal(t,
		"/[null]SENT_START nuovo/[nuovo]ADJ:pos+m+s|nuovo/[nuovo]NOUN-M:s  /[null]null libro/[librare]VER:ind+pres+1+s|libro/[libro]NOUN-M:s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START nuovo/[nuovo]ADJ:pos+m+s  /[null]null libro/[libro]NOUN-M:s",
		myAssertDisambiguate(input, xmlDisam),
		"xml disambiguator UNIFY_ADJ_NOUN")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Italian), ItalianTagger, disambiguator).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagit.EnsureDefaultItalianTagger()
	tagger := tagit.DefaultItalianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("it")
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
