package ga

// Outcome twins for Irish XmlRuleDisambiguator as used by IrishHybridDisambiguator:
// Java new XmlRuleDisambiguator(Irish.getInstance()) — useGlobalDisambiguation=false
// (language XML only; does NOT append disambiguation-global.xml).
// Official ga.dict is not required: use token-built sentences for text-only official
// rules (add POS / immunize / replace on surface match).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireGAXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverIrishDisambiguationXML() == "" {
		t.Skip("official ga/disambiguation.xml not discoverable")
	}
}

// tokenSentence builds SENT_START + tokens with spaces between word tokens.
// Content tokens after the first get IsWhitespaceBefore=true (analyzer-faithful).
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
		tr := languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(w, nil, nil))
		if i > 0 {
			// Preceded by whitespace token — same as AnalyzedTokenReadings from Analyze.
			tr.SetWhitespaceBefore(true)
		}
		toks = append(toks, tr)
	}
	return languagetool.NewAnalyzedSentence(toks)
}

// tokenSentenceNoSpace builds SENT_START + adjacent tokens with spacebefore=no
// (for patterns like rte + . + ie with spacebefore="no").
func tokenSentenceNoSpace(words ...string) *languagetool.AnalyzedSentence {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	pos := 0
	for i, w := range words {
		tr := languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(w, nil, nil))
		if i > 0 {
			tr.SetWhitespaceBefore(false)
		}
		tr.SetStartPos(pos)
		pos += len(w)
		toks = append(toks, tr)
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

func lemmasOn(tr *languagetool.AnalyzedTokenReadings) []string {
	if tr == nil {
		return nil
	}
	var lemmas []string
	for _, r := range tr.GetReadings() {
		if r != nil && r.GetLemma() != nil {
			lemmas = append(lemmas, *r.GetLemma())
		}
	}
	return lemmas
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

func TestDiscoverIrishDisambiguationXML(t *testing.T) {
	p := DiscoverIrishDisambiguationXML()
	if p == "" {
		t.Skip("official ga/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "ga")
	require.NotContains(t, p, "disambiguation-global.xml")
	st, err := os.Stat(p)
	require.NoError(t, err)
	require.True(t, st.Mode().IsRegular())
}

func TestIrishXmlRuleDisambiguator_LoadsOfficialPack(t *testing.T) {
	requireGAXmlResources(t)

	gaPath := DiscoverIrishDisambiguationXML()
	gaCount := countRulesFromXML(t, gaPath, "ga")
	require.Greater(t, gaCount, 0, "ga pack must load rules")
	// Official ga/disambiguation.xml has ~138 <rule> elements (~143 with comments).
	// Loader yields expanded rules; expect a large language-only pack.
	require.GreaterOrEqual(t, gaCount, 100, "ga pack ~100+ expanded rules")

	xml := IrishXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=false: total == ga-only (NOT ga+global).
	require.Equal(t, gaCount, len(xml.Rules),
		"total rules must equal ga-only pack (Java useGlobalDisambiguation=false)")
	require.NotNil(t, xml.UnifierConfig, "official ga XML defines <unification> tables")

	// Process-cache singleton
	require.Same(t, xml, IrishXmlRuleDisambiguator())

	// Spot-check official rule IDs (from ga/disambiguation.xml).
	ids := make(map[string]bool, len(xml.Rules))
	for _, r := range xml.Rules {
		require.NotNil(t, r)
		ids[r.GetID()] = true
		// No GLOBAL_* ids — proves useGlobal=false isolation.
		require.False(t, strings.HasPrefix(r.GetID(), "GLOBAL_"),
			"must not load disambiguation-global.xml rule %q", r.GetID())
	}
	for _, id := range []string{
		"NUM_DIG_ORD", "NUM_DIG_ORD_OBS", "RTE_PONC_IE", "DE_SHIOR",
		"AS_SIN_AMACH", "A_HAON", "DE_LUAIN",
	} {
		require.True(t, ids[id], "missing official rule id %s", id)
	}
}

func discoverGlobalDisambiguationXMLForIsolation() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core",
		"src", "main", "resources", "org", "languagetool", "resource", "disambiguation-global.xml")
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

func TestIrishXmlRuleDisambiguator_UseGlobalFalse_NoGlobalOnlyRules(t *testing.T) {
	// Prove useGlobal=false by isolation: global-only surfaces (QB|LT from
	// disambiguation-global.xml GLOBAL_PROPER_NOUNS) are NOT ignore_spelling
	// under IrishXmlRuleDisambiguator.
	requireGAXmlResources(t)
	xml := IrishXmlRuleDisambiguator()
	require.NotNil(t, xml)

	// Global-only ignore_spelling surface; must stay clean without global pack.
	sent := xml.Disambiguate(tokenSentence("QB|LT"))
	tr := tokenBySurface(sent, "QB|LT")
	require.NotNil(t, tr)
	require.False(t, tr.IsIgnoredBySpeller(), "useGlobal=false must not apply GLOBAL_PROPER_NOUNS")

	// Stronger isolation when global XML is discoverable: Irish count != ga+global.
	globalPath := discoverGlobalDisambiguationXMLForIsolation()
	if globalPath == "" {
		return
	}
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, globalCount, 0)
	gaCount := countRulesFromXML(t, DiscoverIrishDisambiguationXML(), "ga")
	require.NotEqual(t, gaCount+globalCount, len(xml.Rules),
		"IrishXml must NOT equal ga+global (useGlobal=false)")
	require.Equal(t, gaCount, len(xml.Rules))
}

func TestNewIrishHybridDisambiguator_WiresXmlRules(t *testing.T) {
	requireGAXmlResources(t)

	xml := IrishXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewIrishHybridDisambiguator()
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(Irish.getInstance())")
	require.Same(t, xml, d.Rules, "Rules field is process-cached IrishXmlRuleDisambiguator")
}

// --- Official text-only outcome twins (no ga.dict) -------------------------

// NUM_DIG_ORD: surface matching [1-9][0-9]*ú (e.g. 6ú) → add POS Num:Dig:Ord
func TestIrishXmlRule_NUM_DIG_ORD(t *testing.T) {
	requireGAXmlResources(t)
	xml := IrishXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("6ú"))
	tr := tokenBySurface(sent, "6ú")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "Num:Dig:Ord")

	// Multi-digit ordinal form also matches.
	sent = xml.Disambiguate(tokenSentence("21ú"))
	tr = tokenBySurface(sent, "21ú")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "Num:Dig:Ord")

	// Leading zero must NOT match ([1-9]...).
	sent = xml.Disambiguate(tokenSentence("0ú"))
	tr = tokenBySurface(sent, "0ú")
	require.NotNil(t, tr)
	require.NotContains(t, posTagsOn(tr), "Num:Dig:Ord")
}

// NUM_DIG_ORD_OBS: [1-9][0-9]*adh (e.g. 6adh) → Num:Dig:Ord
func TestIrishXmlRule_NUM_DIG_ORD_OBS(t *testing.T) {
	requireGAXmlResources(t)
	xml := IrishXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("6adh"))
	tr := tokenBySurface(sent, "6adh")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "Num:Dig:Ord")
}

// A_HAON subgroup: a + [0-9]+ → add Num:Dig on the digit token.
func TestIrishXmlRule_A_PlusDigits_AddNumDig(t *testing.T) {
	requireGAXmlResources(t)
	xml := IrishXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("a", "6"))
	six := tokenBySurface(sent, "6")
	require.NotNil(t, six)
	require.Contains(t, posTagsOn(six), "Num:Dig")

	// Digit alone (no preceding a) must not invent Num:Dig from this rule.
	sent = xml.Disambiguate(tokenSentence("6"))
	six = tokenBySurface(sent, "6")
	require.NotNil(t, six)
	require.NotContains(t, posTagsOn(six), "Num:Dig")
}

// RTE_PONC_IE: rte + . + ie (spacebefore=no) → immunize marker tokens.
func TestIrishXmlRule_RTE_PONC_IE_Immunize(t *testing.T) {
	requireGAXmlResources(t)
	xml := IrishXmlRuleDisambiguator()

	// Two official rules: immunize "rte" and immunize "ie" in rte.ie.
	sent := xml.Disambiguate(tokenSentenceNoSpace("rte", ".", "ie"))
	rte := tokenBySurface(sent, "rte")
	ie := tokenBySurface(sent, "ie")
	require.NotNil(t, rte)
	require.NotNil(t, ie)
	require.True(t, rte.IsImmunized(), "rte must be immunized by RTE_PONC_IE")
	require.True(t, ie.IsImmunized(), "ie must be immunized by RTE_PONC_IE")

	// With intervening spaces (spacebefore default true) pattern must not fire.
	sentSpaced := xml.Disambiguate(tokenSentence("rte", ".", "ie"))
	require.False(t, tokenBySurface(sentSpaced, "rte").IsImmunized(),
		"spaced rte . ie must not immunize")
	require.False(t, tokenBySurface(sentSpaced, "ie").IsImmunized(),
		"spaced rte . ie must not immunize ie")
}

// DE_SHIOR: de + shíor → replace on shíor with lemma síor / pos Subst:Noun:Sg:Len
func TestIrishXmlRule_DE_SHIOR_Replace(t *testing.T) {
	requireGAXmlResources(t)
	xml := IrishXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("de", "shíor"))
	shi := tokenBySurface(sent, "shíor")
	require.NotNil(t, shi)
	require.Contains(t, posTagsOn(shi), "Subst:Noun:Sg:Len")
	require.Contains(t, lemmasOn(shi), "síor")
}

// Negative: random junk not immunized / no invented POS.
func TestIrishXmlRule_NegativeJunk(t *testing.T) {
	requireGAXmlResources(t)
	xml := IrishXmlRuleDisambiguator()

	const junk = "xyzzyqwertynotgaword"
	sent := xml.Disambiguate(tokenSentence(junk))
	tr := tokenBySurface(sent, junk)
	require.NotNil(t, tr)
	require.False(t, tr.IsImmunized(), "junk must not be immunized")
	require.False(t, tr.IsIgnoredBySpeller(), "junk must not be ignore_spelling")
	// No invented POS tags on unmatched surface.
	for _, p := range posTagsOn(tr) {
		require.NotEqual(t, "Num:Dig:Ord", p)
		require.NotEqual(t, "Num:Dig", p)
		require.NotEqual(t, "Subst:Noun:Sg:Len", p)
	}
}

// Hybrid pipeline (Rules stage only when Chunker left as-is): text-only add still fires.
func TestIrishXmlRule_ViaNewIrishHybridDisambiguator(t *testing.T) {
	requireGAXmlResources(t)
	d := NewIrishHybridDisambiguator()
	require.NotNil(t, d.Rules)

	sent := d.Disambiguate(tokenSentence("6ú"))
	tr := tokenBySurface(sent, "6ú")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "Num:Dig:Ord")

	sent = d.Disambiguate(tokenSentenceNoSpace("rte", ".", "ie"))
	require.True(t, tokenBySurface(sent, "rte").IsImmunized())
	require.True(t, tokenBySurface(sent, "ie").IsImmunized())
}
