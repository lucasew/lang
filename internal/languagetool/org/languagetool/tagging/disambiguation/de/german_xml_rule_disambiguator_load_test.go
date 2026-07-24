package de

// Outcome twins for German XmlRuleDisambiguator as used by GermanRuleDisambiguator:
// Java new XmlRuleDisambiguator(lang, true) — de/disambiguation.xml then disambiguation-global.xml.
// Official german.dict is not vendored: use token-built sentences + IsIgnoredBySpeller
// for rules that match on token text/regexp (no POS dict needed).

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireDEXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverGermanDisambiguationXML() == "" {
		t.Skip("official de/disambiguation.xml not discoverable")
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

func TestDiscoverGermanDisambiguationXML(t *testing.T) {
	p := DiscoverGermanDisambiguationXML()
	if p == "" {
		t.Skip("official de/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "de")
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

func TestGermanXmlRuleDisambiguator_LoadsOfficialPacks(t *testing.T) {
	requireDEXmlResources(t)

	dePath := DiscoverGermanDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	deCount := countRulesFromXML(t, dePath, "de")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, deCount, 0, "de pack must load rules")
	require.Greater(t, globalCount, 0, "global pack must load rules")
	// Official XML has ~885 de <rule> and ~65 global <rule> elements.
	require.GreaterOrEqual(t, deCount, 800, "de pack ~885 rules")
	require.GreaterOrEqual(t, globalCount, 50, "global pack ~65 rules")

	xml := GermanXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=true: total = de + global (deterministic append order).
	require.Equal(t, deCount+globalCount, len(xml.Rules),
		"total rules must be de+global (Java language XML then global)")
	require.NotNil(t, xml.UnifierConfig)

	// Process-cache singleton
	require.Same(t, xml, GermanXmlRuleDisambiguator())
}

func TestNewGermanRuleDisambiguator_WiresXmlRules(t *testing.T) {
	requireDEXmlResources(t)

	xml := GermanXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewGermanRuleDisambiguator()
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(lang, true)")
	require.Same(t, xml, d.Rules, "Rules field is process-cached GermanXmlRuleDisambiguator")
}

// --- de/disambiguation.xml ignore_spelling outcomes (token-built) ----------

func TestGermanXmlRule_ZWEIPFUENDER(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()
	// ZWEIPFÜNDER: regexp [1-9]*-Pfünder → ignore_spelling (marker)
	sent := xml.Disambiguate(tokenSentence("2-Pfünder"))
	requireIgnored(t, sent, "2-Pfünder")
}

func TestGermanXmlRule_KREUCHT_UND_FLEUCHT(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()
	// marker only around kreucht
	sent := xml.Disambiguate(tokenSentence("kreucht", "und", "fleucht"))
	requireIgnored(t, sent, "kreucht")
	requireNotIgnored(t, sent, "und", "fleucht")
}

func TestGermanXmlRule_ENGLISCHE_WOERTER(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()
	sent := xml.Disambiguate(tokenSentence("Bingewatching"))
	requireIgnored(t, sent, "Bingewatching")
}

func TestGermanXmlRule_FUEGUNGEN(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()

	// à + discrétion → discrétion ignored (marker on second)
	sent := xml.Disambiguate(tokenSentence("à", "discrétion"))
	requireNotIgnored(t, sent, "à")
	requireIgnored(t, sent, "discrétion")

	// in + nuce → nuce ignored
	sent = xml.Disambiguate(tokenSentence("in", "nuce"))
	requireNotIgnored(t, sent, "in")
	requireIgnored(t, sent, "nuce")

	// last + minute → minute ignored
	sent = xml.Disambiguate(tokenSentence("last", "minute"))
	requireNotIgnored(t, sent, "last")
	requireIgnored(t, sent, "minute")
}

func TestGermanXmlRule_ABK_MIT_PUNKT(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()
	// belg + . → belg ignored (marker)
	sent := xml.Disambiguate(tokenSentence("belg", "."))
	requireIgnored(t, sent, "belg")
	requireNotIgnored(t, sent, ".")
}

func TestGermanXmlRule_DENegative(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()
	// Random surface not matching ignore_spelling rules on text alone.
	const junk = "xyzzyqwertynotadeword"
	sent := xml.Disambiguate(tokenSentence(junk))
	requireNotIgnored(t, sent, junk)
}

// --- disambiguation-global.xml (proves useGlobal=true) ---------------------

func TestGermanXmlRule_GlobalProperNouns(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()
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

func TestGermanXmlRule_GlobalScientificNames_HomoSpp(t *testing.T) {
	requireDEXmlResources(t)
	xml := GermanXmlRuleDisambiguator()
	// GLOBAL_SCIENTIFIC_NAMES: Homo + spp (no marker → whole match ignored).
	// Note: de pack has HOMO_FABER for Homo+erectus|faber|…, not spp.
	sent := xml.Disambiguate(tokenSentence("Homo", "spp"))
	requireIgnored(t, sent, "Homo", "spp")
}

func TestGermanXmlRule_GlobalOnlyPatternNotInDEAlone(t *testing.T) {
	// Load de-only pack and confirm global-only surface is not ignored; full
	// GermanXml (useGlobal=true) ignores it — proves global was appended.
	requireDEXmlResources(t)

	dePath := DiscoverGermanDisambiguationXML()
	f, err := os.Open(dePath)
	require.NoError(t, err)
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "de", dePath)
	f.Close()
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	deOnly := disambigrules.NewXmlRuleDisambiguator(rules)
	deOnly.UnifierConfig = uni

	// Literal QB|LT and Rem Koolhaas exist only in disambiguation-global.xml.
	deSent := deOnly.Disambiguate(tokenSentence("QB|LT"))
	requireNotIgnored(t, deSent, "QB|LT")
	deSent = deOnly.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, deSent, "Koolhaas")

	full := GermanXmlRuleDisambiguator()
	fullSent := full.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, fullSent, "QB|LT")
	fullSent = full.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireIgnored(t, fullSent, "Koolhaas")
}

func TestGermanXmlRule_FullPipelineViaNewGermanRuleDisambiguator(t *testing.T) {
	requireDEXmlResources(t)
	d := NewGermanRuleDisambiguator()
	require.NotNil(t, d.Rules)

	// de rule
	sent := d.Disambiguate(tokenSentence("2-Pfünder"))
	requireIgnored(t, sent, "2-Pfünder")

	// global-only rule (literal surface from GLOBAL_PROPER_NOUNS)
	sent = d.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")
}
