package br

// Outcome twins for Breton XmlRuleDisambiguator as used by Breton.createDefaultDisambiguator:
// Java new XmlRuleDisambiguator(this) with useGlobalDisambiguation=false.
// Cases derived from official resource/br/disambiguation.xml rule patterns
// (EN_UR, XXI, PREP_A, O, D_O, EZ_AN, EN_EM, EZ_A, MA, UR_N, PAOT_MAT, KOZH, RA_*, PE_INT, GANT)
// + real BretonTagger / breton.dict readings — same bar as Danish/Polish RuleDisambiguator tests.

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
	tagbr "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/br"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
	"github.com/stretchr/testify/require"
)

// loadBRXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Breton())
// (useGlobalDisambiguation default false) over official resource/br/disambiguation.xml.
func loadBRXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	// Prefer process cache (tagging/br loader); fall back to discover for isolation.
	if x := tagbr.BretonXmlRuleDisambiguator(); x != nil {
		return x
	}
	p := discoverBRDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "br", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverBRDisambiguationXML() string {
	if p := os.Getenv("LANG_BR_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "br",
		"src", "main", "resources", "org", "languagetool", "resource", "br", "disambiguation.xml")
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

func setupBRDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagbr.DiscoverBretonPOSDict() == "" {
		t.Skip("breton.dict not in tree")
	}
	tagbr.EnsureDefaultBretonTagger()
	require.NotNil(t, tagbr.DefaultBretonTagger)
	require.NotNil(t, tagbr.DefaultBretonTagger.GetWordTagger())
	require.NotEmpty(t, tagbr.BretonPOSDictPath(), "real breton.dict must load")

	xml = loadBRXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("br/disambiguation.xml not in tree or failed to load")
	}
	// Official BR pack: 75 rules across EN_UR, XXI, PREP_A, O, NA, …
	require.GreaterOrEqual(t, len(xml.Rules), 70)
	return disambigxx.NewDemoDisambiguator(), xml
}

// EN_UR: en ur + V inf → ur gets X EN_UR.
func TestBretonRuleDisambiguator_EnUrDont(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "en ur dont"
	require.Equal(t,
		"/[null]SENT_START en/[e]P  /[null]null ur/[un]D e sp  /[null]null dont/[dont]V inf",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START en/[e]P  /[null]null ur/[un]X EN_UR  /[null]null dont/[dont]V inf",
		myAssertDisambiguate(input, xmlDisam),
		"xml EN_UR en ur dont")
}

// XXI Roman numbers: bare Roman → K e sp; ordinal forms → K e sp o.
func TestBretonRuleDisambiguator_XXIRoman(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START XXI/[null]null",
		myAssertDisambiguate("XXI", demo),
		"demo XXI")
	require.Equal(t,
		"/[null]SENT_START XXI/[XXI]K e sp",
		myAssertDisambiguate("XXI", xmlDisam),
		"xml XXI")

	require.Equal(t,
		"/[null]SENT_START XXI-vet/[null]null",
		myAssertDisambiguate("XXI-vet", demo),
		"demo XXI-vet")
	require.Equal(t,
		"/[null]SENT_START XXI-vet/[XXI-vet]K e sp o",
		myAssertDisambiguate("XXI-vet", xmlDisam),
		"xml XXI-vet")

	require.Equal(t,
		"/[null]SENT_START Iañ/[null]null",
		myAssertDisambiguate("Iañ", demo),
		"demo Iañ")
	require.Equal(t,
		"/[null]SENT_START Iañ/[Iañ]K e sp o",
		myAssertDisambiguate("Iañ", xmlDisam),
		"xml Iañ")
}

// PREP_A: kalz a + N → a gets P.
func TestBretonRuleDisambiguator_PrepAKalz(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "kalz a dud"
	require.Equal(t,
		"/[null]SENT_START kalz/[kalz]A|kalz/[kalz]N m s|kalz/[kalz]P  /[null]null a/[a]L a|a/[a]N m sp|a/[a]P|a/[monet]V impe 2 s|a/[monet]V pres 3 s|a/[mont]V impe 2 s|a/[mont]V pres 3 s  /[null]null dud/[den]N m p t M:1:1a",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START kalz/[kalz]A|kalz/[kalz]N m s|kalz/[kalz]P  /[null]null a/[mont]P  /[null]null dud/[den]N m p t M:1:1a",
		myAssertDisambiguate(input, xmlDisam),
		"xml PREP_A kalz a dud")
}

// O pronoun: o + N (no V inf exception) → R e p 3 obj.
func TestBretonRuleDisambiguator_OPronounTi(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "o ti"
	require.Equal(t,
		"/[null]SENT_START o/[o]D e sp|o/[o]I|o/[o]L o|o/[o]N m sp|o/[o]R e p 3 obj  /[null]null ti/[ti]N m s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START o/[o]R e p 3 obj  /[null]null ti/[ti]N m s",
		myAssertDisambiguate(input, xmlDisam),
		"xml O pronoun o ti")
}

// O + V inf → L o (progressive particle).
func TestBretonRuleDisambiguator_OParticleLabourat(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "o labourat"
	require.Equal(t,
		"/[null]SENT_START o/[o]D e sp|o/[o]I|o/[o]L o|o/[o]N m sp|o/[o]R e p 3 obj  /[null]null labourat/[labourat]V inf",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START o/[o]L o  /[null]null labourat/[labourat]V inf",
		myAssertDisambiguate(input, xmlDisam),
		"xml O particle o labourat")
}

// D_O: d’o → o is R e p 3 obj (BretonWordTokenizer keeps d’).
func TestBretonRuleDisambiguator_DO(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "d’o ti"
	require.Equal(t,
		"/[null]SENT_START d’/[da]P o/[o]D e sp|o/[o]I|o/[o]L o|o/[o]N m sp|o/[o]R e p 3 obj  /[null]null ti/[ti]N m s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START d’/[da]P o/[o]R e p 3 obj  /[null]null ti/[ti]N m s",
		myAssertDisambiguate(input, xmlDisam),
		"xml D_O d’o ti")
}

// EZ_AN: ez + an → ez is L e; an loses D e sp (article) reading.
func TestBretonRuleDisambiguator_EzAn(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "ez an"
	require.Equal(t,
		"/[null]SENT_START ez/[e]L e|ez/[e]P|ez/[monet]V pres 2 s|ez/[mont]V pres 2 s  /[null]null an/[an]D e sp|an/[monet]V pres 1 s|an/[mont]V pres 1 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ez/[e]L e  /[null]null an/[monet]V pres 1 s|an/[mont]V pres 1 s",
		myAssertDisambiguate(input, xmlDisam),
		"xml EZ_AN ez an")
}

// EN_EM: en em → both get X EN_EM.
func TestBretonRuleDisambiguator_EnEm(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "en em gannañ"
	require.Equal(t,
		"/[null]SENT_START en/[e]P  /[null]null em/[e]P  /[null]null gannañ/[kannañ]V inf M:1:1a:",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START en/[e]X EN_EM  /[null]null em/[e]X EN_EM  /[null]null gannañ/[kannañ]V inf M:1:1a:",
		myAssertDisambiguate(input, xmlDisam),
		"xml EN_EM en em gannañ")
}

// EZ_A: ez a → a is mont V pres 3 s.
func TestBretonRuleDisambiguator_EzA(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "ez a"
	require.Equal(t,
		"/[null]SENT_START ez/[e]L e|ez/[e]P|ez/[monet]V pres 2 s|ez/[mont]V pres 2 s  /[null]null a/[a]L a|a/[a]N m sp|a/[a]P|a/[monet]V impe 2 s|a/[monet]V pres 3 s|a/[mont]V impe 2 s|a/[mont]V pres 3 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ez/[e]L e  /[null]null a/[mont]V pres 3 s",
		myAssertDisambiguate(input, xmlDisam),
		"xml EZ_A ez a")
}

// MA: ma digarez → R e s 1 obj; ma ti → D e sp.
func TestBretonRuleDisambiguator_Ma(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START ma/[ma]C sub|ma/[ma]D e sp|ma/[ma]I  /[null]null digarez/[digarez]N m s|digarez/[digareziñ]V impe 2 s|digarez/[digareziñ]V pres 3 s",
		myAssertDisambiguate("ma digarez", demo),
		"demo ma digarez")
	require.Equal(t,
		"/[null]SENT_START ma/[ma]R e s 1 obj  /[null]null digarez/[digarez]N m s|digarez/[digareziñ]V impe 2 s|digarez/[digareziñ]V pres 3 s",
		myAssertDisambiguate("ma digarez", xmlDisam),
		"xml MA ma digarez")

	require.Equal(t,
		"/[null]SENT_START ma/[ma]C sub|ma/[ma]D e sp|ma/[ma]I  /[null]null ti/[ti]N m s",
		myAssertDisambiguate("ma ti", demo),
		"demo ma ti")
	require.Equal(t,
		"/[null]SENT_START ma/[ma]D e sp  /[null]null ti/[ti]N m s",
		myAssertDisambiguate("ma ti", xmlDisam),
		"xml MA ma ti")
}

// UR_N: ur + V-or-N ambiguous → filter to N.
func TestBretonRuleDisambiguator_UrNLabour(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "ur labour"
	require.Equal(t,
		"/[null]SENT_START ur/[un]D e sp  /[null]null labour/[labour]N m s|labour/[labourat]V pres 3 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ur/[un]D e sp  /[null]null labour/[labour]N m s",
		myAssertDisambiguate(input, xmlDisam),
		"xml UR_N ur labour")
}

// PAOT_MAT: paot mat → filter paot to J.
func TestBretonRuleDisambiguator_PaotMat(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "paot mat"
	require.Equal(t,
		"/[null]SENT_START paot/[baot]N f s M:3:|paot/[paot]J  /[null]null mat/[mat]J",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START paot/[paot]J  /[null]null mat/[mat]J",
		myAssertDisambiguate(input, xmlDisam),
		"xml PAOT_MAT paot mat")
}

// KOZH: kozh → filter to J.
func TestBretonRuleDisambiguator_Kozh(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "kozh den"
	require.Equal(t,
		"/[null]SENT_START kozh/[kozh]J|kozh/[kozhañ]V impe 2 s|kozh/[kozhañ]V pres 3 s  /[null]null den/[den]A|den/[den]N m s|den/[denañ]V impe 2 s|den/[denañ]V pres 3 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START kozh/[kozh]J  /[null]null den/[den]A|den/[den]N m s|den/[denañ]V impe 2 s|den/[denañ]V pres 3 s",
		myAssertDisambiguate(input, xmlDisam),
		"xml KOZH kozh den")
}

// PE_INT: sentence-initial pe + N → J itg.
func TestBretonRuleDisambiguator_PeInt(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "pe den"
	require.Equal(t,
		"/[null]SENT_START pe/[bezañ]V conf 3 s M:3:|pe/[pe]C coor|pe/[pe]J itg  /[null]null den/[den]A|den/[den]N m s|den/[denañ]V impe 2 s|den/[denañ]V pres 3 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START pe/[bezañ]J itg  /[null]null den/[den]A|den/[den]N m s|den/[denañ]V impe 2 s|den/[denañ]V pres 3 s",
		myAssertDisambiguate(input, xmlDisam),
		"xml PE_INT pe den")
}

// RA_L: sentence-initial ra + V futu → L r.
func TestBretonRuleDisambiguator_RaL(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "ra vo"
	require.Equal(t,
		"/[null]SENT_START ra/[ober]V impe 2 s|ra/[ober]V pres 3 s|ra/[ra]L r  /[null]null vo/[bezañ]V futu 3 s M:1:1a:1b:4:",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START ra/[ober]L r  /[null]null vo/[bezañ]V futu 3 s M:1:1a:1b:4:",
		myAssertDisambiguate(input, xmlDisam),
		"xml RA_L ra vo")
}

// RA_V + cascade: mont a ra → a becomes L a; ra becomes V pres 3 s.
func TestBretonRuleDisambiguator_MontARa(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "mont a ra"
	require.Equal(t,
		"/[null]SENT_START mont/[mont]V inf  /[null]null a/[a]L a|a/[a]N m sp|a/[a]P|a/[monet]V impe 2 s|a/[monet]V pres 3 s|a/[mont]V impe 2 s|a/[mont]V pres 3 s  /[null]null ra/[ober]V impe 2 s|ra/[ober]V pres 3 s|ra/[ra]L r",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START mont/[mont]V inf  /[null]null a/[mont]L a  /[null]null ra/[ober]V pres 3 s",
		myAssertDisambiguate(input, xmlDisam),
		"xml RA_V mont a ra")
}

// GANT: sentence-initial gant → P (replace; lemma may keep first reading).
func TestBretonRuleDisambiguator_GantStart(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "gant den"
	require.Equal(t,
		"/[null]SENT_START gant/[gant]P|gant/[kant]K e p M:1:1a:|gant/[kant]N m s M:1:1a:  /[null]null den/[den]A|den/[den]N m s|den/[denañ]V impe 2 s|den/[denañ]V pres 3 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START gant/[kant]P  /[null]null den/[den]A|den/[den]N m s|den/[denañ]V impe 2 s|den/[denañ]V pres 3 s",
		myAssertDisambiguate(input, xmlDisam),
		"xml GANT gant den")
}

// A_UNAN_DA_UNAN: a unan da unan → second unan gets X A_UNAN_DA_UNAN.
func TestBretonRuleDisambiguator_AUnanDaUnan(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "a unan da unan"
	require.Equal(t,
		"/[null]SENT_START a/[a]L a|a/[a]N m sp|a/[a]P|a/[monet]V impe 2 s|a/[monet]V pres 3 s|a/[mont]V impe 2 s|a/[mont]V pres 3 s  /[null]null unan/[unan]K e s|unan/[unaniñ]V impe 2 s|unan/[unaniñ]V pres 3 s  /[null]null da/[da]D e sp|da/[da]P  /[null]null unan/[unan]K e s|unan/[unaniñ]V impe 2 s|unan/[unaniñ]V pres 3 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START a/[a]L a|a/[a]N m sp|a/[a]P|a/[monet]V impe 2 s|a/[monet]V pres 3 s|a/[mont]V impe 2 s|a/[mont]V pres 3 s  /[null]null unan/[unan]K e s|unan/[unaniñ]V impe 2 s|unan/[unaniñ]V pres 3 s  /[null]null da/[da]D e sp|da/[da]P  /[null]null unan/[unaniñ]X A_UNAN_DA_UNAN",
		myAssertDisambiguate(input, xmlDisam),
		"xml A_UNAN_DA_UNAN")
}

// SENT_START subject + a + V → a becomes L a (rulegroup O).
func TestBretonRuleDisambiguator_MeALabour(t *testing.T) {
	demo, xmlDisam := setupBRDisambiguation(t)
	const input = "Me a labour"
	require.Equal(t,
		"/[null]SENT_START Me/[me]R suj e s 1  /[null]null a/[a]L a|a/[a]N m sp|a/[a]P|a/[monet]V impe 2 s|a/[monet]V pres 3 s|a/[mont]V impe 2 s|a/[mont]V pres 3 s  /[null]null labour/[labour]N m s|labour/[labourat]V pres 3 s",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Me/[me]R suj e s 1  /[null]null a/[mont]L a  /[null]null labour/[labour]N m s|labour/[labourat]V pres 3 s",
		myAssertDisambiguate(input, xmlDisam),
		"xml O Me a labour")
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// BretonWordTokenizer, SRXSentenceTokenizer(Breton), BretonTagger, disambiguator).
// Format: token/[lemma]POS readings sorted and joined by '|', tokens joined by space;
// null lemma/POS print as the literal "null" (Java string concat of null).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagbr.EnsureDefaultBretonTagger()
	tagger := tagbr.DefaultBretonTagger
	wt := brtok.NewBretonWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("br")
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
