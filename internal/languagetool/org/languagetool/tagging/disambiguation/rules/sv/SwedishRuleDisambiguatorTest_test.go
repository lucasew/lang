package sv

// Outcome twins for Swedish XmlRuleDisambiguator as used by SwedishHybridDisambiguator:
// Java new XmlRuleDisambiguator(new Swedish()) with useGlobalDisambiguation=false.
// Official pack (resource/sv/disambiguation.xml) is almost entirely immunize + ignore_spelling
// (~32 rules). Primary fidelity bar: matched tokens IsImmunized / IsIgnoredBySpeller;
// readings typically unchanged. Same bar as EO/BR/KM with real SwedishTagger + WordTokenizer.
// Upstream has no dedicated Swedish XmlRule unit test.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigsv "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/sv"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	disambigxx "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/xx"
	tagsv "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/sv"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// loadSVXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Swedish())
// (useGlobalDisambiguation default false) over official resource/sv/disambiguation.xml.
func loadSVXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	// Prefer process cache (disambiguation/sv loader); fall back to discover for isolation.
	if x := disambigsv.SwedishXmlRuleDisambiguator(); x != nil {
		return x
	}
	p := discoverSVDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "sv", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

func discoverSVDisambiguationXML() string {
	if p := os.Getenv("LANG_SV_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sv",
		"src", "main", "resources", "org", "languagetool", "resource", "sv", "disambiguation.xml")
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

func setupSVDisambiguation(t *testing.T) (demo disambiguation.Disambiguator, xml *disambigrules.XmlRuleDisambiguator) {
	t.Helper()
	if tagsv.DiscoverSwedishPOSDict() == "" {
		t.Skip("swedish.dict not in tree")
	}
	tagsv.EnsureDefaultSwedishTagger()
	require.NotNil(t, tagsv.DefaultSwedishTagger)
	require.NotNil(t, tagsv.DefaultSwedishTagger.GetWordTagger())
	require.NotEmpty(t, tagsv.SwedishPOSDictPath(), "real swedish.dict must load")

	xml = loadSVXmlRuleDisambiguator()
	if xml == nil || len(xml.Rules) == 0 {
		t.Skip("sv/disambiguation.xml not in tree or failed to load")
	}
	// Official SV pack: ~32 rules (immunize + ignore_spelling families).
	require.GreaterOrEqual(t, len(xml.Rules), 30)
	require.LessOrEqual(t, len(xml.Rules), 40)
	return disambigxx.NewDemoDisambiguator(), xml
}

// analyzeSV builds an AnalyzedSentence like SwedishDisambiguationRuleTest:
// WordTokenizer + SRXSentenceTokenizer(sv) + SwedishTagger + optional disambiguator.
// Returns the first sentence only (tests use single-sentence inputs).
func analyzeSV(input string, dis disambiguation.Disambiguator) *languagetool.AnalyzedSentence {
	tagsv.EnsureDefaultSwedishTagger()
	tagger := tagsv.DefaultSwedishTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("sv")
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

func requireImmunized(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	toks := wordTokens(sent)
	for _, s := range surfaces {
		tr, ok := toks[s]
		require.True(t, ok, "token %q missing in sentence", s)
		require.True(t, tr.IsImmunized(), "%q must be immunized", s)
	}
}

func requireNotImmunized(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	toks := wordTokens(sent)
	for _, s := range surfaces {
		tr, ok := toks[s]
		require.True(t, ok, "token %q missing in sentence", s)
		require.False(t, tr.IsImmunized(), "%q must not be immunized", s)
	}
}

func requireIgnoredBySpeller(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	toks := wordTokens(sent)
	for _, s := range surfaces {
		tr, ok := toks[s]
		require.True(t, ok, "token %q missing in sentence", s)
		require.True(t, tr.IsIgnoredBySpeller(), "%q must be ignore_spelling", s)
	}
}

func requireNotIgnoredBySpeller(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	toks := wordTokens(sent)
	for _, s := range surfaces {
		tr, ok := toks[s]
		require.True(t, ok, "token %q missing in sentence", s)
		require.False(t, tr.IsIgnoredBySpeller(), "%q must not be ignore_spelling", s)
	}
}

// --- Loader / pack shape ----------------------------------------------------

func TestSwedishXmlRuleDisambiguator_LoadsOfficialPack(t *testing.T) {
	_, xmlDisam := setupSVDisambiguation(t)
	// Official file has 32 top-level <rule> entries (count via id=).
	require.Equal(t, 32, len(xmlDisam.Rules), "official SV pack rule count")
	// useGlobal=false: only language XML, no disambiguation-global.xml.
	ids := make(map[string]bool, len(xmlDisam.Rules))
	for _, r := range xmlDisam.Rules {
		require.NotNil(t, r)
		ids[r.GetID()] = true
	}
	// Spot-check immunize + ignore_spelling families from official XML.
	for _, id := range []string{
		"SAMBAL_OELEK", "PULLED_PORK", "EARL_GREY", "LOS_ANGELES", "LAS_VEGAS",
		"RESISTANCE", "SEKUNDERMINUTER", "RHYTHM_AND_BLUES", "WOW", "WTC",
		"QUID_PRO_QUO", "AD_HOC", "PERSONA_NON_GRATA", "EL_NINO",
	} {
		require.True(t, ids[id], "missing rule id %s", id)
	}
}

// --- immunize family --------------------------------------------------------

// SAMBAL_OELEK: case-sensitive "Sambal Oelek" → both immunized.
func TestSwedishRuleDisambiguator_SambalOelekImmunize(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)
	const input = "Sambal Oelek är stark."

	demoSent := analyzeSV(input, demo)
	requireNotImmunized(t, demoSent, "Sambal", "Oelek", "är", "stark")

	xmlSent := analyzeSV(input, xmlDisam)
	requireImmunized(t, xmlSent, "Sambal", "Oelek")
	// Control words outside the match must stay clean.
	requireNotImmunized(t, xmlSent, "är", "stark")

	// case_sensitive=yes: lowercase must not immunize.
	low := analyzeSV("sambal oelek är stark.", xmlDisam)
	requireNotImmunized(t, low, "sambal", "oelek")
}

// PULLED_PORK + EARL_GREY immunize (case-sensitive).
func TestSwedishRuleDisambiguator_PulledPorkAndEarlGreyImmunize(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)

	// Pulled Pork
	const pork = "Pulled Pork serveras."
	requireNotImmunized(t, analyzeSV(pork, demo), "Pulled", "Pork")
	xmlPork := analyzeSV(pork, xmlDisam)
	requireImmunized(t, xmlPork, "Pulled", "Pork")
	requireNotImmunized(t, xmlPork, "serveras")
	requireNotImmunized(t, analyzeSV("pulled pork serveras.", xmlDisam), "pulled", "pork")

	// Earl Grey
	const tea = "Earl Grey te."
	requireNotImmunized(t, analyzeSV(tea, demo), "Earl", "Grey")
	xmlTea := analyzeSV(tea, xmlDisam)
	requireImmunized(t, xmlTea, "Earl", "Grey")
	requireNotImmunized(t, xmlTea, "te")
}

// LOS_ANGELES: "Los Angeles" / "Los Alamos" (case-insensitive pattern) → immunize.
func TestSwedishRuleDisambiguator_LosAngelesImmunize(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)

	for _, input := range []string{"Los Angeles", "Los Alamos"} {
		demoSent := analyzeSV(input, demo)
		xmlSent := analyzeSV(input, xmlDisam)
		// Extract first two non-whitespace word tokens.
		nws := xmlSent.GetTokensWithoutWhitespace()
		// index 0 is SENT_START
		require.GreaterOrEqual(t, len(nws), 3)
		los, city := nws[1], nws[2]
		require.Equal(t, "Los", los.GetToken())
		require.False(t, demoSent.GetTokensWithoutWhitespace()[1].IsImmunized())
		require.True(t, los.IsImmunized(), "Los immunized for %q", input)
		require.True(t, city.IsImmunized(), "%q immunized for %q", city.GetToken(), input)
	}

	// Non-matching second token stays clean.
	miss := analyzeSV("Los Santos", xmlDisam)
	requireNotImmunized(t, miss, "Los", "Santos")
}

// LAS_VEGAS: case-sensitive Las Vegas / Las Ramblas.
func TestSwedishRuleDisambiguator_LasVegasImmunize(t *testing.T) {
	_, xmlDisam := setupSVDisambiguation(t)

	vegas := analyzeSV("Las Vegas", xmlDisam)
	requireImmunized(t, vegas, "Las", "Vegas")

	ramblas := analyzeSV("Las Ramblas", xmlDisam)
	requireImmunized(t, ramblas, "Las", "Ramblas")

	// case_sensitive: lowercase must not match.
	low := analyzeSV("las vegas", xmlDisam)
	requireNotImmunized(t, low, "las", "vegas")
}

// --- ignore_spelling family -------------------------------------------------

// RESISTANCE: "pièce de résistance" → ignore_spelling on all three tokens.
func TestSwedishRuleDisambiguator_PieceDeResistanceIgnore(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)
	const input = "pièce de résistance"
	demoSent := analyzeSV(input, demo)
	requireNotIgnoredBySpeller(t, demoSent, "pièce", "de", "résistance")

	xmlSent := analyzeSV(input, xmlDisam)
	requireIgnoredBySpeller(t, xmlSent, "pièce", "de", "résistance")
	// Immunize flag is separate — this rule is ignore_spelling only.
	requireNotImmunized(t, xmlSent, "pièce", "de", "résistance")
}

// SEKUNDERMINUTER: 60-sekunder / 30-minuters → ignore_spelling.
func TestSwedishRuleDisambiguator_SekunderMinuterIgnore(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)

	for _, input := range []string{"60-sekunder", "30-minuters", "5-minuter"} {
		demoSent := analyzeSV(input, demo)
		xmlSent := analyzeSV(input, xmlDisam)
		// WordTokenizer keeps hyphenated form as one token when digits+letters.
		nws := xmlSent.GetTokensWithoutWhitespace()
		require.GreaterOrEqual(t, len(nws), 2, "input=%q tokens=%v", input, tokenSurfaces(xmlSent))
		tok := nws[1]
		require.Equal(t, input, tok.GetToken(), "expected single token for %q got %v", input, tokenSurfaces(xmlSent))
		require.False(t, demoSent.GetTokensWithoutWhitespace()[1].IsIgnoredBySpeller())
		require.True(t, tok.IsIgnoredBySpeller(), "ignore_spelling for %q", input)
	}

	// Control: plain number not matched.
	ctrl := analyzeSV("60 sekunder", xmlDisam)
	requireNotIgnoredBySpeller(t, ctrl, "60", "sekunder")
}

// RHYTHM_AND_BLUES + ROCK_AND_ROLL multi-token ignore_spelling.
func TestSwedishRuleDisambiguator_RhythmAndRockIgnore(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)

	const rhythm = "Rhythm and Blues"
	requireNotIgnoredBySpeller(t, analyzeSV(rhythm, demo), "Rhythm", "and", "Blues")
	xmlR := analyzeSV(rhythm, xmlDisam)
	requireIgnoredBySpeller(t, xmlR, "Rhythm", "and", "Blues")

	const rock = "Rock and Roll"
	xmlRock := analyzeSV(rock, xmlDisam)
	requireIgnoredBySpeller(t, xmlRock, "Rock", "and", "Roll")
}

// WOW + WTC: World of Warcraft / World Trade Center(s), case-sensitive.
func TestSwedishRuleDisambiguator_WorldPhrasesIgnore(t *testing.T) {
	_, xmlDisam := setupSVDisambiguation(t)

	wow := analyzeSV("World of Warcraft", xmlDisam)
	requireIgnoredBySpeller(t, wow, "World", "of", "Warcraft")

	wtc := analyzeSV("World Trade Center", xmlDisam)
	requireIgnoredBySpeller(t, wtc, "World", "Trade", "Center")

	wtcs := analyzeSV("World Trade Centers", xmlDisam)
	requireIgnoredBySpeller(t, wtcs, "World", "Trade", "Centers")

	// case_sensitive=yes on WOW/WTC.
	low := analyzeSV("world of warcraft", xmlDisam)
	requireNotIgnoredBySpeller(t, low, "world", "of", "warcraft")
}

// Latin / Romance phrase ignore_spelling sample.
func TestSwedishRuleDisambiguator_LatinPhrasesIgnore(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)

	cases := []struct {
		input    string
		surfaces []string
	}{
		{"quid pro quo", []string{"quid", "pro", "quo"}},
		{"rigor mortis", []string{"rigor", "mortis"}},
		{"carpe diem", []string{"carpe", "diem"}},
		{"tabula rasa", []string{"tabula", "rasa"}},
		{"vox populi", []string{"vox", "populi"}},
		{"coitus interruptus", []string{"coitus", "interruptus"}},
		{"summa summarum", []string{"summa", "summarum"}},
		{"modus operandi", []string{"modus", "operandi"}},
		{"ad hoc", []string{"ad", "hoc"}},
		{"ad infinitum", []string{"ad", "infinitum"}},
		{"ad hominem", []string{"ad", "hominem"}},
		{"argumentum ad populum", []string{"argumentum", "ad", "populum"}},
		{"ius primae noctis", []string{"ius", "primae", "noctis"}},
		{"terra incognita", []string{"terra", "incognita"}},
		{"terra nova", []string{"terra", "nova"}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			demoSent := analyzeSV(tc.input, demo)
			xmlSent := analyzeSV(tc.input, xmlDisam)
			for _, s := range tc.surfaces {
				require.False(t, wordTokens(demoSent)[s].IsIgnoredBySpeller(), "demo %q", s)
			}
			requireIgnoredBySpeller(t, xmlSent, tc.surfaces...)
		})
	}
}

// Place / brand ignore_spelling: São Paulo, Addis Abeba, La Paz, Santo Domingo, San Bernardino, El Niño, Enfant terrible, Delirium tremens.
func TestSwedishRuleDisambiguator_PlaceAndBrandIgnore(t *testing.T) {
	_, xmlDisam := setupSVDisambiguation(t)

	cases := []struct {
		input    string
		surfaces []string
	}{
		{"São Paulo", []string{"São", "Paulo"}},
		{"Addis Abeba", []string{"Addis", "Abeba"}},
		{"La Paz", []string{"La", "Paz"}},
		{"Santo Domingo", []string{"Santo", "Domingo"}},
		{"San Bernardino", []string{"San", "Bernardino"}},
		{"El Niño", []string{"El", "Niño"}},
		{"Enfant terrible", []string{"Enfant", "terrible"}},
		{"Delirium tremens", []string{"Delirium", "tremens"}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			xmlSent := analyzeSV(tc.input, xmlDisam)
			requireIgnoredBySpeller(t, xmlSent, tc.surfaces...)
		})
	}

	// LA_PAZ / SAN_BERNARDINO / DELIRIUM are case-sensitive where marked.
	requireNotIgnoredBySpeller(t, analyzeSV("la paz", xmlDisam), "la", "paz")
	requireNotIgnoredBySpeller(t, analyzeSV("san bernardino", xmlDisam), "san", "bernardino")
	requireNotIgnoredBySpeller(t, analyzeSV("delirium tremens", xmlDisam), "delirium", "tremens")
}

// PERSONA_NON_GRATA: optional "non" (min=0) — both "Persona grata" and "Persona non grata".
func TestSwedishRuleDisambiguator_PersonaNonGrataIgnore(t *testing.T) {
	_, xmlDisam := setupSVDisambiguation(t)

	with := analyzeSV("Persona non grata", xmlDisam)
	requireIgnoredBySpeller(t, with, "Persona", "non", "grata")

	without := analyzeSV("Persona grata", xmlDisam)
	requireIgnoredBySpeller(t, without, "Persona", "grata")

	// ingrata variant
	ing := analyzeSV("Persona non ingrata", xmlDisam)
	requireIgnoredBySpeller(t, ing, "Persona", "non", "ingrata")
}

// Non-matched control sentence: common Swedish text remains not immunized / not ignore-spelling.
func TestSwedishRuleDisambiguator_ControlUnmatchedStaysClean(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)
	const input = "Att testa disambiguering är kul."
	demoSent := analyzeSV(input, demo)
	xmlSent := analyzeSV(input, xmlDisam)

	// Readings / flags identical for demo vs xml on this control (no pack rules fire).
	require.Equal(t, myAssertDisambiguate(input, demo), myAssertDisambiguate(input, xmlDisam))

	for _, tr := range xmlSent.GetTokensWithoutWhitespace() {
		if tr == nil || tr.GetToken() == "" {
			continue
		}
		require.False(t, tr.IsImmunized(), "unexpected immunize on %q", tr.GetToken())
		require.False(t, tr.IsIgnoredBySpeller(), "unexpected ignore_spelling on %q", tr.GetToken())
	}
	// demo also clean
	for _, tr := range demoSent.GetTokensWithoutWhitespace() {
		if tr == nil {
			continue
		}
		require.False(t, tr.IsImmunized())
		require.False(t, tr.IsIgnoredBySpeller())
	}
}

// Hybrid Rules stage uses the same official XML (Java eager XmlRuleDisambiguator field).
func TestSwedishHybridDisambiguator_RulesStageMatchesXml(t *testing.T) {
	_, xmlDisam := setupSVDisambiguation(t)
	hybrid := disambigsv.NewSwedishHybridDisambiguator()
	require.NotNil(t, hybrid.Rules, "Java constructs XmlRuleDisambiguator eagerly")

	const input = "Sambal Oelek och pièce de résistance"
	require.Equal(t,
		myAssertDisambiguate(input, xmlDisam),
		myAssertDisambiguate(input, hybrid),
		"hybrid Rules stage readings == standalone XmlRuleDisambiguator")

	// Flags too: immunize + ignore_spelling via hybrid Rules only (Chunker nil).
	hs := analyzeSV(input, hybrid)
	requireImmunized(t, hs, "Sambal", "Oelek")
	requireIgnoredBySpeller(t, hs, "pièce", "de", "résistance")
	// Hybrid Rules pointer is the process-cached instance.
	require.Same(t, xmlDisam, hybrid.Rules)
}

// Immunize does not rewrite readings (myAssert strings equal demo for immunize-only phrases
// when tokens have no other pack side-effects). Readings may still be null-tagged.
func TestSwedishRuleDisambiguator_ImmunizePreservesReadings(t *testing.T) {
	demo, xmlDisam := setupSVDisambiguation(t)
	// Use only-immunize phrase with no ignore_spelling sibling match.
	const input = "Sambal Oelek"
	// Reading strings should match (immunize is a flag, not a POS rewrite).
	require.Equal(t,
		myAssertDisambiguate(input, demo),
		myAssertDisambiguate(input, xmlDisam),
		"immunize must not alter reading strings")
	// But flags differ.
	requireImmunized(t, analyzeSV(input, xmlDisam), "Sambal", "Oelek")
	requireNotImmunized(t, analyzeSV(input, demo), "Sambal", "Oelek")
}

// --- myAssert helpers (parity with other language RuleDisambiguator tests) ---

// myAssertDisambiguate ports Java TestTools.myAssert(input, expected,
// WordTokenizer, SRXSentenceTokenizer(Swedish), SwedishTagger, disambiguator).
func myAssertDisambiguate(input string, dis disambiguation.Disambiguator) string {
	tagsv.EnsureDefaultSwedishTagger()
	tagger := tagsv.DefaultSwedishTagger
	wt := tokenizers.NewWordTokenizer()
	st := tokenizers.NewSRXSentenceTokenizer("sv")
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
