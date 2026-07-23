package fr

// Outcome twins for French XmlRuleDisambiguator as used by FrenchHybridDisambiguator:
// Java new XmlRuleDisambiguator(lang, true) — fr/disambiguation.xml then disambiguation-global.xml.
// Official french.dict is not required: use token-built sentences + IsIgnoredBySpeller
// for rules that match on token text/regexp (no POS dict needed).

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireFRXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverFrenchDisambiguationXML() == "" {
		t.Skip("official fr/disambiguation.xml not discoverable")
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

func TestDiscoverFrenchDisambiguationXML(t *testing.T) {
	p := DiscoverFrenchDisambiguationXML()
	if p == "" {
		t.Skip("official fr/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "fr")
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

func TestFrenchXmlRuleDisambiguator_LoadsOfficialPacks(t *testing.T) {
	requireFRXmlResources(t)

	frPath := DiscoverFrenchDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	frCount := countRulesFromXML(t, frPath, "fr")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, frCount, 0, "fr pack must load rules")
	require.Greater(t, globalCount, 0, "global pack must load rules")
	// Official XML has ~1030 fr <rule> elements; loader yields many expanded rules.
	require.GreaterOrEqual(t, frCount, 500, "fr pack ~500+ expanded rules")
	require.GreaterOrEqual(t, globalCount, 50, "global pack ~65 rules")

	xml := FrenchXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=true: total = fr + global (deterministic append order).
	require.Equal(t, frCount+globalCount, len(xml.Rules),
		"total rules must be fr+global (Java language XML then global)")
	require.NotNil(t, xml.UnifierConfig)

	// Process-cache singleton
	require.Same(t, xml, FrenchXmlRuleDisambiguator())
}

func TestNewFrenchHybridDisambiguator_WiresXmlRules(t *testing.T) {
	requireFRXmlResources(t)

	xml := FrenchXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewFrenchHybridDisambiguator()
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(lang, true)")
	require.Same(t, xml, d.Rules, "Rules field is process-cached FrenchXmlRuleDisambiguator")
}

// --- fr/disambiguation.xml ignore_spelling outcomes (token-built) ----------
// From rulegroup IGNORE_SPELLING_OF_NUMBERS (official fr/disambiguation.xml).

func TestFrenchXmlRule_A4_B4(t *testing.T) {
	requireFRXmlResources(t)
	xml := FrenchXmlRuleDisambiguator()
	// [A-Z]\d+ → A4, B4 (and [A-Z]+\d+ also matches)
	sent := xml.Disambiguate(tokenSentence("A4"))
	requireIgnored(t, sent, "A4")
	sent = xml.Disambiguate(tokenSentence("B4"))
	requireIgnored(t, sent, "B4")
}

func TestFrenchXmlRule_5e(t *testing.T) {
	requireFRXmlResources(t)
	xml := FrenchXmlRuleDisambiguator()
	// \d+[eᵉ]\-? → 5e
	sent := xml.Disambiguate(tokenSentence("5e"))
	requireIgnored(t, sent, "5e")
}

func TestFrenchXmlRule_TimeRegexp(t *testing.T) {
	requireFRXmlResources(t)
	xml := FrenchXmlRuleDisambiguator()
	// ([01]?\d|2[0-3])h[0-5]?\d(min|m)([0-5]?\ds)? — whole-token surface match
	// e.g. 14h30min (hours + minutes with min suffix)
	sent := xml.Disambiguate(tokenSentence("14h30min"))
	requireIgnored(t, sent, "14h30min")
	// [0-5]?\d(min|m)[0-5]?\ds → e.g. 5min30s
	sent = xml.Disambiguate(tokenSentence("5min30s"))
	requireIgnored(t, sent, "5min30s")
}

func TestFrenchXmlRule_4x4(t *testing.T) {
	requireFRXmlResources(t)
	xml := FrenchXmlRuleDisambiguator()
	// \d+x\d+ multiplication
	sent := xml.Disambiguate(tokenSentence("4x4"))
	requireIgnored(t, sent, "4x4")
}

func TestFrenchXmlRule_FRNegative(t *testing.T) {
	requireFRXmlResources(t)
	xml := FrenchXmlRuleDisambiguator()
	// Random surface not matching ignore_spelling rules on text alone.
	const junk = "xyzzyqwertynotfrword"
	sent := xml.Disambiguate(tokenSentence(junk))
	requireNotIgnored(t, sent, junk)
}

// --- disambiguation-global.xml (proves useGlobal=true) ---------------------

func TestFrenchXmlRule_GlobalProperNouns(t *testing.T) {
	requireFRXmlResources(t)
	xml := FrenchXmlRuleDisambiguator()
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

func TestFrenchXmlRule_GlobalScientificNames_HomoSpp(t *testing.T) {
	requireFRXmlResources(t)
	xml := FrenchXmlRuleDisambiguator()
	// GLOBAL_SCIENTIFIC_NAMES: Homo + spp (no marker → whole match ignored).
	sent := xml.Disambiguate(tokenSentence("Homo", "spp"))
	requireIgnored(t, sent, "Homo", "spp")
}

func TestFrenchXmlRule_GlobalOnlyPatternNotInFRAlone(t *testing.T) {
	// Load fr-only pack and confirm global-only surface is not ignored; full
	// FrenchXml (useGlobal=true) ignores it — proves global was appended.
	requireFRXmlResources(t)

	frPath := DiscoverFrenchDisambiguationXML()
	f, err := os.Open(frPath)
	require.NoError(t, err)
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "fr", frPath)
	f.Close()
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	frOnly := disambigrules.NewXmlRuleDisambiguator(rules)
	frOnly.UnifierConfig = uni

	// Literal QB|LT and Rem Koolhaas exist only in disambiguation-global.xml.
	frSent := frOnly.Disambiguate(tokenSentence("QB|LT"))
	requireNotIgnored(t, frSent, "QB|LT")
	frSent = frOnly.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, frSent, "Koolhaas")

	full := FrenchXmlRuleDisambiguator()
	fullSent := full.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, fullSent, "QB|LT")
	fullSent = full.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireIgnored(t, fullSent, "Koolhaas")
}

func TestFrenchXmlRule_FullPipelineViaNewFrenchHybridDisambiguator(t *testing.T) {
	requireFRXmlResources(t)
	d := NewFrenchHybridDisambiguator()
	require.NotNil(t, d.Rules)

	// fr rule (A4 from IGNORE_SPELLING_OF_NUMBERS)
	sent := d.Disambiguate(tokenSentence("A4"))
	requireIgnored(t, sent, "A4")

	// global-only rule (literal surface from GLOBAL_PROPER_NOUNS)
	sent = d.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")
}
