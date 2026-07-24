package eo

// Outcome twins for Esperanto XmlRuleDisambiguator as used by Esperanto.createDefaultDisambiguator:
// Java new XmlRuleDisambiguator(this) with useGlobalDisambiguation=false.
// Cases derived from official resource/eo/disambiguation.xml rule patterns
// (NEDIREKTA_OBJEKTO, DEM_KRI, VIVI) + real EsperantoTagger readings —
// same bar as Breton/Danish RuleDisambiguator tests.

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
	tageo "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/eo"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	eotok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/eo"
	"github.com/stretchr/testify/require"
)

// loadEOXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Esperanto())
// (useGlobalDisambiguation default false) over official resource/eo/disambiguation.xml.
func loadEOXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	// Prefer process cache (tagging/eo loader); fall back to discover for isolation.
	if x := tageo.EsperantoXmlRuleDisambiguator(); x != nil {
		return x
	}
	p := discoverEODisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "eo", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverEODisambiguationXML() string {
	if p := os.Getenv("LANG_EO_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "eo",
		"src", "main", "resources", "org", "languagetool", "resource", "eo", "disambiguation.xml")
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

func setupEODisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if !tageo.EOResourcesAvailable() {
		t.Skip("official EO tagger resources not in tree")
	}
	// Smoke that tagger can initialize with real resources.
	tagger := tageo.NewEsperantoTagger()
	require.NotNil(t, tagger)
	sample := tagger.Tag([]string{"domo"})
	require.NotEmpty(t, sample)
	require.NotNil(t, sample[0].GetAnalyzedToken(0).GetPOSTag(), "domo must get O tag from real EsperantoTagger")

	xml = loadEOXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("eo/disambiguation.xml not in tree or failed to load")
	}
	// Official EO pack: NEDIREKTA_OBJEKTO (4) + DEM_KRI (1) + VIVI (2) → 7 rules.
	require.GreaterOrEqual(t, len(xml.Rules), 7)
	return disambigxx.NewDemoDisambiguator(), xml
}

// NEDIREKTA_OBJEKTO rule 1: prep + la + -o word → add X ndo on object.
func TestEsperantoRuleDisambiguator_NedirektaObjektoEnLaDomo(t *testing.T) {
	demo, xmlDisam := setupEODisambiguation(t)
	const input = "en la domo"
	require.Equal(t,
		"/[null]SENT_START en/[en]P kak  /[null]null la/[la]D  /[null]null domo/[domo]O nak np",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START en/[en]P kak  /[null]null la/[la]D  /[null]null domo/[domo]O nak np|domo/[domo]X ndo",
		myAssertDisambiguate(input, xmlDisam),
		"xml NEDIREKTA_OBJEKTO en la domo")
}

// NEDIREKTA_OBJEKTO rule 2: prep + -o word (no la) → add X ndo on object.
func TestEsperantoRuleDisambiguator_NedirektaObjektoSurTablo(t *testing.T) {
	demo, xmlDisam := setupEODisambiguation(t)
	const input = "sur tablo"
	require.Equal(t,
		"/[null]SENT_START sur/[sur]P kak  /[null]null tablo/[tablo]O nak np",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START sur/[sur]P kak  /[null]null tablo/[tablo]O nak np|tablo/[tablo]X ndo",
		myAssertDisambiguate(input, xmlDisam),
		"xml NEDIREKTA_OBJEKTO sur tablo")
}

// NEDIREKTA_OBJEKTO rule 3: prep + A/O + O/A → add X ndo on last marker token.
func TestEsperantoRuleDisambiguator_NedirektaObjektoEnBelaDomo(t *testing.T) {
	demo, xmlDisam := setupEODisambiguation(t)
	const input = "en bela domo"
	require.Equal(t,
		"/[null]SENT_START en/[en]P kak  /[null]null bela/[bela]A nak np  /[null]null domo/[domo]O nak np",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START en/[en]P kak  /[null]null bela/[bela]A nak np  /[null]null domo/[domo]O nak np|domo/[domo]X ndo",
		myAssertDisambiguate(input, xmlDisam),
		"xml NEDIREKTA_OBJEKTO en bela domo")
}

// NEDIREKTA_OBJEKTO rule 4: prep + la + A/O + O/A → add X ndo on last marker token.
func TestEsperantoRuleDisambiguator_NedirektaObjektoAlLaBelaDomo(t *testing.T) {
	demo, xmlDisam := setupEODisambiguation(t)
	const input = "al la bela domo"
	require.Equal(t,
		"/[null]SENT_START al/[al]P sak  /[null]null la/[la]D  /[null]null bela/[bela]A nak np  /[null]null domo/[domo]O nak np",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START al/[al]P sak  /[null]null la/[la]D  /[null]null bela/[bela]A nak np  /[null]null domo/[domo]O nak np|domo/[domo]X ndo",
		myAssertDisambiguate(input, xmlDisam),
		"xml NEDIREKTA_OBJEKTO al la bela domo")
}

// DEM_KRI: ? then ! → add X demkri on ! (lemma falls back to surface "!").
func TestEsperantoRuleDisambiguator_DemKri(t *testing.T) {
	demo, xmlDisam := setupEODisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START ?/[null]null !/[null]null",
		myAssertDisambiguate("?!", demo),
		"demo ?!")
	require.Equal(t,
		"/[null]SENT_START ?/[null]null !/[!]X demkri",
		myAssertDisambiguate("?!", xmlDisam),
		"xml DEM_KRI ?!")

	require.Equal(t,
		"/[null]SENT_START Kio/[null]T nak np k o ?/[null]null !/[null]null",
		myAssertDisambiguate("Kio?!", demo),
		"demo Kio?!")
	require.Equal(t,
		"/[null]SENT_START Kio/[null]T nak np k o ?/[null]null !/[!]X demkri",
		myAssertDisambiguate("Kio?!", xmlDisam),
		"xml DEM_KRI Kio?!")
}

// VIVI rule 1: viv(e[tg])?i + viv(e[gt])?on → replace V nt → V tr (lemma kept via Match filter).
func TestEsperantoRuleDisambiguator_ViviVivon(t *testing.T) {
	demo, xmlDisam := setupEODisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START vivi/[vivi]V nt i  /[null]null vivon/[vivo]O akz np",
		myAssertDisambiguate("vivi vivon", demo),
		"demo vivi vivon")
	require.Equal(t,
		"/[null]SENT_START vivi/[vivi]V tr i  /[null]null vivon/[vivo]O akz np",
		myAssertDisambiguate("vivi vivon", xmlDisam),
		"xml VIVI vivi vivon")

	require.Equal(t,
		"/[null]SENT_START vivas/[vivi]V nt as  /[null]null vivon/[vivo]O akz np",
		myAssertDisambiguate("vivas vivon", demo),
		"demo vivas vivon")
	require.Equal(t,
		"/[null]SENT_START vivas/[vivi]V tr as  /[null]null vivon/[vivo]O akz np",
		myAssertDisambiguate("vivas vivon", xmlDisam),
		"xml VIVI vivas vivon")

	// viveti / viveton: suffix forms still match viv(?:e[tg])?i / viv(?:e[gt])?on.
	require.Equal(t,
		"/[null]SENT_START viveti/[viveti]V nt i  /[null]null viveton/[viveto]O akz np",
		myAssertDisambiguate("viveti viveton", demo),
		"demo viveti viveton")
	require.Equal(t,
		"/[null]SENT_START viveti/[viveti]V tr i  /[null]null viveton/[viveto]O akz np",
		myAssertDisambiguate("viveti viveton", xmlDisam),
		"xml VIVI viveti viveton")
}

// VIVI rule 2: verb + A akz + vivon → replace V nt → V tr.
func TestEsperantoRuleDisambiguator_VivasBelanVivon(t *testing.T) {
	demo, xmlDisam := setupEODisambiguation(t)
	const input = "vivas belan vivon"
	require.Equal(t,
		"/[null]SENT_START vivas/[vivi]V nt as  /[null]null belan/[bela]A akz np  /[null]null vivon/[vivo]O akz np",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START vivas/[vivi]V tr as  /[null]null belan/[bela]A akz np  /[null]null vivon/[vivo]O akz np",
		myAssertDisambiguate(input, xmlDisam),
		"xml VIVI vivas belan vivon")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// EsperantoWordTokenizer, SRXSentenceTokenizer(Esperanto), EsperantoTagger, disambiguator).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagger := tageo.NewEsperantoTagger()
	wt := eotok.NewEsperantoWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("eo")
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
