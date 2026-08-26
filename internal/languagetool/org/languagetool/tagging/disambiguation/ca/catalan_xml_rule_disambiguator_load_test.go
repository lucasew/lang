package ca

// Outcome twins for Catalan XmlRuleDisambiguator as used by CatalanHybridDisambiguator:
// Java new XmlRuleDisambiguator(lang, true) — ca/disambiguation.xml then disambiguation-global.xml.
// Official catalan.dict is not required: use token-built sentences + IsIgnoredBySpeller
// for rules that match on token text/regexp (no POS dict needed).

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireCAXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverCatalanDisambiguationXML() == "" {
		t.Skip("official ca/disambiguation.xml not discoverable")
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

func TestDiscoverCatalanDisambiguationXML(t *testing.T) {
	p := DiscoverCatalanDisambiguationXML()
	if p == "" {
		t.Skip("official ca/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "ca")
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

func TestCatalanXmlRuleDisambiguator_LoadsOfficialPacks(t *testing.T) {
	requireCAXmlResources(t)

	caPath := DiscoverCatalanDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	caCount := countRulesFromXML(t, caPath, "ca")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, caCount, 0, "ca pack must load rules")
	require.Greater(t, globalCount, 0, "global pack must load rules")
	// Official XML has ~1973 ca <rule> elements; loader yields many expanded rules.
	require.GreaterOrEqual(t, caCount, 1000, "ca pack ~1000+ expanded rules")
	require.GreaterOrEqual(t, globalCount, 50, "global pack ~65 rules")

	xml := CatalanXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=true: total = ca + global (deterministic append order).
	require.Equal(t, caCount+globalCount, len(xml.Rules),
		"total rules must be ca+global (Java language XML then global)")
	require.NotNil(t, xml.UnifierConfig)

	// Process-cache singleton
	require.Same(t, xml, CatalanXmlRuleDisambiguator())
}

func TestNewCatalanHybridDisambiguator_WiresXmlRules(t *testing.T) {
	requireCAXmlResources(t)

	xml := CatalanXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewCatalanHybridDisambiguator()
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(lang, true)")
	require.Same(t, xml, d.Rules, "Rules field is process-cached CatalanXmlRuleDisambiguator")
}

// --- ca/disambiguation.xml ignore_spelling outcomes (token-built) ----------

func TestCatalanXmlRule_HAHAHA(t *testing.T) {
	requireCAXmlResources(t)
	xml := CatalanXmlRuleDisambiguator()
	// HAHAHA: regexp ha(ha)+|he(he)+|hi(hi)+ → ignore_spelling
	sent := xml.Disambiguate(tokenSentence("hahaha"))
	requireIgnored(t, sent, "hahaha")
	sent = xml.Disambiguate(tokenSentence("hehehe"))
	requireIgnored(t, sent, "hehehe")
}

func TestCatalanXmlRule_MesInfo(t *testing.T) {
	requireCAXmlResources(t)
	xml := CatalanXmlRuleDisambiguator()
	// mes_info: més|+ + marker(info) → ignore_spelling on info only
	sent := xml.Disambiguate(tokenSentence("més", "info"))
	requireNotIgnored(t, sent, "més")
	requireIgnored(t, sent, "info")
	sent = xml.Disambiguate(tokenSentence("+", "info"))
	requireNotIgnored(t, sent, "+")
	requireIgnored(t, sent, "info")
}

func TestCatalanXmlRule_MADE_IN(t *testing.T) {
	requireCAXmlResources(t)
	xml := CatalanXmlRuleDisambiguator()
	// MADE_IN: made + in + USA|China|… (case_sensitive country) → ignore on marker (all three)
	sent := xml.Disambiguate(tokenSentence("made", "in", "USA"))
	requireIgnored(t, sent, "made", "in", "USA")
}

func TestCatalanXmlRule_IGNORE_SPELLING_LE(t *testing.T) {
	requireCAXmlResources(t)
	xml := CatalanXmlRuleDisambiguator()
	// IGNORE_SPELLING_LE: marker(le) + nozze|jour|… → ignore on le only
	sent := xml.Disambiguate(tokenSentence("le", "nozze"))
	requireIgnored(t, sent, "le")
	requireNotIgnored(t, sent, "nozze")
	sent = xml.Disambiguate(tokenSentence("le", "journal"))
	requireIgnored(t, sent, "le")
}

func TestCatalanXmlRule_CANegative(t *testing.T) {
	requireCAXmlResources(t)
	xml := CatalanXmlRuleDisambiguator()
	// Random surface not matching ignore_spelling rules on text alone.
	const junk = "xyzzyqwertynotacaword"
	sent := xml.Disambiguate(tokenSentence(junk))
	requireNotIgnored(t, sent, junk)
}

// --- disambiguation-global.xml (proves useGlobal=true) ---------------------

func TestCatalanXmlRule_GlobalProperNouns(t *testing.T) {
	requireCAXmlResources(t)
	xml := CatalanXmlRuleDisambiguator()
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

func TestCatalanXmlRule_GlobalScientificNames_HomoSpp(t *testing.T) {
	requireCAXmlResources(t)
	xml := CatalanXmlRuleDisambiguator()
	// GLOBAL_SCIENTIFIC_NAMES: Homo + spp (no marker → whole match ignored).
	sent := xml.Disambiguate(tokenSentence("Homo", "spp"))
	requireIgnored(t, sent, "Homo", "spp")
}

func TestCatalanXmlRule_GlobalOnlyPatternNotInCAAlone(t *testing.T) {
	// Load ca-only pack and confirm global-only surface is not ignored; full
	// CatalanXml (useGlobal=true) ignores it — proves global was appended.
	requireCAXmlResources(t)

	caPath := DiscoverCatalanDisambiguationXML()
	f, err := os.Open(caPath)
	require.NoError(t, err)
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ca", caPath)
	f.Close()
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	caOnly := disambigrules.NewXmlRuleDisambiguator(rules)
	caOnly.UnifierConfig = uni

	// Literal QB|LT and Rem Koolhaas exist only in disambiguation-global.xml.
	caSent := caOnly.Disambiguate(tokenSentence("QB|LT"))
	requireNotIgnored(t, caSent, "QB|LT")
	caSent = caOnly.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, caSent, "Koolhaas")

	full := CatalanXmlRuleDisambiguator()
	fullSent := full.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, fullSent, "QB|LT")
	fullSent = full.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireIgnored(t, fullSent, "Koolhaas")
}

func TestCatalanXmlRule_FullPipelineViaNewCatalanHybridDisambiguator(t *testing.T) {
	requireCAXmlResources(t)
	d := NewCatalanHybridDisambiguator()
	require.NotNil(t, d.Rules)

	// ca rule (HAHAHA)
	sent := d.Disambiguate(tokenSentence("hahaha"))
	requireIgnored(t, sent, "hahaha")

	// global-only rule (literal surface from GLOBAL_PROPER_NOUNS)
	sent = d.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")
}
