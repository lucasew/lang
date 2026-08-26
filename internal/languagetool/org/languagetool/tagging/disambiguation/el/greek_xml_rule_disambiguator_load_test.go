package el

// Outcome twins for Greek XmlRuleDisambiguator as used by Greek.createDefaultDisambiguator:
// Java new XmlRuleDisambiguator(this) — useGlobalDisambiguation=false
// (language XML only; does NOT append disambiguation-global.xml).
// Official el/disambiguation.xml has a single rule HAVE_INF:
//   pattern: inflected έχω + marker <and> postag V + postag INF
//   <disambig postag="INF"/> → default REPLACE (Java DisambiguationRuleHandler)
// greek.dict is a tiny stub; seeded readings are used (same bar as Irish text-only twins).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireELXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverGreekDisambiguationXML() == "" {
		t.Skip("official el/disambiguation.xml not discoverable")
	}
}

func strp(s string) *string { return &s }

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
			tr.SetWhitespaceBefore(true)
		}
		toks = append(toks, tr)
	}
	return languagetool.NewAnalyzedSentence(toks)
}

// seededHaveInf builds έχω (lemma έχω for inflected match) + marker with dual V+INF readings.
// Optional surface for έχω (default "έχω"); marker surface defaults to "πάει".
func seededHaveInf(echoSurface, markerSurface string, markerReadings ...*languagetool.AnalyzedToken) *languagetool.AnalyzedSentence {
	if echoSurface == "" {
		echoSurface = "έχω"
	}
	if markerSurface == "" {
		markerSurface = "πάει"
	}
	ss := languagetool.SentenceStartTagName
	lemma := "έχω"
	echo := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken(echoSurface, strp("V"), &lemma))
	// Marker: if no readings provided, seed V + INF (HAVE_INF <and>).
	var marker *languagetool.AnalyzedTokenReadings
	if len(markerReadings) == 0 {
		v, inf := "V", "INF"
		marker = languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(markerSurface, &v, &markerSurface))
		marker.AddReading(languagetool.NewAnalyzedToken(markerSurface, &inf, &markerSurface), "seed")
	} else {
		marker = languagetool.NewAnalyzedTokenReadings(markerReadings[0])
		for _, r := range markerReadings[1:] {
			marker.AddReading(r, "seed")
		}
	}
	marker.SetWhitespaceBefore(true)
	// Positions for Replace / endPos
	pos := 0
	echo.SetStartPos(pos)
	pos += len(echoSurface)
	// space
	sp := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil))
	sp.SetStartPos(pos)
	pos++
	marker.SetStartPos(pos)
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &ss, nil)),
		echo,
		sp,
		marker,
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

func countRulesFromXML(t *testing.T, path, langCode string) int {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	rules, _, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, langCode, path)
	require.NoError(t, err)
	return len(rules)
}

func TestDiscoverGreekDisambiguationXML(t *testing.T) {
	p := DiscoverGreekDisambiguationXML()
	if p == "" {
		t.Skip("official el/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "el")
	require.NotContains(t, p, "disambiguation-global.xml")
	st, err := os.Stat(p)
	require.NoError(t, err)
	require.True(t, st.Mode().IsRegular())
}

func TestGreekXmlRuleDisambiguator_LoadsOfficialPack(t *testing.T) {
	requireELXmlResources(t)

	elPath := DiscoverGreekDisambiguationXML()
	elCount := countRulesFromXML(t, elPath, "el")
	require.Greater(t, elCount, 0, "el pack must load rules")
	// Official el/disambiguation.xml has exactly 1 rule (HAVE_INF).
	require.GreaterOrEqual(t, elCount, 1, "el pack has HAVE_INF")
	require.Equal(t, 1, elCount, "el pack is a single-rule official file")

	xml := GreekXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=false: total == el-only (NOT el+global).
	require.Equal(t, elCount, len(xml.Rules),
		"total rules must equal el-only pack (Java useGlobalDisambiguation=false)")

	// Process-cache singleton
	require.Same(t, xml, GreekXmlRuleDisambiguator())

	ids := make(map[string]bool, len(xml.Rules))
	for _, r := range xml.Rules {
		require.NotNil(t, r)
		ids[r.GetID()] = true
		// No GLOBAL_* ids — proves useGlobal=false isolation.
		require.False(t, strings.HasPrefix(r.GetID(), "GLOBAL_"),
			"must not load disambiguation-global.xml rule %q", r.GetID())
	}
	require.True(t, ids["HAVE_INF"], "missing official rule id HAVE_INF")

	// Spot-check loaded rule semantics: default REPLACE + postag INF + and-group.
	var have *disambigrules.DisambiguationPatternRule
	for _, r := range xml.Rules {
		if r.GetID() == "HAVE_INF" {
			have = r
			break
		}
	}
	require.NotNil(t, have)
	require.Equal(t, disambigrules.ActionReplace, have.Action,
		"Java default when action attr absent is REPLACE")
	require.Equal(t, "INF", have.DisambiguatedPOS)
	require.Len(t, have.Tokens, 2)
	require.True(t, have.Tokens[0].MatchInflected, "έχω is inflected=yes")
	require.Equal(t, "έχω", have.Tokens[0].Token)
	require.True(t, have.Tokens[1].InsideMarker, "second token is in <marker>")
	require.True(t, have.Tokens[1].HasAndGroup(), "marker uses <and> of V + INF")
	require.NotNil(t, have.Tokens[1].Pos)
	require.Equal(t, "V", have.Tokens[1].Pos.PosTag)
	require.Len(t, have.Tokens[1].AndGroup, 1)
	require.NotNil(t, have.Tokens[1].AndGroup[0].Pos)
	require.Equal(t, "INF", have.Tokens[1].AndGroup[0].Pos.PosTag)
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

func TestGreekXmlRuleDisambiguator_UseGlobalFalse_NoGlobalOnlyRules(t *testing.T) {
	// Prove useGlobal=false by isolation: global-only surfaces (QB|LT from
	// disambiguation-global.xml GLOBAL_PROPER_NOUNS) are NOT ignore_spelling
	// under GreekXmlRuleDisambiguator.
	requireELXmlResources(t)
	xml := GreekXmlRuleDisambiguator()
	require.NotNil(t, xml)

	sent := xml.Disambiguate(tokenSentence("QB|LT"))
	tr := tokenBySurface(sent, "QB|LT")
	require.NotNil(t, tr)
	require.False(t, tr.IsIgnoredBySpeller(), "useGlobal=false must not apply GLOBAL_PROPER_NOUNS")

	globalPath := discoverGlobalDisambiguationXMLForIsolation()
	if globalPath == "" {
		return
	}
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, globalCount, 0)
	elCount := countRulesFromXML(t, DiscoverGreekDisambiguationXML(), "el")
	require.NotEqual(t, elCount+globalCount, len(xml.Rules),
		"GreekXml must NOT equal el+global (useGlobal=false)")
	require.Equal(t, elCount, len(xml.Rules))
}

// --- HAVE_INF outcome twins (seeded readings; no full Greek tagger required) ---

// HAVE_INF: έχω (inflected) + token with V and INF → REPLACE marker to INF only.
func TestGreekXmlRule_HAVE_INF_SelectsINF(t *testing.T) {
	requireELXmlResources(t)
	xml := GreekXmlRuleDisambiguator()
	require.NotNil(t, xml)

	// Surface έχω + marker πάει with dual V|INF (matches stub dict shape).
	sent := xml.Disambiguate(seededHaveInf("έχω", "πάει"))
	marker := tokenBySurface(sent, "πάει")
	require.NotNil(t, marker)
	tags := posTagsOn(marker)
	require.Contains(t, tags, "INF", "HAVE_INF must leave INF")
	require.NotContains(t, tags, "V", "HAVE_INF REPLACE postag=INF must drop V")
	require.Len(t, tags, 1, "REPLACE leaves a single INF reading")

	// Inflected surface of έχω (lemma still έχω) also matches.
	sent = xml.Disambiguate(seededHaveInf("έχει", "πάει"))
	marker = tokenBySurface(sent, "πάει")
	require.NotNil(t, marker)
	tags = posTagsOn(marker)
	require.Contains(t, tags, "INF")
	require.NotContains(t, tags, "V")
}

// Negative: dual V+INF without preceding έχω must not force INF-only.
func TestGreekXmlRule_HAVE_INF_NegativeWithoutEcho(t *testing.T) {
	requireELXmlResources(t)
	xml := GreekXmlRuleDisambiguator()
	require.NotNil(t, xml)

	v, inf := "V", "INF"
	surf := "πάει"
	ss := languagetool.SentenceStartTagName
	marker := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken(surf, &v, &surf))
	marker.AddReading(languagetool.NewAnalyzedToken(surf, &inf, &surf), "seed")
	marker.SetStartPos(0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &ss, nil)),
		marker,
	})
	out := xml.Disambiguate(sent)
	tr := tokenBySurface(out, surf)
	require.NotNil(t, tr)
	tags := posTagsOn(tr)
	require.Contains(t, tags, "V", "without έχω, V must remain")
	require.Contains(t, tags, "INF", "without έχω, INF must remain")
	require.Len(t, tags, 2, "rule must not fire without έχω")
}

// Negative: έχω + token with only V (no INF) — <and> fails; no invent INF.
func TestGreekXmlRule_HAVE_INF_NegativeOnlyV(t *testing.T) {
	requireELXmlResources(t)
	xml := GreekXmlRuleDisambiguator()
	require.NotNil(t, xml)

	v := "V"
	surf := "πάω"
	sent := xml.Disambiguate(seededHaveInf("έχω", surf,
		languagetool.NewAnalyzedToken(surf, &v, &surf)))
	tr := tokenBySurface(sent, surf)
	require.NotNil(t, tr)
	tags := posTagsOn(tr)
	require.Contains(t, tags, "V")
	require.NotContains(t, tags, "INF", "must not invent INF when INF reading absent")
}

// Negative: έχω + token with only INF (no V) — <and> fails; INF stays as-is.
func TestGreekXmlRule_HAVE_INF_NegativeOnlyINF(t *testing.T) {
	requireELXmlResources(t)
	xml := GreekXmlRuleDisambiguator()
	require.NotNil(t, xml)

	inf := "INF"
	surf := "πάει"
	sent := xml.Disambiguate(seededHaveInf("έχω", surf,
		languagetool.NewAnalyzedToken(surf, &inf, &surf)))
	tr := tokenBySurface(sent, surf)
	require.NotNil(t, tr)
	tags := posTagsOn(tr)
	require.Equal(t, []string{"INF"}, tags, "only-INF input must stay INF (and-group needs V too)")
}

// Negative: junk surfaces — no invented tags / no global ignore_spelling.
func TestGreekXmlRule_NegativeJunk(t *testing.T) {
	requireELXmlResources(t)
	xml := GreekXmlRuleDisambiguator()
	require.NotNil(t, xml)

	const junk = "xyzzyqwertynotelword"
	sent := xml.Disambiguate(tokenSentence(junk))
	tr := tokenBySurface(sent, junk)
	require.NotNil(t, tr)
	require.False(t, tr.IsImmunized(), "junk must not be immunized")
	require.False(t, tr.IsIgnoredBySpeller(), "junk must not be ignore_spelling")
	for _, p := range posTagsOn(tr) {
		require.NotEqual(t, "INF", p)
		require.NotEqual(t, "V", p)
	}
}
