package ru

// Outcome twins for Russian XmlRuleDisambiguator as used by RussianHybridDisambiguator:
// Java new XmlRuleDisambiguator(Russian.getInstance()) with useGlobalDisambiguation=false.
// Cases derived from official resource/ru/disambiguation.xml <example type="ambiguous">
// (NumD_*_tag, VERB-KA) + active remove-rules grounded in the same pack (NOUN_V6, NOUN_V7, NOUN_R)
// + real RussianTagger readings — same bar as Danish/Arabic/Italian/RomanianRuleDisambiguatorTest.
// Note: NOUN_V5 / ADJ_V5 "Где телефон?" examples are commented out in official XML and do not fire.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	// rulesru: WireRussianFilterTagger + init registers VERB-KA NoDisambiguation filters.
	rulesru "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ru"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ru"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadRUXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(Russian.getInstance())
// (useGlobalDisambiguation default false) over official resource/ru/disambiguation.xml.
func loadRUXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := discoverRUDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ru", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverRUDisambiguationXML() string {
	if p := os.Getenv("LANG_RU_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ru",
		"src", "main", "resources", "org", "languagetool", "resource", "ru", "disambiguation.xml")
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

// wireRUPartialPosTagFilter installs RussianTagger for NoDisambiguationRussianPartialPosTagFilter
// (Java: Languages.getLanguageForShortCode("ru").getTagger()) so VERB-KA filters accept.
func wireRUPartialPosTagFilter(t *testing.T) {
	t.Helper()
	tagru.EnsureDefaultRussianTagger()
	rulesru.WireRussianFilterTaggerFromTagWord(func(token string) []languagetool.TokenTag {
		tagru.EnsureDefaultRussianTagger()
		trs := tagru.DefaultRussianTagger.Tag([]string{token})
		if len(trs) == 0 || trs[0] == nil {
			return nil
		}
		var out []languagetool.TokenTag
		for _, r := range trs[0].GetReadings() {
			if r == nil {
				continue
			}
			pos, lemma := "", ""
			if p := r.GetPOSTag(); p != nil {
				pos = *p
			}
			if l := r.GetLemma(); l != nil {
				lemma = *l
			}
			if pos != "" {
				out = append(out, languagetool.TokenTag{POS: pos, Lemma: lemma})
			}
		}
		return out
	})
	t.Cleanup(rulesru.ClearDefaultRussianPartialPosTagger)
}

func setupRUDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagru.DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	tagru.EnsureDefaultRussianTagger()
	require.NotNil(t, tagru.DefaultRussianTagger)
	require.NotNil(t, tagru.DefaultRussianTagger.GetWordTagger())
	require.NotEmpty(t, tagru.RussianPOSDictPath(), "real russian.dict must load")

	wireRUPartialPosTagFilter(t)

	xml = loadRUXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("ru/disambiguation.xml not in tree or failed to load")
	}
	// Official RU pack (with RU filters registered): ≥50 pattern rules (numeric, case, VERB-KA, …).
	require.GreaterOrEqual(t, len(xml.Rules), 50)
	return disambigxx.NewDemoDisambiguator(), xml
}

// NumD_D_tag: 73 + noun → NumD_D (official ambiguous example).
func TestRussianRuleDisambiguator_NumD_D_73Procenta(t *testing.T) {
	demo, xmlDisam := setupRUDisambiguation(t)
	const input = "73 процента"
	require.Equal(t,
		"/[null]SENT_START 73/[null]null  /[null]null процента/[процент]NN:Inanim:Masc:Sin:R",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START 73/[73]NumD_D  /[null]null процента/[процент]NN:Inanim:Masc:Sin:R",
		myAssertDisambiguate(input, xmlDisam),
		"xml NumD_D_tag")
}

// NumD_S_tag: 71 + noun → NumD_S (official ambiguous example).
func TestRussianRuleDisambiguator_NumD_S_71Procent(t *testing.T) {
	demo, xmlDisam := setupRUDisambiguation(t)
	const input = "71 процент"
	require.Equal(t,
		"/[null]SENT_START 71/[null]null  /[null]null процент/[процент]NN:Inanim:Masc:Sin:Nom|процент/[процент]NN:Inanim:Masc:Sin:V",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START 71/[71]NumD_S  /[null]null процент/[процент]NN:Inanim:Masc:Sin:Nom|процент/[процент]NN:Inanim:Masc:Sin:V",
		myAssertDisambiguate(input, xmlDisam),
		"xml NumD_S_tag")
}

// NumD_P_tag: 75 / 11 + noun → NumD_P (official ambiguous examples).
func TestRussianRuleDisambiguator_NumD_P_Procentov(t *testing.T) {
	demo, xmlDisam := setupRUDisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START 75/[null]null  /[null]null процентов/[процент]NN:Inanim:Masc:PL:R",
		myAssertDisambiguate("75 процентов", demo),
		"demo 75")
	require.Equal(t,
		"/[null]SENT_START 75/[75]NumD_P  /[null]null процентов/[процент]NN:Inanim:Masc:PL:R",
		myAssertDisambiguate("75 процентов", xmlDisam),
		"xml NumD_P_tag 75")

	require.Equal(t,
		"/[null]SENT_START 11/[null]null  /[null]null процентов/[процент]NN:Inanim:Masc:PL:R",
		myAssertDisambiguate("11 процентов", demo),
		"demo 11")
	require.Equal(t,
		"/[null]SENT_START 11/[11]NumD_P  /[null]null процентов/[процент]NN:Inanim:Masc:PL:R",
		myAssertDisambiguate("11 процентов", xmlDisam),
		"xml NumD_P_tag 11")
}

// VERB-KA: дай-ка → add VB:IMP:TRANS:PFV:Sin:P2 (official ambiguous example; needs RU tagger filter).
func TestRussianRuleDisambiguator_VerbKaDaiKa(t *testing.T) {
	demo, xmlDisam := setupRUDisambiguation(t)
	const input = "Ваня, дай-ка мне этот молоток."
	require.Equal(t,
		"/[null]SENT_START Ваня/[ваня]NN:Name:Masc:Sin:Nom ,/[null]null  /[null]null дай-ка/[null]null  /[null]null мне/[я]PNN:Sin:D:P1|мне/[я]PNN:Sin:P:P1  /[null]null этот/[этот]ADJ:MPR:Masc:Nom|этот/[этот]ADJ:MPR:Masc:V  /[null]null молоток/[молоток]NN:Inanim:Masc:Sin:Nom|молоток/[молоток]NN:Inanim:Masc:Sin:V ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// XmlRuleDisambiguator VERB-KA + NoDisambiguationRussianPartialPosTagFilter: add imperative tag.
	// Note: "этот"/"молоток" stay multi-reading here (ADJ_V5/NOUN_V5 commented out in official XML).
	require.Equal(t,
		"/[null]SENT_START Ваня/[ваня]NN:Name:Masc:Sin:Nom ,/[null]null  /[null]null дай-ка/[дай-ка]VB:IMP:TRANS:PFV:Sin:P2  /[null]null мне/[я]PNN:Sin:D:P1|мне/[я]PNN:Sin:P:P1  /[null]null этот/[этот]ADJ:MPR:Masc:Nom|этот/[этот]ADJ:MPR:Masc:V  /[null]null молоток/[молоток]NN:Inanim:Masc:Sin:Nom|молоток/[молоток]NN:Inanim:Masc:Sin:V ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml VERB-KA дай-ка")
}

// NOUN_V6: на + этот|тот… + NN Nom/V → remove Nom (active rule; grounded in official pack comment "на этот вопрос").
func TestRussianRuleDisambiguator_NounV6NaEtotVopros(t *testing.T) {
	demo, xmlDisam := setupRUDisambiguation(t)
	const input = "на этот вопрос"
	require.Equal(t,
		"/[null]SENT_START на/[на]PREP  /[null]null этот/[этот]ADJ:MPR:Masc:Nom|этот/[этот]ADJ:MPR:Masc:V  /[null]null вопрос/[вопрос]NN:Inanim:Masc:Sin:Nom|вопрос/[вопрос]NN:Inanim:Masc:Sin:V",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// NOUN_V6 removes Nom on вопрос; ADJ_V / ADJ_V2 cascade keeps only V on этот.
	require.Equal(t,
		"/[null]SENT_START на/[на]PREP  /[null]null этот/[этот]ADJ:MPR:Masc:V  /[null]null вопрос/[вопрос]NN:Inanim:Masc:Sin:V",
		myAssertDisambiguate(input, xmlDisam),
		"xml NOUN_V6 + ADJ_V*")
}

// NOUN_V7: всё|все|… + время → remove Nom (active rule; pack comment "все время").
func TestRussianRuleDisambiguator_NounV7VseVremya(t *testing.T) {
	demo, xmlDisam := setupRUDisambiguation(t)
	const input = "все время"
	require.Equal(t,
		"/[null]SENT_START все/[весь]ADJ:MPR:PL:Nom|все/[весь]ADJ:MPR:PL:V|все/[все]PNN:PL:Nom|все/[все]PNN:PL:V|все/[все]PNN:Sin:Nom|все/[все]PNN:Sin:V  /[null]null время/[время]NN:Inanim:Neut:Sin:Nom|время/[время]NN:Inanim:Neut:Sin:V",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START все/[весь]ADJ:MPR:PL:Nom|все/[весь]ADJ:MPR:PL:V|все/[все]PNN:PL:Nom|все/[все]PNN:PL:V|все/[все]PNN:Sin:Nom|все/[все]PNN:Sin:V  /[null]null время/[время]NN:Inanim:Neut:Sin:V",
		myAssertDisambiguate(input, xmlDisam),
		"xml NOUN_V7")
}

// NOUN_R: для + NN V/R → remove V (active rule; pack comment "без - родительный").
func TestRussianRuleDisambiguator_NounRDlyaDoma(t *testing.T) {
	demo, xmlDisam := setupRUDisambiguation(t)
	const input = "для дома"
	require.Equal(t,
		"/[null]SENT_START для/[длить]DPT:Real:TRANS:IMPFV|для/[для]PREP  /[null]null дома/[дом]NN:Inanim:Masc:PL:Nom|дома/[дом]NN:Inanim:Masc:PL:V|дома/[дом]NN:Inanim:Masc:Sin:R|дома/[дома]ADV",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// NOUN_R removes V; also drops PL:Nom that was paired with V in and-marker path → Sin:R + ADV remain.
	require.Equal(t,
		"/[null]SENT_START для/[длить]DPT:Real:TRANS:IMPFV|для/[для]PREP  /[null]null дома/[дом]NN:Inanim:Masc:Sin:R|дома/[дома]ADV",
		myAssertDisambiguate(input, xmlDisam),
		"xml NOUN_R")
}

// Hybrid Rules stage uses the same official XML (Java eager XmlRuleDisambiguator field).
func TestRussianHybridDisambiguator_RulesStageMatchesXml(t *testing.T) {
	_, xmlDisam := setupRUDisambiguation(t)
	hybrid := disambigru.NewRussianHybridDisambiguator()
	require.NotNil(t, hybrid.Rules, "Java constructs XmlRuleDisambiguator eagerly")
	const input = "73 процента"
	require.Equal(t,
		myAssertDisambiguate(input, xmlDisam),
		myAssertDisambiguate(input, hybrid),
		"hybrid Rules stage == standalone XmlRuleDisambiguator")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Russian), RussianTagger, disambiguator).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagru.EnsureDefaultRussianTagger()
	tagger := tagru.DefaultRussianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("ru")
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
