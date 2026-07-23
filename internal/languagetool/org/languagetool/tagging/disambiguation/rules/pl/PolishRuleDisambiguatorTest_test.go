package pl

// Outcome twins for Polish XmlRuleDisambiguator as used by PolishHybridDisambiguator:
// Java new XmlRuleDisambiguator(new Polish()) with useGlobalDisambiguation=false.
// Cases derived from official resource/pl/disambiguation.xml <example type="ambiguous">
// (DWUKROPEK_GODZINA, COMP_COMMA, number_comma, przeszlo, quote_no_interp, bez prep, mają filter, …)
// + real PolishTagger readings — same bar as Russian/Danish/Arabic RuleDisambiguator tests.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pl"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	tagpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadPLXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Polish())
// (useGlobalDisambiguation default false) over official resource/pl/disambiguation.xml.
func loadPLXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	// Prefer process cache (hybrid wire path); fall back to discover for isolation.
	if x := disambigpl.PolishXmlRuleDisambiguator(); x != nil {
		return x
	}
	p := discoverPLDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "pl", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverPLDisambiguationXML() string {
	if p := os.Getenv("LANG_PL_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pl",
		"src", "main", "resources", "org", "languagetool", "resource", "pl", "disambiguation.xml")
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

func setupPLDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagpl.DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not in tree")
	}
	tagpl.EnsureDefaultPolishTagger()
	require.NotNil(t, tagpl.DefaultPolishTagger)
	require.NotNil(t, tagpl.DefaultPolishTagger.GetWordTagger())
	require.NotEmpty(t, tagpl.PolishPOSDictPath(), "real polish.dict must load")

	xml = loadPLXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("pl/disambiguation.xml not in tree or failed to load")
	}
	// Official PL pack is large (unifications + punctuation + POS filters + …).
	require.GreaterOrEqual(t, len(xml.Rules), 200)
	return disambigxx.NewDemoDisambiguator(), xml
}

// DWUKROPEK_GODZINA: 15:34 → colon gets interp:nospace (official ambiguous example).
func TestPolishRuleDisambiguator_ColonTime1534(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "Pociąg odjeżdża o 15:34."
	require.Equal(t,
		"/[null]SENT_START Pociąg/[pociąg]subst:sg:acc:m3|Pociąg/[pociąg]subst:sg:nom:m3  /[null]null odjeżdża/[odjeżdżać]verb:fin:sg:ter:imperf:nonrefl  /[null]null o/[o]interj|o/[o]prep:acc|o/[o]prep:loc|o/[ocean]brev:pun|o/[ojciec]brev:pun  /[null]null 15/[null]null :/[null]null 34/[null]null ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// DWUKROPEK_GODZINA adds interp:nospace on ':'; cascade also trims "o" brev readings.
	require.Equal(t,
		"/[null]SENT_START Pociąg/[pociąg]subst:sg:acc:m3|Pociąg/[pociąg]subst:sg:nom:m3  /[null]null odjeżdża/[odjeżdżać]verb:fin:sg:ter:imperf:nonrefl  /[null]null o/[o]prep:acc|o/[o]prep:loc  /[null]null 15/[null]null :/[:]interp:nospace 34/[null]null ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml DWUKROPEK_GODZINA 15:34")
}

// COMP_COMMA: ale after comma → add comp:comma (official ambiguous example).
func TestPolishRuleDisambiguator_CompCommaAle(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "Lubię go, ale kupię mu pistolet."
	require.Equal(t,
		"/[null]SENT_START Lubię/[lubić]verb:fin:sg:pri:imperf:nonrefl|Lubię/[lubić]verb:fin:sg:pri:imperf:refl.nonrefl  /[null]null go/[go]subst:pl:acc:n2|go/[go]subst:pl:dat:n2|go/[go]subst:pl:gen:n2|go/[go]subst:pl:inst:n2|go/[go]subst:pl:loc:n2|go/[go]subst:pl:nom:n2|go/[go]subst:pl:voc:n2|go/[go]subst:sg:acc:n2|go/[go]subst:sg:dat:n2|go/[go]subst:sg:gen:n2|go/[go]subst:sg:inst:n2|go/[go]subst:sg:loc:n2|go/[go]subst:sg:nom:n2|go/[go]subst:sg:voc:n2|go/[on]ppron3:sg:acc:m1.m2.m3:ter:nakc:npraep|go/[on]ppron3:sg:gen:m1.m2.m3:ter:nakc:npraep|go/[on]ppron3:sg:gen:n1.n2:ter:nakc:npraep ,/[null]null  /[null]null ale/[ale]conj|ale/[ale]qub  /[null]null kupię/[kupia]subst:sg:acc:f|kupię/[kupić]verb:fin:sg:pri:imperf:refl.nonrefl|kupię/[kupić]verb:fin:sg:pri:perf:refl.nonrefl  /[null]null mu/[mu]interj|mu/[on]ppron3:sg:dat:m1.m2.m3:ter:nakc:npraep|mu/[on]ppron3:sg:dat:n1.n2:ter:nakc:npraep  /[null]null pistolet/[pistolet]subst:sg:acc:m3|pistolet/[pistolet]subst:sg:nom:m3 ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// COMP_COMMA adds ale/comp:comma; other pack rules filter "go" and "kupię".
	require.Equal(t,
		"/[null]SENT_START Lubię/[lubić]verb:fin:sg:pri:imperf:nonrefl|Lubię/[lubić]verb:fin:sg:pri:imperf:refl.nonrefl  /[null]null go/[on]ppron3:sg:acc:m1.m2.m3:ter:nakc:npraep ,/[null]null  /[null]null ale/[ale]comp:comma|ale/[ale]conj|ale/[ale]qub  /[null]null kupię/[kupić]verb:fin:sg:pri:imperf:refl.nonrefl|kupię/[kupić]verb:fin:sg:pri:perf:refl.nonrefl  /[null]null mu/[mu]interj|mu/[on]ppron3:sg:dat:m1.m2.m3:ter:nakc:npraep|mu/[on]ppron3:sg:dat:n1.n2:ter:nakc:npraep  /[null]null pistolet/[pistolet]subst:sg:acc:m3|pistolet/[pistolet]subst:sg:nom:m3 ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml COMP_COMMA ale")
}

// number_comma: 85,45 → decimal comma is interp not interp:comma (official ambiguous example).
func TestPolishRuleDisambiguator_NumberComma8545(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "85,45"
	require.Equal(t,
		"/[null]SENT_START 85/[null]null ,/[null]null 45/[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	// number_comma replace → interp (after PUNCT_NO_DOT would have added interp:comma).
	require.Equal(t,
		"/[null]SENT_START 85/[null]null ,/[,]interp 45/[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml number_comma 85,45")
}

// przeszlo: Przeszło 30 → keep only qub (official ambiguous example).
func TestPolishRuleDisambiguator_Przeszlo30Panow(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "Przeszło 30 panów pije wódkę."
	require.Equal(t,
		"/[null]SENT_START Przeszło/[przejść]verb:praet:sg:n1.n2:ter:perf:nonrefl|Przeszło/[przejść]verb:praet:sg:n1.n2:ter:perf:refl|Przeszło/[przeszło]qub|Przeszło/[przeszły]adja  /[null]null 30/[null]null  /[null]null panów/[pan]subst:pl:acc:m1|panów/[pan]subst:pl:gen:m1  /[null]null pije/[pić]verb:fin:sg:ter:imperf:refl.nonrefl  /[null]null wódkę/[wódka]subst:sg:acc:f ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Przeszło/[przeszło]qub  /[null]null 30/[null]null  /[null]null panów/[pan]subst:pl:gen:m1  /[null]null pije/[pić]verb:fin:sg:ter:imperf:refl.nonrefl  /[null]null wódkę/[wódka]subst:sg:acc:f ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml przeszlo Przeszło 30")
}

// quote_no_interp: „wyjątkowo” → quotes become interp (official ambiguous examples).
func TestPolishRuleDisambiguator_QuoteWyjatkowo(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "On był „wyjątkowo” wredny."
	require.Equal(t,
		"/[null]SENT_START On/[on]adj:sg:acc:m3:pos|On/[on]adj:sg:nom.voc:m1.m2.m3:pos|On/[on]ppron3:sg:nom:m1.m2.m3:ter:akc.nakc:praep.npraep  /[null]null był/[być]verb:praet:sg:m1.m2.m3:ter:imperf:nonrefl  /[null]null „/[null]null wyjątkowo/[wyjątkowo]adv:pos|wyjątkowo/[wyjątkowy]adja ”/[null]null  /[null]null wredny/[wredny]adj:sg:acc:m3:pos|wredny/[wredny]adj:sg:nom.voc:m1.m2.m3:pos ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START On/[on]ppron3:sg:nom:m1.m2.m3:ter:akc.nakc:praep.npraep  /[null]null był/[być]verb:praet:sg:m1.m2.m3:ter:imperf:nonrefl  /[null]null „/[„]interp wyjątkowo/[wyjątkowo]adv:pos ”/[”]interp  /[null]null wredny/[wredny]adj:sg:acc:m3:pos|wredny/[wredny]adj:sg:nom.voc:m1.m2.m3:pos ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml quote_no_interp")
}

// bez + gen: keep prep (official ambiguous "Proszę kremówkę bez kremu").
func TestPolishRuleDisambiguator_BezKremuPrep(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "Proszę kremówkę bez kremu."
	require.Equal(t,
		"/[null]SENT_START Proszę/[prosić]verb:fin:sg:pri:imperf:refl|Proszę/[prosić]verb:fin:sg:pri:imperf:refl.nonrefl  /[null]null kremówkę/[kremówka]subst:sg:acc:f  /[null]null bez/[bez]prep:gen:nwok|bez/[bez]subst:sg:acc:m3|bez/[bez]subst:sg:nom:m3|bez/[beza]subst:pl:gen:f  /[null]null kremu/[krem]subst:sg:gen:m3 ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Proszę/[prosić]verb:fin:sg:pri:imperf:refl|Proszę/[prosić]verb:fin:sg:pri:imperf:refl.nonrefl  /[null]null kremówkę/[kremówka]subst:sg:acc:f  /[null]null bez/[bez]prep:gen:nwok  /[null]null kremu/[krem]subst:sg:gen:m3 ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml bez prep")
}

// Oni mają robaki → filter to mieć (official ambiguous example).
func TestPolishRuleDisambiguator_MajaMiec(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "Oni mają robaki."
	require.Equal(t,
		"/[null]SENT_START Oni/[on]adj:pl:nom.voc:m1.p1:pos|Oni/[on]ppron3:pl:nom:m1.p1:ter:akc.nakc:praep.npraep  /[null]null mają/[maić]verb:fin:pl:ter:imperf:refl.nonrefl|mają/[maja]subst:sg:inst:f|mają/[mieć]verb:fin:pl:ter:imperf:refl.nonrefl  /[null]null robaki/[robak]subst:pl:acc:m2|robaki/[robak]subst:pl:nom:m2|robaki/[robak]subst:pl:voc:m2 ./[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Oni/[on]adj:pl:nom.voc:m1.p1:pos|Oni/[on]ppron3:pl:nom:m1.p1:ter:akc.nakc:praep.npraep  /[null]null mają/[mieć]verb:fin:pl:ter:imperf:refl.nonrefl  /[null]null robaki/[robak]subst:pl:acc:m2|robaki/[robak]subst:pl:nom:m2|robaki/[robak]subst:pl:voc:m2 ./[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml mają → mieć")
}

// od dawna → dawna/adjp (official ambiguous example).
func TestPolishRuleDisambiguator_OdDawna(t *testing.T) {
	demo, xmlDisam := setupPLDisambiguation(t)
	const input = "Czy od dawna tak jest?"
	require.Equal(t,
		"/[null]SENT_START Czy/[czy]conj|Czy/[czy]qub  /[null]null od/[od]prep:gen:nwok|od/[oda]subst:pl:gen:f  /[null]null dawna/[dawny]adj:sg:nom.voc:f:pos  /[null]null tak/[tak]adv:pos|tak/[tak]qub|tak/[taka]subst:pl:gen:f  /[null]null jest/[być]verb:fin:sg:ter:imperf:nonrefl ?/[null]null",
		myAssertDisambiguate(input, demo),
		"demo disambiguator")
	require.Equal(t,
		"/[null]SENT_START Czy/[Czy]comp:comma|Czy/[czy]conj|Czy/[czy]qub  /[null]null od/[od]prep:gen:nwok  /[null]null dawna/[dawny]adjp  /[null]null tak/[tak]adv:pos  /[null]null jest/[być]verb:fin:sg:ter:imperf:nonrefl ?/[null]null",
		myAssertDisambiguate(input, xmlDisam),
		"xml od dawna adjp")
}

// Hybrid Rules stage uses the same official XML (Java eager XmlRuleDisambiguator field).
func TestPolishHybridDisambiguator_RulesStageMatchesXml(t *testing.T) {
	_, xmlDisam := setupPLDisambiguation(t)
	hybrid := disambigpl.NewPolishHybridDisambiguator()
	require.NotNil(t, hybrid.Rules, "Java constructs XmlRuleDisambiguator eagerly")
	const input = "Przeszło 30 panów pije wódkę."
	require.Equal(t,
		myAssertDisambiguate(input, xmlDisam),
		myAssertDisambiguate(input, hybrid),
		"hybrid Rules stage == standalone XmlRuleDisambiguator")
}

// Multiword isolation: hybrid with Rules=nil still runs MultiWordChunker (Java stage order XML→MW).
func TestPolishHybridDisambiguator_MultiwordIsolationRulesNil(t *testing.T) {
	if tagpl.DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not in tree")
	}
	tagpl.EnsureDefaultPolishTagger()
	chunker := loadPolishMultiWordChunker(t)
	hybrid := disambigpl.NewPolishHybridDisambiguatorWithStages(chunker, nil)
	require.Nil(t, hybrid.Rules)
	const input = "Test..."
	want := "/[null]SENT_START Test/[test]subst:sg:acc:m3|Test/[test]subst:sg:nom:m3 ./[...]<ELLIPSIS> ./[null]null ./[...]</ELLIPSIS>"
	require.Equal(t, want, myAssertDisambiguate(input, hybrid), "multiword ELLIPSIS with Rules=nil")
}

func loadPolishMultiWordChunker(t *testing.T) *disambiguation.MultiWordChunker {
	t.Helper()
	p := plMultiwordsPath(t)
	f, err := os.Open(p)
	require.NoError(t, err)
	defer f.Close()
	c, err := disambiguation.NewMultiWordChunkerFromReader(f, disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: false,
		AllowAllUppercase:     false,
		AllowTitlecase:        false,
	})
	require.NoError(t, err)
	return c
}

func plMultiwordsPath(t *testing.T) string {
	t.Helper()
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pl",
		"src", "main", "resources", "org", "languagetool", "resource", "pl", "multiwords.txt")
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
	t.Fatal("pl/multiwords.txt not found")
	return ""
}

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Polish), PolishTagger, disambiguator).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagpl.EnsureDefaultPolishTagger()
	tagger := tagpl.DefaultPolishTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("pl")
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
