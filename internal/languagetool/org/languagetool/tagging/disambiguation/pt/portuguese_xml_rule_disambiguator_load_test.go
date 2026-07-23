package pt

// Outcome twins for Portuguese XmlRuleDisambiguator as used by PortugueseHybridDisambiguator:
// Java new XmlRuleDisambiguator(lang, true) — pt/disambiguation.xml then disambiguation-global.xml.
// Official portuguese.dict is not required: use token-built sentences + IsIgnoredBySpeller
// for rules that match on token text/regexp (no POS dict needed).

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requirePTXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverPortugueseDisambiguationXML() == "" {
		t.Skip("official pt/disambiguation.xml not discoverable")
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

func TestDiscoverPortugueseDisambiguationXML(t *testing.T) {
	p := DiscoverPortugueseDisambiguationXML()
	if p == "" {
		t.Skip("official pt/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "pt")
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

func TestPortugueseXmlRuleDisambiguator_LoadsOfficialPacks(t *testing.T) {
	requirePTXmlResources(t)

	ptPath := DiscoverPortugueseDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	ptCount := countRulesFromXML(t, ptPath, "pt")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, ptCount, 0, "pt pack must load rules")
	require.Greater(t, globalCount, 0, "global pack must load rules")
	// Official XML has ~555 pt <rule> elements; loader yields hundreds of expanded rules.
	require.GreaterOrEqual(t, ptCount, 200, "pt pack hundreds of expanded rules")
	require.GreaterOrEqual(t, globalCount, 50, "global pack ~65 rules")

	xml := PortugueseXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=true: total = pt + global (deterministic append order).
	require.Equal(t, ptCount+globalCount, len(xml.Rules),
		"total rules must be pt+global (Java language XML then global)")
	require.NotNil(t, xml.UnifierConfig)

	// Process-cache singleton
	require.Same(t, xml, PortugueseXmlRuleDisambiguator())
}

func TestNewPortugueseHybridDisambiguator_WiresXmlRules(t *testing.T) {
	requirePTXmlResources(t)

	xml := PortugueseXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewPortugueseHybridDisambiguator()
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(lang, true)")
	require.Same(t, xml, d.Rules, "Rules field is process-cached PortugueseXmlRuleDisambiguator")
}

// --- pt/disambiguation.xml ignore_spelling outcomes (token-built) ----------

func TestPortugueseXmlRule_UNIVERSITY_OF(t *testing.T) {
	requirePTXmlResources(t)
	xml := PortugueseXmlRuleDisambiguator()
	// UNIVERSITY_OF: case_sensitive University + of → ignore_spelling on match
	sent := xml.Disambiguate(tokenSentence("University", "of"))
	requireIgnored(t, sent, "University", "of")
	// lowercase must not match (case_sensitive)
	sent = xml.Disambiguate(tokenSentence("university", "of"))
	requireNotIgnored(t, sent, "university")
}

func TestPortugueseXmlRule_FIFTH_AVENUE(t *testing.T) {
	requirePTXmlResources(t)
	xml := PortugueseXmlRuleDisambiguator()
	// FIFTH_AVENUE: Park|… + Avenue|Street
	sent := xml.Disambiguate(tokenSentence("Park", "Avenue"))
	requireIgnored(t, sent, "Park", "Avenue")
}

func TestPortugueseXmlRule_SIZES(t *testing.T) {
	requirePTXmlResources(t)
	xml := PortugueseXmlRuleDisambiguator()
	// SIZES (LOOSE_UPPERCASES): tamanhos? + X{0,2}[SLM]|…
	sent := xml.Disambiguate(tokenSentence("tamanho", "S"))
	requireIgnored(t, sent, "tamanho", "S")
}

func TestPortugueseXmlRule_RomanNumeralIgnore(t *testing.T) {
	requirePTXmlResources(t)
	xml := PortugueseXmlRuleDisambiguator()
	// ROMAN_NUMBER_IGNORE_SPELLING: case_sensitive uppercase Roman numerals e.g. XIV
	sent := xml.Disambiguate(tokenSentence("XIV"))
	requireIgnored(t, sent, "XIV")
	// lowercase alone may match other roman rules; use non-roman junk for negative below
}

func TestPortugueseXmlRule_PTNegative(t *testing.T) {
	requirePTXmlResources(t)
	xml := PortugueseXmlRuleDisambiguator()
	// Random surface not matching ignore_spelling rules on text alone.
	const junk = "xyzzyqwertynotaptword"
	sent := xml.Disambiguate(tokenSentence(junk))
	requireNotIgnored(t, sent, junk)
}

// --- disambiguation-global.xml (proves useGlobal=true) ---------------------

func TestPortugueseXmlRule_GlobalProperNouns(t *testing.T) {
	requirePTXmlResources(t)
	xml := PortugueseXmlRuleDisambiguator()
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

func TestPortugueseXmlRule_GlobalScientificNames_HomoSpp(t *testing.T) {
	requirePTXmlResources(t)
	xml := PortugueseXmlRuleDisambiguator()
	// GLOBAL_SCIENTIFIC_NAMES: Homo + spp (no marker → whole match ignored).
	sent := xml.Disambiguate(tokenSentence("Homo", "spp"))
	requireIgnored(t, sent, "Homo", "spp")
}

func TestPortugueseXmlRule_GlobalOnlyPatternNotInPTAlone(t *testing.T) {
	// Load pt-only pack and confirm global-only surface is not ignored; full
	// PortugueseXml (useGlobal=true) ignores it — proves global was appended.
	requirePTXmlResources(t)

	ptPath := DiscoverPortugueseDisambiguationXML()
	f, err := os.Open(ptPath)
	require.NoError(t, err)
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "pt", ptPath)
	f.Close()
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	ptOnly := disambigrules.NewXmlRuleDisambiguator(rules)
	ptOnly.UnifierConfig = uni

	// Literal QB|LT and Rem Koolhaas exist only in disambiguation-global.xml.
	ptSent := ptOnly.Disambiguate(tokenSentence("QB|LT"))
	requireNotIgnored(t, ptSent, "QB|LT")
	ptSent = ptOnly.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, ptSent, "Koolhaas")

	full := PortugueseXmlRuleDisambiguator()
	fullSent := full.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, fullSent, "QB|LT")
	fullSent = full.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireIgnored(t, fullSent, "Koolhaas")
}

func TestPortugueseXmlRule_FullPipelineViaNewPortugueseHybridDisambiguator(t *testing.T) {
	requirePTXmlResources(t)
	d := NewPortugueseHybridDisambiguator()
	require.NotNil(t, d.Rules)

	// pt rule (UNIVERSITY_OF)
	sent := d.Disambiguate(tokenSentence("University", "of"))
	requireIgnored(t, sent, "University", "of")

	// global-only rule (literal surface from GLOBAL_PROPER_NOUNS)
	sent = d.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")
}
