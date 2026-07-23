package gl

// Outcome twins for Galician XmlRuleDisambiguator as used by GalicianHybridDisambiguator:
// Java new XmlRuleDisambiguator(new Galician()) with useGlobalDisambiguation=false.
// Cases derived from official resource/gl/disambiguation.xml <example type="ambiguous">
// (ADVERB_VERB_NOUN vía, QUOT open) + active pack rules with clear surface effects
// (NON_ADVERB, DET_CANS, PV, NUMBER, PERCENTAGES, CONTRACAO_L, VOGAIS, PUNCT)
// + real GalicianTagger readings — same bar as Polish/Russian RuleDisambiguator tests.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambiggl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/gl"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	taggl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/gl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadGLXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Galician())
// (useGlobalDisambiguation default false) over official resource/gl/disambiguation.xml.
func loadGLXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	// Prefer process cache (hybrid wire path); fall back to discover for isolation.
	if x := disambiggl.GalicianXmlRuleDisambiguator(); x != nil {
		return x
	}
	p := discoverGLDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "gl", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverGLDisambiguationXML() string {
	if p := os.Getenv("LANG_GL_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "gl",
		"src", "main", "resources", "org", "languagetool", "resource", "gl", "disambiguation.xml")
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

func setupGLDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if taggl.DiscoverGalicianPOSDict() == "" {
		t.Skip("galician.dict not in tree")
	}
	taggl.EnsureDefaultGalicianTagger()
	require.NotNil(t, taggl.DefaultGalicianTagger)
	require.NotNil(t, taggl.DefaultGalicianTagger.GetWordTagger())
	require.NotEmpty(t, taggl.GalicianPOSDictPath(), "real galician.dict must load")

	xml = loadGLXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("gl/disambiguation.xml not in tree or failed to load")
	}
	// Official GL pack (unifications + multiword POS + POS filters + number/punct/quot + …).
	require.GreaterOrEqual(t, len(xml.Rules), 200)
	return disambigxx.NewDemoDisambiguator(), xml
}

// ADVERB_VERB_NOUN + NON_ADVERB + PUNCT: official ambiguous example
// "Hai moito tempo que non vía publicacións da Ana." — vía drops NCFS000 noun reading.
func TestGalicianRuleDisambiguator_ViaVebNoun(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "Hai moito tempo que non vía publicacións da Ana."
	require.Equal(t,
		"/[null]SENT_START Hai/[haber]VMIP3S0  /[null]null moito/[moito]DI0MS0|moito/[moito]PI0MS000|moito/[moito]RG  /[null]null tempo/[tempo]NCMS000  /[null]null que/[que]CS|que/[que]DE0CN0|que/[que]DT0CN0|que/[que]NCMS000|que/[que]PE0CN000|que/[que]PR0CN000|que/[que]PT0CN000  /[null]null non/[non]NCMS000|non/[non]RN  /[null]null vía/[ver]VMII1S0|vía/[ver]VMII3S0|vía/[vía]NCFS000  /[null]null publicacións/[publicación]NCFP000  /[null]null da/[de]SPS00:DA  /[null]null Ana/[null]null ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// NON_ADVERB keeps non/RN; ADVERB_VERB_NOUN filters vía to V.*; PUNCT adds _PUNCT on '.'.
	require.Equal(t,
		"/[null]SENT_START Hai/[haber]VMIP3S0  /[null]null moito/[moito]DI0MS0|moito/[moito]PI0MS000|moito/[moito]RG  /[null]null tempo/[tempo]NCMS000  /[null]null que/[que]CS|que/[que]DE0CN0|que/[que]DT0CN0|que/[que]NCMS000|que/[que]PE0CN000|que/[que]PR0CN000|que/[que]PT0CN000  /[null]null non/[non]RN  /[null]null vía/[ver]VMII1S0|vía/[ver]VMII3S0  /[null]null publicacións/[publicación]NCFP000  /[null]null da/[de]SPS00:DA  /[null]null Ana/[null]null ./[.]_PUNCT",
		myAssertDisambiguate(input, xmlDisam),
		"xml ADVERB_VERB_NOUN vía + NON_ADVERB")
}

// QUOT rule 1: sentence-initial " → `` (official ambiguous example "\"Um teste.").
func TestGalicianRuleDisambiguator_QuotOpenSentenceStart(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "\"Um teste."
	require.Equal(t,
		"/[null]SENT_START \"/[null]null Um/[null]null  /[null]null teste/[ter]VMIP2S0:PP2CSA00|teste/[testar]VMM03S0|teste/[testar]VMSP1S0|teste/[testar]VMSP3S0 ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// QUOT open: disambig postag "``"; PUNCT adds _PUNCT on '.'.
	// (backticks are the open-quote POS tag from official gl/disambiguation.xml)
	require.Equal(t,
		"/[null]SENT_START \"/[\"]"+"``"+" Um/[null]null  /[null]null teste/[ter]VMIP2S0:PP2CSA00|teste/[testar]VMM03S0|teste/[testar]VMSP1S0|teste/[testar]VMSP3S0 ./[.]_PUNCT",
		myAssertDisambiguate(input, xmlDisam),
		"xml QUOT open + PUNCT")
}

// NON_ADVERB: non + verb → keep only RN (active rule; pack pattern).
func TestGalicianRuleDisambiguator_NonAdverb(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "non quero"
	require.Equal(t,
		"/[null]SENT_START non/[non]NCMS000|non/[non]RN  /[null]null quero/[querer]VMIP1S0",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START non/[non]RN  /[null]null quero/[querer]VMIP1S0",
		myAssertDisambiguate(input, xmlDisam),
		"xml NON_ADVERB")
}

// DET_CANS: det + can(s) → keep N.* on can (active rule; pack pattern).
func TestGalicianRuleDisambiguator_DetCans(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "o can ladra"
	require.Equal(t,
		"/[null]SENT_START o/[o]DA0MS0|o/[o]NCMS000|o/[o]PP3MSA00  /[null]null can/[can]NCMS000|can/[can]RG  /[null]null ladra/[ladra]NCFS000|ladra/[ladrar]VMIP3S0|ladra/[ladrar]VMM02S0",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// DET_SUBST keeps o as D.*; DET_CANS filters can to N.* (drops RG).
	require.Equal(t,
		"/[null]SENT_START o/[o]DA0MS0  /[null]null can/[can]NCMS000  /[null]null ladra/[ladra]NCFS000|ladra/[ladrar]VMIP3S0|ladra/[ladrar]VMM02S0",
		myAssertDisambiguate(input, xmlDisam),
		"xml DET_SUBST + DET_CANS")
}

// PV: pronoun + N|V → keep V (active rule; pack comment "Ele casa").
func TestGalicianRuleDisambiguator_PVEuCasa(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "eu casa"
	require.Equal(t,
		"/[null]SENT_START eu/[eu]PP1CSN00  /[null]null casa/[casa]NCFS000|casa/[casar]VMIP3S0|casa/[casar]VMM02S0",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START eu/[eu]PP1CSN00  /[null]null casa/[casar]VMIP3S0|casa/[casar]VMM02S0",
		myAssertDisambiguate(input, xmlDisam),
		"xml PV eu casa")
}

// NUMBER: plain digits → Z0CN0 (active rule; pack NUMBER rulegroup).
func TestGalicianRuleDisambiguator_NumberDigits(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "123"
	require.Equal(t,
		"/[null]SENT_START 123/[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START 123/[123]Z0CN0",
		myAssertDisambiguate(input, xmlDisam),
		"xml NUMBER digits")
}

// NUMBER written: cinco → Z0CP0 (active rule; pack NUMBER rulegroup).
func TestGalicianRuleDisambiguator_NumberCinco(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "cinco gatos"
	require.Equal(t,
		"/[null]SENT_START cinco/[cincar]VMIP1S0|cinco/[cinco]NCMS000  /[null]null gatos/[gato]NCMP000",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START cinco/[cinco]Z0CP0  /[null]null gatos/[gato]NCMP000",
		myAssertDisambiguate(input, xmlDisam),
		"xml NUMBER cinco")
}

// PERCENTAGES: 25% → NCMP000; 1% → NCMC000 (active rules).
func TestGalicianRuleDisambiguator_Percentages(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START 25%/[null]null",
		myAssertDisambiguate("25%", demo),
		"demo 25%")
	require.Equal(t,
		"/[null]SENT_START 25%/[25%]NCMP000",
		myAssertDisambiguate("25%", xmlDisam),
		"xml PERCENTAGES 25%")

	require.Equal(t,
		"/[null]SENT_START 1%/[null]null",
		myAssertDisambiguate("1%", demo),
		"demo 1%")
	require.Equal(t,
		"/[null]SENT_START 1%/[1%]NCMC000",
		myAssertDisambiguate("1%", xmlDisam),
		"xml PERCENTAGES 1%")
}

// CONTRACAO_L: l' + masc → DA0MS0; l' + fem → DA0FS0 (active rulegroup).
func TestGalicianRuleDisambiguator_ContracaoL(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)

	require.Equal(t,
		"/[null]SENT_START l/[l]NCMS000 '/[null]null amigo/[amigar]VMIP1S0|amigo/[amigo]AQ0MS0|amigo/[amigo]NCMS000",
		myAssertDisambiguate("l'amigo", demo),
		"demo l'amigo")
	require.Equal(t,
		"/[null]SENT_START l/[l]DA0MS0 '/[']DA0MS0 amigo/[amigar]VMIP1S0|amigo/[amigo]AQ0MS0|amigo/[amigo]NCMS000",
		myAssertDisambiguate("l'amigo", xmlDisam),
		"xml CONTRACAO_L l'amigo")

	require.Equal(t,
		"/[null]SENT_START l/[l]NCMS000 '/[null]null amiga/[amigar]VMIP3S0|amiga/[amigar]VMM02S0|amiga/[amigo]AQ0FS0|amiga/[amigo]NCFS000",
		myAssertDisambiguate("l'amiga", demo),
		"demo l'amiga")
	require.Equal(t,
		"/[null]SENT_START l/[l]DA0FS0 '/[']DA0FS0 amiga/[amigar]VMIP3S0|amiga/[amigar]VMM02S0|amiga/[amigo]AQ0FS0|amiga/[amigo]NCFS000",
		myAssertDisambiguate("l'amiga", xmlDisam),
		"xml CONTRACAO_L l'amiga")
}

// VOGAIS: a e b → remove NCMS000 on e (active rule; exception o|letra|vogal).
func TestGalicianRuleDisambiguator_VogaisE(t *testing.T) {
	demo, xmlDisam := setupGLDisambiguation(t)
	const input = "a e b"
	require.Equal(t,
		"/[null]SENT_START a/[a]SPS00|a/[o]DA0FS0|a/[o]PP3FSA00  /[null]null e/[e]CC|e/[e]NCMS000  /[null]null b/[b]NCMS000",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// VOGAIS drops e/NCMS000; DET_SUBST keeps a as D.* only.
	require.Equal(t,
		"/[null]SENT_START a/[o]DA0FS0  /[null]null e/[e]CC  /[null]null b/[b]NCMS000",
		myAssertDisambiguate(input, xmlDisam),
		"xml VOGAIS + DET_SUBST")
}

// Hybrid Rules stage uses the same official XML (Java eager XmlRuleDisambiguator field).
func TestGalicianHybridDisambiguator_RulesStageMatchesXml(t *testing.T) {
	_, xmlDisam := setupGLDisambiguation(t)
	hybrid := disambiggl.NewGalicianHybridDisambiguator()
	require.NotNil(t, hybrid.Rules, "Java constructs XmlRuleDisambiguator eagerly")
	const input = "Hai moito tempo que non vía publicacións da Ana."
	require.Equal(t,
		myAssertDisambiguate(input, xmlDisam),
		myAssertDisambiguate(input, hybrid),
		"hybrid Rules stage == standalone XmlRuleDisambiguator")
}

// Multiword isolation: hybrid with Rules=nil still runs MultiWordChunker
// (Java stage order multiword → XML).
func TestGalicianHybridDisambiguator_MultiwordIsolationRulesNil(t *testing.T) {
	if taggl.DiscoverGalicianPOSDict() == "" {
		t.Skip("galician.dict not in tree")
	}
	taggl.EnsureDefaultGalicianTagger()
	chunker := loadGalicianMultiWordChunker(t)
	hybrid := disambiggl.NewGalicianHybridDisambiguatorWithStages(chunker, nil)
	require.Nil(t, hybrid.Rules)
	const input = "abaixo de"
	want := "/[null]SENT_START abaixo/[abaixar]VMIP1S0|abaixo/[abaixo de]<SP000>|abaixo/[abaixo]RG  /[null]null de/[abaixo de]</SP000>|de/[de]NCMS000|de/[de]SPS00"
	require.Equal(t, want, myAssertDisambiguate(input, hybrid), "multiword SP000 with Rules=nil")
}

func loadGalicianMultiWordChunker(t *testing.T) *disambiguation.MultiWordChunker {
	t.Helper()
	p := glMultiwordsPath(t)
	f, err := os.Open(p)
	require.NoError(t, err)
	defer f.Close()
	c, err := disambiggl.OpenGalicianMultiWordChunker(f)
	require.NoError(t, err)
	return c
}

func glMultiwordsPath(t *testing.T) string {
	t.Helper()
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "gl",
		"src", "main", "resources", "org", "languagetool", "resource", "gl", "multiwords.txt")
	wd, err := os.Getwd()
	require.NoError(t, err)
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
	t.Fatal("gl/multiwords.txt not found")
	return ""
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Galician), GalicianTagger, disambiguator).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	taggl.EnsureDefaultGalicianTagger()
	tagger := taggl.DefaultGalicianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("gl")
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
