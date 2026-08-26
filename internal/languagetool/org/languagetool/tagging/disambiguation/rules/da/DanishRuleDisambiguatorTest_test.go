package da

// Outcome twins for Danish XmlRuleDisambiguator as used by Danish.createDefaultDisambiguator:
// Java new XmlRuleDisambiguator(this) with useGlobalDisambiguation=false.
// Cases derived from official resource/da/disambiguation.xml <example type="ambiguous">
// (sub-pron, sub-ver, ver-sub, pron-sub) + real DanishTagger readings —
// same bar as ArabicRuleDisambiguatorTest / RomanianRuleDisambiguatorTest / ItalianRuleDisambiguatorTest.

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
	tagda "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/da"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadDAXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Danish())
// (useGlobalDisambiguation default false) over official resource/da/disambiguation.xml.
func loadDAXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := discoverDADisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "da", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverDADisambiguationXML() string {
	if p := os.Getenv("LANG_DA_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "da",
		"src", "main", "resources", "org", "languagetool", "resource", "da", "disambiguation.xml")
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

func setupDADisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagda.DiscoverDanishPOSDict() == "" {
		t.Skip("danish.dict not in tree")
	}
	tagda.EnsureDefaultDanishTagger()
	require.NotNil(t, tagda.DefaultDanishTagger)
	require.NotNil(t, tagda.DefaultDanishTagger.GetWordTagger())

	xml = loadDAXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("da/disambiguation.xml not in tree or failed to load")
	}
	// Official DA pack: sub-pron (2), sub-ver (3), ver-sub (7), pron-sub (1) → 13 rules.
	require.GreaterOrEqual(t, len(xml.Rules), 13)
	return disambigxx.NewDemoDisambiguator(), xml
}

// sub-pron rule 1: article + ambiguous pron|sub → keep sub.
// XML example: Et <marker>jeg</marker> er interresant.
func TestDanishRuleDisambiguator_SubPronEtJeg(t *testing.T) {
	demo, xmlDisam := setupDADisambiguation(t)
	const input = "Et jeg er interresant."
	require.Equal(t,
		"/[null]SENT_START Et/[et]art  /[null]null jeg/[jeg]pron:sin:nom|jeg/[jeg]sub:ube:sin:neu:nom  /[null]null er/[være]ver:præ:akt  /[null]null interresant/[null]null ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Et/[et]art  /[null]null jeg/[jeg]sub:ube:sin:neu:nom  /[null]null er/[være]ver:præ:akt  /[null]null interresant/[null]null ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml sub-pron Et jeg")
}

// sub-pron rule 2: pure pronoun + ambiguous pron|sub → keep sub.
// XML example: Mig <marker>jeg</marker> er meget interessant.
func TestDanishRuleDisambiguator_SubPronMigJeg(t *testing.T) {
	demo, xmlDisam := setupDADisambiguation(t)
	const input = "Mig jeg er meget interessant."
	require.Equal(t,
		"/[null]SENT_START Mig/[jeg]pron:sin:akk  /[null]null jeg/[jeg]pron:sin:nom|jeg/[jeg]sub:ube:sin:neu:nom  /[null]null er/[være]ver:præ:akt  /[null]null meget/[meget]adv  /[null]null interessant/[null]null ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Mig/[jeg]pron:sin:akk  /[null]null jeg/[jeg]sub:ube:sin:neu:nom  /[null]null er/[være]ver:præ:akt  /[null]null meget/[meget]adv  /[null]null interessant/[null]null ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml sub-pron Mig jeg")
}

// sub-ver rule 1: article + ambiguous sub|ver + non-passive ver → keep sub.
// XML example: En <marker>skal</marker> er spist af en fugl.
func TestDanishRuleDisambiguator_SubVerEnSkal(t *testing.T) {
	demo, xmlDisam := setupDADisambiguation(t)
	const input = "En skal er spist af en fugl."
	require.Equal(t,
		"/[null]SENT_START En/[en]art  /[null]null skal/[skal]sub:ube:sin:utr:nom|skal/[skulle]ver:præ:akt  /[null]null er/[være]ver:præ:akt  /[null]null spist/[spise]ver:kor:akt  /[null]null af/[af]adv|af/[af]pra  /[null]null en/[en]art  /[null]null fugl/[fugl]sub:ube:sin:utr:nom ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START En/[en]art  /[null]null skal/[skal]sub:ube:sin:utr:nom  /[null]null er/[være]ver:præ:akt  /[null]null spist/[spise]ver:kor:akt  /[null]null af/[af]adv|af/[af]pra  /[null]null en/[en]art  /[null]null fugl/[fugl]sub:ube:sin:utr:nom ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml sub-ver En skal")
}

// ver-sub rule 3: pronoun-like subject + ambiguous sub|ver → keep ver.
// XML example: Vi <marker>skal</marker> fiske.
func TestDanishRuleDisambiguator_VerSubViSkal(t *testing.T) {
	demo, xmlDisam := setupDADisambiguation(t)
	const input = "Vi skal fiske."
	require.Equal(t,
		"/[null]SENT_START Vi/[Vi]pro:nom|Vi/[vi]pron:plu:nom|Vi/[vi]sub:ube:sin:neu:nom|Vi/[vi]ver:imp:akt|Vi/[vi]ver:inf:akt  /[null]null skal/[skal]sub:ube:sin:utr:nom|skal/[skulle]ver:præ:akt  /[null]null fiske/[fiske]ver:inf:akt ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Vi/[Vi]pro:nom|Vi/[vi]pron:plu:nom|Vi/[vi]sub:ube:sin:neu:nom|Vi/[vi]ver:imp:akt|Vi/[vi]ver:inf:akt  /[null]null skal/[skulle]ver:præ:akt  /[null]null fiske/[fiske]ver:inf:akt ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml ver-sub Vi skal")
}

// pron-sub rule 1: sentence-initial ambiguous Jeg + ver → keep pronoun.
// XML example: <marker>Jeg</marker> er mig!
func TestDanishRuleDisambiguator_PronSubJegErMig(t *testing.T) {
	demo, xmlDisam := setupDADisambiguation(t)
	const input = "Jeg er mig!"
	require.Equal(t,
		"/[null]SENT_START Jeg/[jeg]pron:sin:nom|Jeg/[jeg]sub:ube:sin:neu:nom  /[null]null er/[være]ver:præ:akt  /[null]null mig/[jeg]pron:sin:akk !/[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Jeg/[jeg]pron:sin:nom  /[null]null er/[være]ver:præ:akt  /[null]null mig/[jeg]pron:sin:akk !/[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml pron-sub Jeg er mig")
}

// Cascade: sub-ver filters first "skal" to noun; ver-sub filters second "skal" to verb.
// XML examples: En skal <marker>skal</marker> noget. / En <marker>skal</marker> er …
func TestDanishRuleDisambiguator_EnSkalSkalNoget(t *testing.T) {
	demo, xmlDisam := setupDADisambiguation(t)
	const input = "En skal skal noget."
	require.Equal(t,
		"/[null]SENT_START En/[en]art  /[null]null skal/[skal]sub:ube:sin:utr:nom|skal/[skulle]ver:præ:akt  /[null]null skal/[skal]sub:ube:sin:utr:nom|skal/[skulle]ver:præ:akt  /[null]null noget/[null]null ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START En/[en]art  /[null]null skal/[skal]sub:ube:sin:utr:nom  /[null]null skal/[skulle]ver:præ:akt  /[null]null noget/[null]null ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml cascade En skal skal noget")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Danish), DanishTagger, disambiguator).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagda.EnsureDefaultDanishTagger()
	tagger := tagda.DefaultDanishTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("da")
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
