package ro

// Twin of RomanianRuleDisambiguatorTest.java — XmlRuleDisambiguator(Romanian)
// with useGlobalDisambiguation=false vs DemoDisambiguator, TestTools.myAssert strings.

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
	tagro "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ro"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	tokro "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ro"
	"github.com/stretchr/testify/require"
)

// loadROXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Romanian())
// (useGlobalDisambiguation default false) over official resource/ro/disambiguation.xml.
func loadROXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := discoverRODisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ro", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverRODisambiguationXML() string {
	if p := os.Getenv("LANG_RO_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ro",
		"src", "main", "resources", "org", "languagetool", "resource", "ro", "disambiguation.xml")
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

func setupRODisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagro.DiscoverRomanianPOSDict() == "" {
		t.Skip("romanian.dict not in tree")
	}
	tagro.EnsureDefaultRomanianTagger()
	require.NotNil(t, tagro.DefaultRomanianTagger)
	require.NotNil(t, tagro.DefaultRomanianTagger.GetWordTagger())

	xml = loadROXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("ro/disambiguation.xml not in tree or failed to load")
	}
	return disambigxx.NewDemoDisambiguator(), xml
}

// Twin of RomanianRuleDisambiguatorTest.testCare1
func TestRomanianRuleDisambiguator_Care1(t *testing.T) {
	demo, xmlDisam := setupRODisambiguation(t)
	const input = "Persoana care face treabă."
	// DemoDisambiguator — full tagger readings (no XML)
	require.Equal(t,
		"/[null]SENT_START Persoana/[persoană]Sfs3aac000  /[null]null care/[car]Snp3anc000|care/[care]0000000000|care/[care]N000a0l000|care/[căra]V0p3000cz0|care/[căra]V0s3000cz0  /[null]null face/[face]V000000f00|face/[face]V0s3000iz0  /[null]null treabă/[treabă]Sfs3anc000 ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// XmlRuleDisambiguator — SUBST_CARE_VERB keeps care as relative pronoun N000a0l000
	require.Equal(t,
		"/[null]SENT_START Persoana/[persoană]Sfs3aac000  /[null]null care/[care]N000a0l000  /[null]null face/[face]V000000f00|face/[face]V0s3000iz0  /[null]null treabă/[treabă]Sfs3anc000 ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml disambiguator")
}

// Twin of RomanianRuleDisambiguatorTest.testEsteO
func TestRomanianRuleDisambiguator_EsteO(t *testing.T) {
	demo, xmlDisam := setupRODisambiguation(t)

	// DemoDisambiguator keeps ambiguous masă (verb|noun)
	require.Equal(t,
		"/[null]SENT_START este/[fi]V0s3000izb  /[null]null o/[o]Dfs3a0t000|o/[o]I00000o000|o/[o]Nfs3a0p00c|o/[o]Sms3anc000|o/[vrea]V0s3000iov  /[null]null masă/[masa]V0s3000is0|masă/[masă]Sfs3anc000 ./[null]null",
		myAssertDisambiguate("este o masă.", demo),
		"demo disambiguator with period")
	// XmlRuleDisambiguator VERB_o_SUBST keeps only noun reading on masă
	require.Equal(t,
		"/[null]SENT_START este/[fi]V0s3000izb  /[null]null o/[o]Dfs3a0t000|o/[o]I00000o000|o/[o]Nfs3a0p00c|o/[o]Sms3anc000|o/[vrea]V0s3000iov  /[null]null masă/[masă]Sfs3anc000 ./[null]null",
		myAssertDisambiguate("este o masă.", xmlDisam),
		"xml disambiguator with period")
	// Same without trailing period
	require.Equal(t,
		"/[null]SENT_START este/[fi]V0s3000izb  /[null]null o/[o]Dfs3a0t000|o/[o]I00000o000|o/[o]Nfs3a0p00c|o/[o]Sms3anc000|o/[vrea]V0s3000iov  /[null]null masă/[masă]Sfs3anc000",
		myAssertDisambiguate("este o masă", xmlDisam),
		"xml disambiguator without period")
}

// Twin of RomanianRuleDisambiguatorTest.testDezambiguizareVerb
func TestRomanianRuleDisambiguator_DezambiguizareVerb(t *testing.T) {
	demo, xmlDisam := setupRODisambiguation(t)

	// vom participa la — demo keeps both infinitive and imperfect on participa
	require.Equal(t,
		"/[null]SENT_START vom/[vrea]V0p1000ivv  /[null]null participa/[participa]V000000f00|participa/[participa]V0s3000ii0  /[null]null la/[la]P000000000|la/[la]Sms3anc000",
		myAssertDisambiguate("vom participa la", demo),
		"demo: vom participa la")
	// xml VOM_PARTICIPA_LA keeps only V000000f00
	require.Equal(t,
		"/[null]SENT_START vom/[vrea]V0p1000ivv  /[null]null participa/[participa]V000000f00  /[null]null la/[la]P000000000|la/[la]Sms3anc000",
		myAssertDisambiguate("vom participa la", xmlDisam),
		"xml: vom participa la")

	// vom culege — demo multi-reading
	require.Equal(t,
		"/[null]SENT_START vom/[vrea]V0p1000ivv  /[null]null culege/[culege]V000000f00|culege/[culege]V0s2000m00|culege/[culege]V0s3000iz0",
		myAssertDisambiguate("vom culege", demo),
		"demo: vom culege")
	// xml keeps infinitive only
	require.Equal(t,
		"/[null]SENT_START vom/[vrea]V0p1000ivv  /[null]null culege/[culege]V000000f00",
		myAssertDisambiguate("vom culege", xmlDisam),
		"xml: vom culege")
	// veți culege — only xml case in Java
	require.Equal(t,
		"/[null]SENT_START veți/[vrea]V0p2000ivv  /[null]null culege/[culege]V000000f00",
		myAssertDisambiguate("veți culege", xmlDisam),
		"xml: veți culege")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// RomanianWordTokenizer, SRXSentenceTokenizer(Romanian), RomanianTagger, disambiguator).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagro.EnsureDefaultRomanianTagger()
	tagger := tagro.DefaultRomanianTagger
	wt := tokro.NewRomanianWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("ro")
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
				// Java BaseTagger.createNullToken
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

// testToolsIsWord ports TestTools.isWord: any letter or digit → word token.
func testToolsIsWord(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// formatMyAssertSentence ports TestTools.getAsStrings + join for one sentence.
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
		// Java Collections.sort — force stable order across lexicon versions
		sort.Strings(readings)
		parts = append(parts, strings.Join(readings, "|"))
	}
	return strings.Join(parts, " ")
}

// testToolsGetAsString ports TestTools.getAsString: token/[lemma]POS with null literals.
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
