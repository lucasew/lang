package uk

// Outcome twins for Ukrainian XmlRuleDisambiguator as used by UkrainianHybridDisambiguator:
// Java new XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT) — useGlobalDisambiguation=false
// (language XML only; does NOT append disambiguation-global.xml).
// Official uk.dict is not required for text-only official rules (add / replace / immunize /
// filter / remove on surface match with optional seeded readings).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func requireUKXmlResources(t *testing.T) {
	t.Helper()
	if DiscoverUkrainianDisambiguationXML() == "" {
		t.Skip("official uk/disambiguation.xml not discoverable")
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
			tr.SetWhitespaceBefore(true)
		}
		toks = append(toks, tr)
	}
	return languagetool.NewAnalyzedSentence(toks)
}

// tokenSentenceNoSpace builds SENT_START + adjacent tokens with spacebefore=no
// (for patterns like а + ) with spacebefore="no").
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

// tokenSentenceWithReadings builds SENT_START + content tokens; readings[i] is a list of
// (pos, lemma) pairs for words[i]. Empty readings[i] → untagged surface-only token.
func tokenSentenceWithReadings(words []string, readings [][][2]string) *languagetool.AnalyzedSentence {
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	for i, w := range words {
		if i > 0 {
			toks = append(toks, languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(" ", nil, nil)))
		}
		var anToks []*languagetool.AnalyzedToken
		if i < len(readings) && len(readings[i]) > 0 {
			for _, pl := range readings[i] {
				pos, lem := pl[0], pl[1]
				var p, l *string
				if pos != "" {
					p = &pos
				}
				if lem != "" {
					l = &lem
				}
				anToks = append(anToks, languagetool.NewAnalyzedToken(w, p, l))
			}
		} else {
			anToks = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(w, nil, nil)}
		}
		tr := languagetool.NewAnalyzedTokenReadingsList(anToks, 0)
		if i > 0 {
			tr.SetWhitespaceBefore(true)
		}
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

func TestDiscoverUkrainianDisambiguationXML(t *testing.T) {
	p := DiscoverUkrainianDisambiguationXML()
	if p == "" {
		t.Skip("official uk/disambiguation.xml not discoverable")
	}
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, "uk")
	require.NotContains(t, p, "disambiguation-global.xml")
	st, err := os.Stat(p)
	require.NoError(t, err)
	require.True(t, st.Mode().IsRegular())
}

func TestUkrainianXmlRuleDisambiguator_LoadsOfficialPack(t *testing.T) {
	requireUKXmlResources(t)

	ukPath := DiscoverUkrainianDisambiguationXML()
	ukCount := countRulesFromXML(t, ukPath, "uk")
	require.Greater(t, ukCount, 0, "uk pack must load rules")
	// Official uk/disambiguation.xml has ~475 <rule> elements; loader expands rulegroups.
	require.GreaterOrEqual(t, ukCount, 400, "uk pack ~400+ expanded rules")

	xml := UkrainianXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.NotEmpty(t, xml.Rules)
	// useGlobal=false: total == uk-only (NOT uk+global).
	require.Equal(t, ukCount, len(xml.Rules),
		"total rules must equal uk-only pack (Java useGlobalDisambiguation=false)")
	require.NotNil(t, xml.UnifierConfig, "official uk XML defines <unification> tables")

	// Process-cache singleton
	require.Same(t, xml, UkrainianXmlRuleDisambiguator())

	// Spot-check official rule IDs (from uk/disambiguation.xml).
	ids := make(map[string]bool, len(xml.Rules))
	for _, r := range xml.Rules {
		require.NotNil(t, r)
		ids[r.GetID()] = true
		// No GLOBAL_* ids — proves useGlobal=false isolation.
		require.False(t, strings.HasPrefix(r.GetID(), "GLOBAL_"),
			"must not load disambiguation-global.xml rule %q", r.GetID())
	}
	for _, id := range []string{
		"DIS_KOT_D_IVUAR1", "DIS_KOT_D_IVUAR2", "freq_infix", "POINT_NUMBER",
		"letters_in_lists_noninfl_2", "c-r_1", "c-r_2", "sviatogo_yura",
		"myrnoho_atomu", "date_month", "td_bad", "grata_noninfl_1", "grata_noninfl_2",
		"ska_mistake", "conj_or_part_da", "сyrillic_digit_I_2",
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

func TestUkrainianXmlRuleDisambiguator_UseGlobalFalse_NoGlobalOnlyRules(t *testing.T) {
	// Prove useGlobal=false by isolation: global-only surfaces (QB|LT from
	// disambiguation-global.xml GLOBAL_PROPER_NOUNS) are NOT ignore_spelling
	// under UkrainianXmlRuleDisambiguator.
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()
	require.NotNil(t, xml)

	// Global-only ignore_spelling surface; must stay clean without global pack.
	sent := xml.Disambiguate(tokenSentence("QB|LT"))
	tr := tokenBySurface(sent, "QB|LT")
	require.NotNil(t, tr)
	require.False(t, tr.IsIgnoredBySpeller(), "useGlobal=false must not apply GLOBAL_PROPER_NOUNS")

	// Stronger isolation when global XML is discoverable: Ukrainian count != uk+global.
	globalPath := discoverGlobalDisambiguationXMLForIsolation()
	if globalPath == "" {
		return
	}
	globalCount := countRulesFromXML(t, globalPath, "global")
	require.Greater(t, globalCount, 0)
	ukCount := countRulesFromXML(t, DiscoverUkrainianDisambiguationXML(), "uk")
	require.NotEqual(t, ukCount+globalCount, len(xml.Rules),
		"UkrainianXml must NOT equal uk+global (useGlobal=false)")
	require.Equal(t, ukCount, len(xml.Rules))
}

func TestNewUkrainianHybridDisambiguator_WiresXmlRules(t *testing.T) {
	requireUKXmlResources(t)

	xml := UkrainianXmlRuleDisambiguator()
	require.NotNil(t, xml)

	d := NewUkrainianHybridDisambiguator()
	require.NotNil(t, d.Inner, "Java eagerly constructs XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT)")
	require.Same(t, xml, d.Inner, "Inner field is process-cached UkrainianXmlRuleDisambiguator")
}

// --- Official text-only / seeded outcome twins --------------------------------

// DIS_KOT_D_IVUAR1 (add) then DIS_KOT_D_IVUAR2 (replace postag) on "Кот" + "д'Івуара".
// Final Java-visible outcome: single noninfl:foreign:prop:geo:bad.
func TestUkrainianXmlRule_DIS_KOT_D_IVUAR(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("Кот", "д'Івуара"))
	tr := tokenBySurface(sent, "Кот")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl:foreign:prop:geo:bad")
	// Alone "Кот" must not invent the geo/foreign POS (untouched example).
	sent = xml.Disambiguate(tokenSentence("Кот"))
	tr = tokenBySurface(sent, "Кот")
	require.NotNil(t, tr)
	require.NotContains(t, posTagsOn(tr), "noninfl:foreign:prop:geo:bad")
	require.NotContains(t, posTagsOn(tr), "noninfl:foreign")
}

// freq_infix: surface matching [-–][а-яіїєґ]{1,5}[-–] → immunize.
func TestUkrainianXmlRule_freq_infix_Immunize(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("-ськ-"))
	tr := tokenBySurface(sent, "-ськ-")
	require.NotNil(t, tr)
	require.True(t, tr.IsImmunized(), "freq_infix must immunize -ськ-")

	// Not an infix form → not immunized.
	sent = xml.Disambiguate(tokenSentence("ськ"))
	tr = tokenBySurface(sent, "ськ")
	require.NotNil(t, tr)
	require.False(t, tr.IsImmunized())
}

// POINT_NUMBER: номер + [0-9]+-[а-жєґ] → replace noninfl on the number token.
// Pattern uses inflected=yes with regexp; untagged surface falls back to token text.
func TestUkrainianXmlRule_POINT_NUMBER_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("номер", "17-а"))
	tr := tokenBySurface(sent, "17-а")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl")

	// Alone "17-а" must not fire POINT_NUMBER.
	sent = xml.Disambiguate(tokenSentence("17-а"))
	tr = tokenBySurface(sent, "17-а")
	require.NotNil(t, tr)
	require.NotContains(t, posTagsOn(tr), "noninfl")
}

// house_numbers_noninfl_1: CapitalizedStreet, , N-а .
func TestUkrainianXmlRule_house_numbers_noninfl_1_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("Чорновола", ",", "45-а", "."))
	tr := tokenBySurface(sent, "45-а")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl")
}

// letters_in_lists_noninfl_2: letter + ) (spacebefore=no) → noninfl.
func TestUkrainianXmlRule_letters_in_lists_noninfl(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentenceNoSpace("а", ")"))
	tr := tokenBySurface(sent, "а")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl")

	// Spaced "а )" must not fire spacebefore=no pattern.
	sent = xml.Disambiguate(tokenSentence("а", ")"))
	tr = tokenBySurface(sent, "а")
	require.NotNil(t, tr)
	require.NotContains(t, posTagsOn(tr), "noninfl")
}

// c-r_1 / c-r_2: "ц." "р." → replace POS on each marker.
func TestUkrainianXmlRule_c_r_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("ц.", "р."))
	c := tokenBySurface(sent, "ц.")
	r := tokenBySurface(sent, "р.")
	require.NotNil(t, c)
	require.NotNil(t, r)
	require.Contains(t, posTagsOn(c), "adj:m:v_rod:pron:dem")
	require.Contains(t, posTagsOn(r), "noun:inanim:m:v_rod")
}

// sviatogo_yura: святого + Юра → noun:anim:m:v_rod:prop:fname
func TestUkrainianXmlRule_sviatogo_yura_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("святого", "Юра"))
	tr := tokenBySurface(sent, "Юра")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noun:anim:m:v_rod:prop:fname")

	sent = xml.Disambiguate(tokenSentence("Юра"))
	tr = tokenBySurface(sent, "Юра")
	require.NotNil(t, tr)
	require.NotContains(t, posTagsOn(tr), "noun:anim:m:v_rod:prop:fname")
}

// myrnoho_atomu: мирного + атому → noun:inanim:m:v_rod
func TestUkrainianXmlRule_myrnoho_atomu_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("Мирного", "атому"))
	// Pattern is case-insensitive by default for surface "мирного|..."
	tr := tokenBySurface(sent, "атому")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noun:inanim:m:v_rod")
}

// date_month: day + month abbr + year → noninfl:abbr on month.
func TestUkrainianXmlRule_date_month_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("15", "Тра", "2019"))
	tr := tokenBySurface(sent, "Тра")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl:abbr")
}

// td_bad: і|й + тд → noninfl:bad
func TestUkrainianXmlRule_td_bad_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("і", "тд"))
	tr := tokenBySurface(sent, "тд")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl:bad")

	sent = xml.Disambiguate(tokenSentence("4", "тд"))
	tr = tokenBySurface(sent, "тд")
	require.NotNil(t, tr)
	require.NotContains(t, posTagsOn(tr), "noninfl:bad")
}

// grata_noninfl_2: нон + ґрата → replace noninfl on ґрата.
func TestUkrainianXmlRule_grata_noninfl_2_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("нон", "ґрата"))
	tr := tokenBySurface(sent, "ґрата")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl")
}

// grata_noninfl_1: filter on "нон" before грата — keep noninfl, drop other POS.
func TestUkrainianXmlRule_grata_noninfl_1_Filter(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := tokenSentenceWithReadings(
		[]string{"нон", "грата"},
		[][][2]string{
			{{"noninfl", "нон"}, {"noun:inanim:p:v_rod", "нона"}},
			{{"noun:inanim:f:v_naz", "грата"}},
		},
	)
	out := xml.Disambiguate(sent)
	tr := tokenBySurface(out, "нон")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl")
	require.NotContains(t, posTagsOn(tr), "noun:inanim:p:v_rod")
	require.Len(t, posTagsOn(tr), 1)
}

// ska_mistake: ска + зав… → remove postag=".*" (strip readings).
func TestUkrainianXmlRule_ska_mistake_Remove(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := tokenSentenceWithReadings(
		[]string{"ска", "зав"},
		[][][2]string{
			{
				{"intj:vulg", "ска"},
				{"noun:inanim:m:v_naz:nv", "ска"},
			},
			nil,
		},
	)
	out := xml.Disambiguate(sent)
	tr := tokenBySurface(out, "ска")
	require.NotNil(t, tr)
	// All POS removed (postag=".*"); leftover may be empty/untagged.
	require.Empty(t, posTagsOn(tr), "ska_mistake must remove all POS readings")
}

// o_noninfl_prep_intj: / + о → replace noninfl.
func TestUkrainianXmlRule_o_slash_noninfl_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	// Adjacent or spaced: pattern has no spacebefore constraint on "о" after "/".
	sent := xml.Disambiguate(tokenSentence("/", "о"))
	tr := tokenBySurface(sent, "о")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noninfl")
}

// conj_or_part_da: да + Гамма → part:pers
func TestUkrainianXmlRule_conj_or_part_da_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	sent := xml.Disambiguate(tokenSentence("да", "Гамма"))
	tr := tokenBySurface(sent, "да")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "part:pers")
}

// сyrillic_digit_I_2: І + ст. → number:latin:bad
func TestUkrainianXmlRule_cyrillic_digit_I_2_Replace(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	// Cyrillic capital І (U+0406), not Latin I.
	sent := xml.Disambiguate(tokenSentence("І", "ст."))
	tr := tokenBySurface(sent, "І")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "number:latin:bad")

	// Without "ст." must not fire this rule alone (may fire other І rules with POS context).
	sent = xml.Disambiguate(tokenSentence("І", "все"))
	tr = tokenBySurface(sent, "І")
	require.NotNil(t, tr)
	// Untagged "І" without fname/prep context should not get number:latin:bad from I_2.
	require.NotContains(t, posTagsOn(tr), "number:latin:bad")
}

// Negative: random junk not immunized / no invented POS from official rules above.
func TestUkrainianXmlRule_NegativeJunk(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()

	const junk = "xyzzyqwertynotukword"
	sent := xml.Disambiguate(tokenSentence(junk))
	tr := tokenBySurface(sent, junk)
	require.NotNil(t, tr)
	require.False(t, tr.IsImmunized(), "junk must not be immunized")
	require.False(t, tr.IsIgnoredBySpeller(), "junk must not be ignore_spelling")
	for _, p := range posTagsOn(tr) {
		require.NotEqual(t, "noninfl", p)
		require.NotEqual(t, "noninfl:bad", p)
		require.NotEqual(t, "number:latin:bad", p)
		require.NotEqual(t, "part:pers", p)
	}
}

// Hybrid pipeline (Inner/XML stage): text-only immunize still fires after preDisambiguate+chunker.
func TestUkrainianXmlRule_ViaNewUkrainianHybridDisambiguator(t *testing.T) {
	requireUKXmlResources(t)
	d := NewUkrainianHybridDisambiguator()
	require.NotNil(t, d.Inner)

	sent := d.Disambiguate(tokenSentence("-ськ-"))
	tr := tokenBySurface(sent, "-ськ-")
	require.NotNil(t, tr)
	require.True(t, tr.IsImmunized(), "hybrid must run official XML immunize")

	sent = d.Disambiguate(tokenSentence("святого", "Юра"))
	tr = tokenBySurface(sent, "Юра")
	require.NotNil(t, tr)
	require.Contains(t, posTagsOn(tr), "noun:anim:m:v_rod:prop:fname")
}

// Ensure lemmasOn is exercised for at least one add/replace with explicit lemma.
// DIS_KOT_D_IVUAR1 adds lemma "Кот"; final replace keeps surface-derived lemma behavior.
func TestUkrainianXmlRule_DIS_KOT_LemmaSurface(t *testing.T) {
	requireUKXmlResources(t)
	xml := UkrainianXmlRuleDisambiguator()
	sent := xml.Disambiguate(tokenSentence("Кот", "д'Івуара"))
	tr := tokenBySurface(sent, "Кот")
	require.NotNil(t, tr)
	// Replace-with-postag-only typically keeps token surface as lemma source; assert non-empty readings.
	require.NotEmpty(t, tr.GetReadings())
	_ = lemmasOn(tr)
}
