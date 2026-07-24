package en

// Outcome twins for English XmlRuleDisambiguator as used by EnglishHybridDisambiguator:
// Java new XmlRuleDisambiguator(lang, true) — en/disambiguation.xml then disambiguation-global.xml.
// Official english.dict is not required for most cases: use token-built sentences +
// IsIgnoredBySpeller / added POS for rules that match on token text/regexp.

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireENXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverEnglishDisambiguationXML() == "" {
		t.Skip("official en/disambiguation.xml not discoverable")
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

// tokenSentenceNoSpace builds SENT_START + adjacent non-whitespace tokens (no " "
// between them) for spacebefore="no" patterns such as contractions.
func tokenSentenceNoSpace(words ...string) *languagetool.AnalyzedSentence {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for _, w := range words {
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

func posTagsOn(tr *languagetool.AnalyzedTokenReadings) []string {
	if tr == nil {
		return nil
	}
	var tags []string
	for _, r := range tr.GetReadings() {
		if r != nil && r.GetPOSTag() != nil {
			tags = append(tags, *r.GetPOSTag())
		}
	}
	return tags
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

func TestDiscoverEnglishDisambiguationXML(t *testing.T) {
	p := DiscoverEnglishDisambiguationXML()
	if p == "" {
		t.Skip("official en/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "en")
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

func TestEnglishHybridXmlRuleDisambiguator_LoadsOfficialPacks(t *testing.T) {
	requireENXmlResources(t)

	enPath := DiscoverEnglishDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	enCount := countRulesFromXML(t, enPath, "en")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, enCount, 0, "en pack must load rules")
	require.Greater(t, globalCount, 0, "global pack must load rules")
	// Official XML expands to ~1006 en rules and ~65 global rules.
	require.GreaterOrEqual(t, enCount, 900, "en pack ~1006 rules")
	require.GreaterOrEqual(t, globalCount, 50, "global pack ~65 rules")

	xml := EnglishHybridXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=true: total = en + global (deterministic append order).
	require.Equal(t, enCount+globalCount, len(xml.Rules),
		"total rules must be en+global (Java language XML then global)")
	require.NotNil(t, xml.UnifierConfig)

	// Process-cache singleton
	require.Same(t, xml, EnglishHybridXmlRuleDisambiguator())
}

func TestEnglishXmlRuleDisambiguator_LocalOnlyNoGlobal(t *testing.T) {
	requireENXmlResources(t)

	enPath := DiscoverEnglishDisambiguationXML()
	globalPath := DiscoverGlobalDisambiguationXML()
	enCount := countRulesFromXML(t, enPath, "en")
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, enCount, 0)
	require.Greater(t, globalCount, 0)

	// Java new XmlRuleDisambiguator(English) — useGlobalDisambiguation=false
	local := EnglishXmlRuleDisambiguator()
	require.NotNil(t, local)
	require.Equal(t, enCount, len(local.Rules), "useGlobal=false must be EN pack only")
	require.Same(t, local, EnglishXmlRuleDisambiguator())

	hybrid := EnglishHybridXmlRuleDisambiguator()
	require.NotNil(t, hybrid)
	require.Equal(t, enCount+globalCount, len(hybrid.Rules))
	require.NotSame(t, local, hybrid, "false and true packs are distinct singletons")
}

func TestDefaultEnglishHybrid_WiresHybridXmlRules(t *testing.T) {
	requireENXmlResources(t)

	xml := EnglishHybridXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, d.RulesDisambiguator, "Java constructs XmlRuleDisambiguator(lang, true)")
	require.Same(t, xml, d.RulesDisambiguator, "RulesDisambiguator is process-cached hybrid XML")
}

// --- en/disambiguation.xml outcomes (token-built) --------------------------

func TestEnglishXmlRule_UNKNOWN_PCT(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	// UNKNOWN_PCT: regexp [\.,;:…!\?] → add PCT
	sent := xml.Disambiguate(tokenSentence("."))
	tr := tokenBySurface(sent, ".")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "PCT")
	sent = xml.Disambiguate(tokenSentence(","))
	tr = tokenBySurface(sent, ",")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "PCT")
	// COMMA_POSTAG also adds ","
	require.Contains(t, posTagsOn(tr), ",")
}

func TestEnglishXmlRule_CD(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	// CD: regexp \d+ → add CD
	sent := xml.Disambiguate(tokenSentence("10"))
	tr := tokenBySurface(sent, "10")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "CD")
}

func TestEnglishXmlRule_KUNG_FU(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	// KUNG_FU: kung fu → ignore_spelling
	sent := xml.Disambiguate(tokenSentence("kung", "fu"))
	requireIgnored(t, sent, "kung", "fu")
}

func TestEnglishXmlRule_SPELLING_IN_VIVO(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	// SPELLING_IN_VIVO: in vivo → ignore_spelling
	sent := xml.Disambiguate(tokenSentence("in", "vivo"))
	requireIgnored(t, sent, "in", "vivo")
}

func TestEnglishXmlRule_SPELLING_KETO(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	// SPELLING_KETO: keto + acids?|diets?
	sent := xml.Disambiguate(tokenSentence("keto", "diet"))
	requireIgnored(t, sent, "keto", "diet")
}

func TestEnglishXmlRule_CONTRACTION_CANT_Ignore(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	// CONTRACTIONS: ca + n't (spacebefore=no) → ignore_spelling on "ca"
	sent := xml.Disambiguate(tokenSentenceNoSpace("ca", "n't"))
	requireIgnored(t, sent, "ca")
}

func TestEnglishXmlRule_ENNegative(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	const junk = "xyzzyqwertynotadeword"
	sent := xml.Disambiguate(tokenSentence(junk))
	requireNotIgnored(t, sent, junk)
	tr := tokenBySurface(sent, junk)
	require.NotNil(t, tr)
	require.NotContains(t, posTagsOn(tr), "PCT")
	require.NotContains(t, posTagsOn(tr), "CD")
}

// --- disambiguation-global.xml (proves useGlobal=true) ---------------------

func TestEnglishXmlRule_GlobalProperNouns(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
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

func TestEnglishXmlRule_GlobalScientificNames_HomoSpp(t *testing.T) {
	requireENXmlResources(t)
	xml := EnglishHybridXmlRuleDisambiguator()
	// GLOBAL_SCIENTIFIC_NAMES: Homo + spp (no marker → whole match ignored).
	sent := xml.Disambiguate(tokenSentence("Homo", "spp"))
	requireIgnored(t, sent, "Homo", "spp")
}

func TestEnglishXmlRule_GlobalOnlyPatternNotInENAlone(t *testing.T) {
	// useGlobal=false pack must not apply global-only surfaces; hybrid true must.
	requireENXmlResources(t)

	local := EnglishXmlRuleDisambiguator()
	require.NotNil(t, local)

	// Literal QB|LT and Rem Koolhaas exist only in disambiguation-global.xml.
	nlSent := local.Disambiguate(tokenSentence("QB|LT"))
	requireNotIgnored(t, nlSent, "QB|LT")
	nlSent = local.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireNotIgnored(t, nlSent, "Koolhaas")

	full := EnglishHybridXmlRuleDisambiguator()
	fullSent := full.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, fullSent, "QB|LT")
	fullSent = full.Disambiguate(tokenSentence("Rem", "Koolhaas"))
	requireIgnored(t, fullSent, "Koolhaas")
}

func TestEnglishXmlRule_FullPipelineViaDefaultEnglishHybrid(t *testing.T) {
	requireENXmlResources(t)
	d := DefaultEnglishHybridDisambiguator()
	require.NotNil(t, d.RulesDisambiguator)

	// en rule
	sent := d.Disambiguate(tokenSentence("kung", "fu"))
	requireIgnored(t, sent, "kung", "fu")

	// global-only rule (literal surface from GLOBAL_PROPER_NOUNS)
	sent = d.Disambiguate(tokenSentence("QB|LT"))
	requireIgnored(t, sent, "QB|LT")

	// UNKNOWN_PCT still via hybrid RulesDisambiguator
	sent = d.Disambiguate(tokenSentence("."))
	tr := tokenBySurface(sent, ".")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "PCT")
}
