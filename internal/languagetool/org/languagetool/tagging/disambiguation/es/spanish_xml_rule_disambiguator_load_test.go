package es

// Outcome twins for Spanish XmlRuleDisambiguator as used by SpanishHybridDisambiguator:
// Java new XmlRuleDisambiguator(lang, true) — es/disambiguation.xml then disambiguation-global.xml.
// Official spanish.dict is not required: use token-built sentences + IsIgnoredBySpeller
// for rules that match on token text/regexp (no POS dict needed).

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireESXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverSpanishDisambiguationXML() == "" {
		t.Skip("official es/disambiguation.xml not discoverable")
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

func TestDiscoverSpanishDisambiguationXML(t *testing.T) {
	p := DiscoverSpanishDisambiguationXML()
	if p == "" {
		t.Skip("official es/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "es")
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

func TestSpanishXmlRuleDisambiguator_LoadsOfficialPacks(t *testing.T) {
	requireESXmlResources(t)

	esPath := DiscoverSpanishDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	esCount := countRulesFromXML(t, esPath, "es")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, esCount, 0, "es pack must load rules")
	require.Greater(t, globalCount, 0, "global pack must load rules")
	// Official XML has ~869 es <rule> elements; loader yields ~700+ expanded rules.
	require.GreaterOrEqual(t, esCount, 600, "es pack ~700+ expanded rules")
	require.GreaterOrEqual(t, globalCount, 50, "global pack ~65 rules")

	xml := SpanishXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=true: total = es + global (deterministic append order).
	require.Equal(t, esCount+globalCount, len(xml.Rules),
		"total rules must be es+global (Java language XML then global)")
	require.NotNil(t, xml.UnifierConfig)

	// Process-cache singleton
	require.Same(t, xml, SpanishXmlRuleDisambiguator())
}

func TestNewSpanishHybridDisambiguator_WiresXmlRules(t *testing.T) {
	requireESXmlResources(t)

	xml := SpanishXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewSpanishHybridDisambiguator()
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(lang, true)")
	require.Same(t, xml, d.Rules, "Rules field is process-cached SpanishXmlRuleDisambiguator")
}

// --- es/disambiguation.xml ignore_spelling outcomes (token-built) ----------

func TestSpanishXmlRule_ABBREVIATIONS_ShortDates(t *testing.T) {
	requireESXmlResources(t)
	xml := SpanishXmlRuleDisambiguator()
	// ABBREVIATIONS: case_sensitive regexp (\d|[12]\d|3[01])[EFMAJSOND] → e.g. 15E
	sent := xml.Disambiguate(tokenSentence("15E"))
	requireIgnored(t, sent, "15E")
	sent = xml.Disambiguate(tokenSentence("3J"))
	requireIgnored(t, sent, "3J")
	// lowercase must not match (case_sensitive)
	sent = xml.Disambiguate(tokenSentence("15e"))
	requireNotIgnored(t, sent, "15e")
}

func TestSpanishXmlRule_ABBREVIATIONS_CtrlAlt(t *testing.T) {
	requireESXmlResources(t)
	xml := SpanishXmlRuleDisambiguator()
	// Ctrl/Alt + shortcuts: marker on Ctrl|Alt when followed by +
	sent := xml.Disambiguate(tokenSentence("Ctrl", "+"))
	requireIgnored(t, sent, "Ctrl")
	// Second rule: Ctrl [+] key → ignore key (marker on last)
	sent = xml.Disambiguate(tokenSentence("Alt", "+", "F4"))
	requireIgnored(t, sent, "F4")
}

func TestSpanishXmlRule_UNIDADES_SI_Min(t *testing.T) {
	requireESXmlResources(t)
	xml := SpanishXmlRuleDisambiguator()
	// UNIDADES_SI: \d+ + min → ignore_spelling on min
	sent := xml.Disambiguate(tokenSentence("30", "min"))
	requireNotIgnored(t, sent, "30")
	requireIgnored(t, sent, "min")
}

func TestSpanishXmlRule_UNIDADES_SI_Degrees(t *testing.T) {
	requireESXmlResources(t)
	xml := SpanishXmlRuleDisambiguator()
	// UNIDADES_SI: [\d,\.]+º[\d,\.]+ → ignore_spelling (single token)
	sent := xml.Disambiguate(tokenSentence("1.5º3"))
	requireIgnored(t, sent, "1.5º3")
}

func TestSpanishXmlRule_ESNegative(t *testing.T) {
	requireESXmlResources(t)
	xml := SpanishXmlRuleDisambiguator()
	// Random surface not matching ignore_spelling rules on text alone.
	const junk = "xyzzyqwertynotadeword"
	sent := xml.Disambiguate(tokenSentence(junk))
	requireNotIgnored(t, sent, junk)
}

// --- disambiguation-global.xml (proves useGlobal=true) ---------------------

func TestSpanishXmlRule_GlobalProperNouns(t *testing.T) {
	requireESXmlResources(t)
	xml := SpanishXmlRuleDisambiguator()
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

func TestSpanishXmlRule_GlobalScientificNames_HomoSpp(t *testing.T) {
	requireESXmlResources(t)
	xml := SpanishXmlRuleDisambiguator()
	// GLOBAL_SCIENTIFIC_NAMES: Homo + spp (no marker → whole match ignored).
	sent := xml.Disambiguate(tokenSentence("Homo", "spp"))
	requireIgnored(t, sent, "Homo", "spp")
}

func TestSpanishXmlRule_GlobalOnlyPatternNotInESAlone(t *testing.T) {
	// Load es-only pack and confirm global-only surface is not ignored; full
	// SpanishXml (useGlobal=true) ignores it — proves global was appended.
	requireESXmlResources(t)

	esPath := DiscoverSpanishDisambiguationXML()
	f, err := os.Open(esPath)
	require.NoError(t, err)
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "es", esPath)
	f.Close()
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	esOnly := disambigrules.NewXmlRuleDisambiguator(rules)
	esOnly.UnifierConfig = uni

	// Literal QB|LT and Rem Koolhaas exist only in disambiguation-global.xml.
	esSent := esOnly.Disambiguate(tokenSentence("QB|LT"))
	requireNotIgnored(t, esSent, "QB|LT")
	esSent = esOnly.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, esSent, "Koolhaas")

	full := SpanishXmlRuleDisambiguator()
	fullSent := full.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, fullSent, "QB|LT")
	fullSent = full.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireIgnored(t, fullSent, "Koolhaas")
}

func TestSpanishXmlRule_FullPipelineViaNewSpanishHybridDisambiguator(t *testing.T) {
	requireESXmlResources(t)
	d := NewSpanishHybridDisambiguator()
	require.NotNil(t, d.Rules)

	// es rule (ABBREVIATIONS short date)
	sent := d.Disambiguate(tokenSentence("15E"))
	requireIgnored(t, sent, "15E")

	// global-only rule (literal surface from GLOBAL_PROPER_NOUNS)
	sent = d.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")
}
