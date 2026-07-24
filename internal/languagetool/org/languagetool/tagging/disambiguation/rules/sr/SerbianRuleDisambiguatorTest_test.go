package sr

// Outcome twins for Serbian XmlRuleDisambiguator as used by SerbianHybridDisambiguator:
// Java new XmlRuleDisambiguator(new Serbian()) with useGlobalDisambiguation=false.
// Official pack (resource/sr/disambiguation.xml) has exactly 1 rule:
//   RIMSKI_BROJEVI — case_sensitive regexp Roman numerals → ignore_spelling.
// Primary fidelity bar: matched tokens IsIgnoredBySpeller; readings typically unchanged.
// Real EkavianTagger + WordTokenizer (Java Serbian uses Language default WordTokenizer +
// SRXSentenceTokenizer(this) + EkavianTagger). Same bar as SV/EO/BR.
// Upstream has no dedicated Serbian XmlRule unit test.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigsr "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/sr"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	tagsr "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/sr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadSRXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Serbian())
// (useGlobalDisambiguation default false) over official resource/sr/disambiguation.xml.
func loadSRXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	// Prefer process cache (disambiguation/sr loader); fall back to discover for isolation.
	if x := disambigsr.SerbianXmlRuleDisambiguator(); x != nil {
		return x
	}
	p := discoverSRDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "sr", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverSRDisambiguationXML() string {
	if p := os.Getenv("LANG_SR_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sr",
		"src", "main", "resources", "org", "languagetool", "resource", "sr", "disambiguation.xml")
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

func setupSRDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagsr.DiscoverEkavianPOSDict() == "" {
		t.Skip("ekavian serbian.dict not in tree")
	}
	tagsr.EnsureDefaultEkavianTagger()
	require.NotNil(t, tagsr.DefaultEkavianTagger)
	require.NotNil(t, tagsr.DefaultEkavianTagger.GetWordTagger())
	require.NotEmpty(t, tagsr.EkavianPOSDictPath(), "real ekavian serbian.dict must load")
	// SerbianTagger default is also Ekavian (Java Serbian.getTagger → EkavianTagger).
	tagsr.EnsureDefaultSerbianTagger()
	require.NotNil(t, tagsr.DefaultSerbianTagger)

	xml = loadSRXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("sr/disambiguation.xml not in tree or failed to load")
	}
	// Official SR pack: exactly 1 rule (RIMSKI_BROJEVI).
	require.Equal(t, 1, len(xml.Rules), "official SR pack has exactly 1 rule")
	return disambigxx.NewDemoDisambiguator(), xml
}

// analyzeSR builds an AnalyzedSentence like other *RuleDisambiguator tests:
// WordTokenizer + SRXSentenceTokenizer(sr) + EkavianTagger + optional disambiguator.
// Returns the first sentence only (tests use single-sentence inputs).
func analyzeSR(input string, dis disambiguation.Disambiguator) *languagetool.AnalyzedSentence {
	tagsr.EnsureDefaultEkavianTagger()
	tagger := tagsr.DefaultEkavianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("sr")
	sentences := st.Tokenize(input)
	if len(sentences) == 0 {
		return nil
	}
	sentence := sentences[0]
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
	return finalSentence
}

func wordTokens(sent *languagetool.AnalyzedSentence) map[string]*languagetool.AnalyzedTokenReadings {
	out := make(map[string]*languagetool.AnalyzedTokenReadings)
	if sent == nil {
		return out
	}
	for _, tr := range sent.GetTokensWithoutWhitespace() {
		if tr == nil {
			continue
		}
		// First occurrence wins (fine for our short phrases).
		if _, ok := out[tr.GetToken()]; !ok {
			out[tr.GetToken()] = tr
		}
	}
	return out
}

func requireIgnoredBySpeller(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	toks := wordTokens(sent)
	for _, s := range surfaces {
		tr, ok := toks[s]
		require.True(t, ok, "token %q missing in sentence (have %v)", s, tokenSurfaces(sent))
		require.True(t, tr.IsIgnoredBySpeller(), "%q must be ignore_spelling", s)
	}
}

func requireNotIgnoredBySpeller(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	toks := wordTokens(sent)
	for _, s := range surfaces {
		tr, ok := toks[s]
		require.True(t, ok, "token %q missing in sentence (have %v)", s, tokenSurfaces(sent))
		require.False(t, tr.IsIgnoredBySpeller(), "%q must not be ignore_spelling", s)
	}
}

// --- Loader / pack shape ----------------------------------------------------

func TestSerbianXmlRuleDisambiguator_LoadsOfficialPack(t *testing.T) {
	_, xmlDisam := setupSRDisambiguation(t)
	// Official file has exactly 1 top-level <rule>: RIMSKI_BROJEVI.
	require.Equal(t, 1, len(xmlDisam.Rules), "official SR pack rule count")
	require.Equal(t, "RIMSKI_BROJEVI", xmlDisam.Rules[0].GetID())
	// useGlobal=false: only language XML, no disambiguation-global.xml.
	// Action is ignore_spelling (Java DisambiguatorAction.IGNORE_SPELLING).
	require.Equal(t, disambigrules.ActionIgnoreSpelling, xmlDisam.Rules[0].Action)
}

// --- RIMSKI_BROJEVI: Roman numerals → ignore_spelling ----------------------

// Positive Roman numerals that match the official case_sensitive regexp:
//   (?:M*(?:D?C{0,3}|C[DM])(?:L?X{0,3}|X[LC])(?:V?I{0,3}|I[VX]))
func TestSerbianRuleDisambiguator_RimskiBrojeviIgnore(t *testing.T) {
	demo, xmlDisam := setupSRDisambiguation(t)

	positives := []string{
		"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX",
		"X", "XII", "XX", "XL", "L", "XC", "C", "CD", "D", "CM", "M",
		"MCMXCIX", "MMXX",
	}
	for _, roman := range positives {
		t.Run(roman, func(t *testing.T) {
			demoSent := analyzeSR(roman, demo)
			xmlSent := analyzeSR(roman, xmlDisam)
			requireNotIgnoredBySpeller(t, demoSent, roman)
			requireIgnoredBySpeller(t, xmlSent, roman)
			// ignore_spelling is a flag only — does not immunize.
			require.False(t, wordTokens(xmlSent)[roman].IsImmunized(),
				"%q ignore_spelling must not immunize", roman)
		})
	}
}

// case_sensitive=yes: lowercase Roman forms must NOT match.
func TestSerbianRuleDisambiguator_RimskiBrojeviCaseSensitive(t *testing.T) {
	_, xmlDisam := setupSRDisambiguation(t)

	for _, low := range []string{"i", "ii", "iii", "iv", "v", "x", "xii", "xx", "mcmxcix"} {
		t.Run(low, func(t *testing.T) {
			xmlSent := analyzeSR(low, xmlDisam)
			requireNotIgnoredBySpeller(t, xmlSent, low)
		})
	}
}

// Negative controls: non-Roman words / invalid Roman forms stay clean.
func TestSerbianRuleDisambiguator_RimskiBrojeviNegatives(t *testing.T) {
	demo, xmlDisam := setupSRDisambiguation(t)

	// Invalid Roman / non-matching surfaces.
	// IIII and VV are not valid under the official regex; IC / VX / ABC don't match.
	negatives := []string{"IIII", "VV", "IC", "VX", "ABC", "hello", "test", "Zdravo"}
	for _, s := range negatives {
		t.Run(s, func(t *testing.T) {
			demoSent := analyzeSR(s, demo)
			xmlSent := analyzeSR(s, xmlDisam)
			requireNotIgnoredBySpeller(t, demoSent, s)
			requireNotIgnoredBySpeller(t, xmlSent, s)
		})
	}

	// Cyrillic Serbian word (from Ekavian tagger tests) must not be ignored.
	const cyr = "радим"
	demoSent := analyzeSR(cyr, demo)
	xmlSent := analyzeSR(cyr, xmlDisam)
	requireNotIgnoredBySpeller(t, demoSent, cyr)
	requireNotIgnoredBySpeller(t, xmlSent, cyr)
}

// Roman numeral in a short phrase: only the Roman token is ignored.
func TestSerbianRuleDisambiguator_RimskiBrojeviInContext(t *testing.T) {
	demo, xmlDisam := setupSRDisambiguation(t)
	// "vek XX" style: XX matches, surrounding words do not.
	const input = "vek XX je"
	demoSent := analyzeSR(input, demo)
	xmlSent := analyzeSR(input, xmlDisam)

	requireNotIgnoredBySpeller(t, demoSent, "vek", "XX", "je")
	requireIgnoredBySpeller(t, xmlSent, "XX")
	requireNotIgnoredBySpeller(t, xmlSent, "vek", "je")
}

// Demo vs xml: demo leaves IsIgnoredBySpeller==false; xml sets true on matches.
func TestSerbianRuleDisambiguator_DemoVsXmlIgnoreFlag(t *testing.T) {
	demo, xmlDisam := setupSRDisambiguation(t)
	const input = "XII"

	demoSent := analyzeSR(input, demo)
	xmlSent := analyzeSR(input, xmlDisam)

	require.False(t, wordTokens(demoSent)["XII"].IsIgnoredBySpeller())
	require.True(t, wordTokens(xmlSent)["XII"].IsIgnoredBySpeller())

	// Reading strings equal (ignore_spelling is a flag, not a POS rewrite).
	require.Equal(t,
		myAssertDisambiguate(input, demo),
		myAssertDisambiguate(input, xmlDisam),
		"ignore_spelling must not alter reading strings")
}

// Control sentence with no Roman numerals: demo and xml flags/readings match.
func TestSerbianRuleDisambiguator_ControlUnmatchedStaysClean(t *testing.T) {
	demo, xmlDisam := setupSRDisambiguation(t)
	// Use a phrase that is not a Roman numeral under the official regex.
	const input = "Zdravo svete"
	demoSent := analyzeSR(input, demo)
	xmlSent := analyzeSR(input, xmlDisam)

	require.Equal(t, myAssertDisambiguate(input, demo), myAssertDisambiguate(input, xmlDisam))

	for _, tr := range xmlSent.GetTokensWithoutWhitespace() {
		if tr == nil || tr.GetToken() == "" {
			continue
		}
		require.False(t, tr.IsIgnoredBySpeller(), "unexpected ignore_spelling on %q", tr.GetToken())
		require.False(t, tr.IsImmunized(), "unexpected immunize on %q", tr.GetToken())
	}
	for _, tr := range demoSent.GetTokensWithoutWhitespace() {
		if tr == nil {
			continue
		}
		require.False(t, tr.IsIgnoredBySpeller())
		require.False(t, tr.IsImmunized())
	}
}

// Hybrid stages use the same official multiwords + XML Java constructs eagerly.
// Official sr/multiwords.txt is empty → multiword stage is a no-op; readings match
// standalone XmlRuleDisambiguator (multiword first, then XML).
func TestSerbianHybridDisambiguator_RulesStageMatchesXml(t *testing.T) {
	_, xmlDisam := setupSRDisambiguation(t)
	hybrid := disambigsr.NewSerbianHybridDisambiguator()
	require.NotNil(t, hybrid.Rules, "Java constructs XmlRuleDisambiguator eagerly")
	require.NotNil(t, hybrid.Chunker, "Java constructs MultiWordChunker eagerly (empty multiwords OK)")

	const input = "XXI i MCMXCIX"
	// Note: XXI matches (X + X + I); "i" lowercase does not (case_sensitive).
	// Empty multiword stage does not change readings/flags vs XML alone.
	require.Equal(t,
		myAssertDisambiguate(input, xmlDisam),
		myAssertDisambiguate(input, hybrid),
		"hybrid multiword→XML readings == standalone XmlRuleDisambiguator")

	hs := analyzeSR(input, hybrid)
	requireIgnoredBySpeller(t, hs, "XXI", "MCMXCIX")
	requireNotIgnoredBySpeller(t, hs, "i")
	// Hybrid Rules pointer is the process-cached instance.
	require.Same(t, xmlDisam, hybrid.Rules)
}

// --- myAssert helpers (parity with other language RuleDisambiguator tests) ---

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Serbian), EkavianTagger, disambiguator).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagsr.EnsureDefaultEkavianTagger()
	tagger := tagsr.DefaultEkavianTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("sr")
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

func tokenSurfaces(sent *languagetool.AnalyzedSentence) []string {
	var out []string
	if sent == nil {
		return out
	}
	for _, tr := range sent.GetTokensWithoutWhitespace() {
		if tr != nil {
			out = append(out, tr.GetToken())
		}
	}
	return out
}
