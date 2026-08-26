package nl

// Outcome twins for Dutch XmlRuleDisambiguator as used by DutchHybridDisambiguator:
// Java new XmlRuleDisambiguator(lang, true) — nl/disambiguation.xml then disambiguation-global.xml.
// Official dutch.dict is not required: use token-built sentences + IsIgnoredBySpeller
// for rules that match on token text/regexp (no POS dict needed).

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireNLXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverDutchDisambiguationXML() == "" {
		t.Skip("official nl/disambiguation.xml not discoverable")
	}
	if DiscoverGlobalDisambiguationXML() == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
}

// tokenSentence builds SENT_START + tokens with spaces between word tokens.
// Whitespace tokens are included for natural spacing; pattern match uses non-WS.
func tokenSentence(words ...string) *languagetool.AnalyzedSentence {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for i, w := range words {
		if i > 0 {
			toks = append(toks, languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(" ", nil, nil)))
		}
		toks = append(toks, languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(w, nil, nil)))
	}
	return languagetool.NewAnalyzedSentence(toks)
}

func tokenBySurface(sent *languagetool.AnalyzedSentence, surface string) *languagetool.AnalyzedTokenReadings {
	if sent == nil {
		return nil
	}
	for _, tr := range sent.GetTokensWithoutWhitespace() {
		if tr != nil && tr.GetToken() == surface {
			return tr
		}
	}
	return nil
}

func requireIgnored(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.True(t, tr.IsIgnoredBySpeller(), "%q must be ignore_spelling", s)
	}
}

func requireNotIgnored(t *testing.T, sent *languagetool.AnalyzedSentence, surfaces ...string) {
	t.Helper()
	for _, s := range surfaces {
		tr := tokenBySurface(sent, s)
		require.NotNil(t, tr, "token %q missing", s)
		require.False(t, tr.IsIgnoredBySpeller(), "%q must not be ignore_spelling", s)
	}
}

func countRulesFromXML(t *testing.T, path, langCode string) int {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	rules, _, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, langCode, path)
	require.NoError(t, err)
	return len(rules)
}

func TestDiscoverDutchDisambiguationXML(t *testing.T) {
	p := DiscoverDutchDisambiguationXML()
	if p == "" {
		t.Skip("official nl/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "nl")
	st, err := os.Stat(p)
	require.NoError(t, err)
	require.True(t, st.Mode().IsRegular())
}

func TestDiscoverGlobalDisambiguationXML(t *testing.T) {
	p := DiscoverGlobalDisambiguationXML()
	if p == "" {
		t.Skip("official disambiguation-global.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation-global.xml")
	st, err := os.Stat(p)
	require.NoError(t, err)
	require.True(t, st.Mode().IsRegular())
}

func TestDutchXmlRuleDisambiguator_LoadsOfficialPacks(t *testing.T) {
	requireNLXmlResources(t)

	nlPath := DiscoverDutchDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	nlCount := countRulesFromXML(t, nlPath, "nl")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, nlCount, 0, "nl pack must load rules")
	require.Greater(t, globalCount, 0, "global pack must load rules")
	// Official XML has ~892 nl <rule> and ~65 global <rule> elements (loader expands groups).
	require.GreaterOrEqual(t, nlCount, 800, "nl pack ~892 rules")
	require.GreaterOrEqual(t, globalCount, 50, "global pack ~65 rules")

	xml := DutchXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=true: total = nl + global (deterministic append order).
	require.Equal(t, nlCount+globalCount, len(xml.Rules),
		"total rules must be nl+global (Java language XML then global)")
	require.NotNil(t, xml.UnifierConfig)

	// Process-cache singleton
	require.Same(t, xml, DutchXmlRuleDisambiguator())
}

func TestNewDutchHybridDisambiguator_WiresXmlRules(t *testing.T) {
	requireNLXmlResources(t)

	xml := DutchXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewDutchHybridDisambiguator()
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(lang, true)")
	require.Same(t, xml, d.Rules, "Rules field is process-cached DutchXmlRuleDisambiguator")
}

// --- nl/disambiguation.xml ignore_spelling outcomes (token-built) ----------

func TestDutchXmlRule_ROADS(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// ROADS: regexp [AENR][0-9]{1,3}+ → ignore_spelling
	sent := xml.Disambiguate(tokenSentence("A12"))
	requireIgnored(t, sent, "A12")
	sent = xml.Disambiguate(tokenSentence("N201"))
	requireIgnored(t, sent, "N201")
}

func TestDutchXmlRule_PLANES(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// PLANES: case_sensitive regexp PH-[A-Z]{3}
	sent := xml.Disambiguate(tokenSentence("PH-ABC"))
	requireIgnored(t, sent, "PH-ABC")
}

func TestDutchXmlRule_NUMBER_UNITS(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// NUMBER_UNITS: [1-9][0-9]+(m2|…|km|…) — needs ≥2 digits
	sent := xml.Disambiguate(tokenSentence("10km"))
	requireIgnored(t, sent, "10km")
}

func TestDutchXmlRule_IGNORE_SPELLER_ROMAN_NUMBERS(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// IGNORE_SPELLER_ROMAN_NUMBERS: case_sensitive roman numeral regex
	sent := xml.Disambiguate(tokenSentence("XIV"))
	requireIgnored(t, sent, "XIV")
}

func TestDutchXmlRule_COMMISSIE_NAAM(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// COMMISSIE_NAAM: (regering|commissie|wet|college|zaak)-[A-Z][a-z].*
	sent := xml.Disambiguate(tokenSentence("commissie-Foo"))
	requireIgnored(t, sent, "commissie-Foo")
}

func TestDutchXmlRule_NLNegative(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// Random surface not matching ignore_spelling rules on text alone.
	const junk = "xyzzyqwertynotadeword"
	sent := xml.Disambiguate(tokenSentence(junk))
	requireNotIgnored(t, sent, junk)
}

// --- disambiguation-global.xml (proves useGlobal=true) ---------------------

func TestDutchXmlRule_GlobalProperNouns(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// Official GLOBAL_PROPER_NOUNS: token is literal "QB|LT" (no regexp="yes"),
	// case_sensitive — matches the single surface QB|LT, not QB or LT alone.
	sent := xml.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")
	// Rem + Koolhaas: marker on Koolhaas only (skip=1 on first token).
	sent = xml.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, sent, "Rem")
	requireIgnored(t, sent, "Koolhaas")
	// case_sensitive=yes: lowercase must not match
	sent = xml.Disambiguate(tokenSentence("qb|lt"))
	requireNotIgnored(t, sent, "qb|lt")
}

func TestDutchXmlRule_GlobalScientificNames_HomoSpp(t *testing.T) {
	requireNLXmlResources(t)
	xml := DutchXmlRuleDisambiguator()
	// GLOBAL_SCIENTIFIC_NAMES: Homo + spp (no marker → whole match ignored).
	sent := xml.Disambiguate(tokenSentence("Homo", "spp"))
	requireIgnored(t, sent, "Homo", "spp")
}

func TestDutchXmlRule_GlobalOnlyPatternNotInNLAlone(t *testing.T) {
	// Load nl-only pack and confirm global-only surface is not ignored; full
	// DutchXml (useGlobal=true) ignores it — proves global was appended.
	requireNLXmlResources(t)

	nlPath := DiscoverDutchDisambiguationXML()
	f, err := os.Open(nlPath)
	require.NoError(t, err)
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "nl", nlPath)
	f.Close()
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	nlOnly := disambigrules.NewXmlRuleDisambiguator(rules)
	nlOnly.UnifierConfig = uni

	// Literal QB|LT and Rem Koolhaas exist only in disambiguation-global.xml.
	nlSent := nlOnly.Disambiguate(tokenSentence("QB|LT"))
	requireNotIgnored(t, nlSent, "QB|LT")
	nlSent = nlOnly.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, nlSent, "Koolhaas")

	full := DutchXmlRuleDisambiguator()
	fullSent := full.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, fullSent, "QB|LT")
	fullSent = full.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireIgnored(t, fullSent, "Koolhaas")
}

func TestDutchXmlRule_FullPipelineViaNewDutchHybridDisambiguator(t *testing.T) {
	requireNLXmlResources(t)
	d := NewDutchHybridDisambiguator()
	require.NotNil(t, d.Rules)

	// nl rule
	sent := d.Disambiguate(tokenSentence("A12"))
	requireIgnored(t, sent, "A12")

	// global-only rule (literal surface from GLOBAL_PROPER_NOUNS)
	sent = d.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")
}
