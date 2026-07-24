package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/suggestions"
	symimpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/symspell/implementation"
	"github.com/stretchr/testify/require"
)

// symSpellLang is a minimal Language for SymSpellRule path/cache tests.
type symSpellLang struct {
	short string
	full  string
	name  string
}

func (l symSpellLang) GetName() string {
	if l.name != "" {
		return l.name
	}
	return l.full
}
func (l symSpellLang) GetShortCode() string { return l.short }
func (l symSpellLang) GetShortCodeWithCountryAndVariant() string {
	if l.full != "" {
		return l.full
	}
	return l.short
}
func (l symSpellLang) GetCountries() []string                     { return nil }
func (l symSpellLang) CreateDefaultTagger() languagetool.Tagger { return nil }

func enUSLang() languagetool.Language {
	return symSpellLang{short: "en", full: "en-US", name: "English (US)"}
}

// withSymSpellFixtures installs an in-memory broker + clears SymSpell caches.
func withSymSpellFixtures(t *testing.T, resources map[string]string) {
	t.Helper()
	prev := GetDataBroker()
	clearSymSpellRuleCaches()
	m := broker.NewMapResourceDataBroker()
	for k, v := range resources {
		m.Resource[k] = v
	}
	SetDataBroker(m)
	t.Cleanup(func() {
		clearSymSpellRuleCaches()
		SetDataBroker(prev)
		suggestions.ResetSuggestionsChanges()
	})
}

func defaultEnUSResources() map[string]string {
	return map[string]string{
		"en/hunspell/spelling.txt":        "hello\nworld\nhallo\n",
		"en/hunspell/spelling_en-US.txt":  "color\n",
		"en/hunspell/en_US.dic":           "hello+Z\nworld+M\n",
		"en/hunspell/ignore.txt":          "xyzzy\n",
		// Java loads "probibit.txt" (typo), not prohibit.txt.
		"en/hunspell/probibit.txt": "hallo\nbadword\n",
	}
}

// Twin: getId / getDescription.
func TestSymSpellRule_IDAndDescription(t *testing.T) {
	withSymSpellFixtures(t, nil)
	r := NewSymSpellRule(enUSLang(), languagetool.NewUserConfig())
	require.Equal(t, "SYMSPELL_RULE", r.GetID())
	require.Equal(t, "Spell checking rule using SymSpell algorithm", r.GetDescription())
}

// Twin: getSpellingDictBaseDir.
func TestSymSpellRule_GetSpellingDictBaseDir(t *testing.T) {
	require.Equal(t, "en/hunspell/", GetSpellingDictBaseDir(enUSLang()))
	require.Equal(t, "de/hunspell/", GetSpellingDictBaseDir(symSpellLang{short: "de", full: "de-DE"}))
}

// Twin: initUserDictSpeller — null/empty → null; accepted words → committed speller.
func TestSymSpellRule_InitUserDictSpeller(t *testing.T) {
	require.Nil(t, InitUserDictSpeller(nil))
	require.Nil(t, InitUserDictSpeller(languagetool.NewUserConfig()))
	cfg := languagetool.NewUserConfigWithWords([]string{"myword", "other"}, nil)
	sp := InitUserDictSpeller(cfg)
	require.NotNil(t, sp)
	// exact lookup distance 0
	got := sp.LookupMax("myword", symimpl.VerbosityTop, 0)
	require.Len(t, got, 1)
	require.Equal(t, "myword", got[0].Term)
}

// Twin: initDefaultDictSpeller loads spelling.txt + variant + freq .dic; skips prohibited.
func TestSymSpellRule_InitDefaultDictSpeller(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	sp := InitDefaultDictSpeller(enUSLang())
	require.NotNil(t, sp)
	// hello from dic (freq) and spelling.txt
	require.NotEmpty(t, sp.LookupMax("hello", symimpl.VerbosityTop, 0))
	// color from spelling_en-US.txt
	require.NotEmpty(t, sp.LookupMax("color", symimpl.VerbosityTop, 0))
	// hallo is in spelling.txt but also in probibit.txt → not in dictionary
	require.Empty(t, sp.LookupMax("hallo", symimpl.VerbosityTop, 0))
	// world from dic
	require.NotEmpty(t, sp.LookupMax("world", symimpl.VerbosityTop, 0))
}

// Twin: bad frequency dictionary line → IllegalArgumentException.
func TestSymSpellRule_InitDefaultDictSpeller_BadDicLine(t *testing.T) {
	withSymSpellFixtures(t, map[string]string{
		"en/hunspell/en_US.dic": "noplus\n",
	})
	require.Panics(t, func() { InitDefaultDictSpeller(enUSLang()) })
}

// Twin: isMisspelled always throws RuntimeException("not implemented").
func TestSymSpellRule_IsMisspelledNotImplemented(t *testing.T) {
	withSymSpellFixtures(t, nil)
	r := NewSymSpellRule(enUSLang(), nil)
	require.PanicsWithValue(t, "not implemented", func() { r.IsMisspelled("x") })
}

// Twin: match — empty candidates → "Misspelling or unknown word!".
func TestSymSpellRule_Match_UnknownWord(t *testing.T) {
	withSymSpellFixtures(t, nil) // empty resources → empty default dict
	r := NewSymSpellRule(enUSLang(), nil)
	sent := languagetool.AnalyzePlain("foobar")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, "Misspelling or unknown word!", ms[0].Message)
	require.Equal(t, "foobar", sent.GetText()[ms[0].FromPos:ms[0].ToPos])
}

// Twin: match — known exact word (first candidate == word) → no RuleMatch.
func TestSymSpellRule_Match_ExactKnownWord(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	r := NewSymSpellRule(enUSLang(), nil)
	sent := languagetool.AnalyzePlain("hello")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Twin: match — misspelling with suggestions message "Misspelling!".
func TestSymSpellRule_Match_MisspellingWithSuggestions(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	r := NewSymSpellRule(enUSLang(), nil)
	sent := languagetool.AnalyzePlain("helo")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, "Misspelling!", ms[0].Message)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "hello")
}

// Twin: ignoredWordsCache (ignore.txt raw) skips token entirely.
func TestSymSpellRule_Match_IgnoredWordSkipped(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	r := NewSymSpellRule(enUSLang(), nil)
	sent := languagetool.AnalyzePlain("xyzzy")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Twin: filterCandidates drops ignore + prohibited (probibit.txt cache).
func TestSymSpellRule_FilterCandidates(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	r := NewSymSpellRule(enUSLang(), nil)
	got := r.filterCandidates([]string{"hello", "hallo", "xyzzy", "world"})
	require.Equal(t, []string{"hello", "world"}, got)
}

// Twin: user candidates are not filterCandidates'd (Java match path).
func TestSymSpellRule_UserCandidatesNotFiltered(t *testing.T) {
	// hallo is prohibited for default filter, but user speller can still return it.
	withSymSpellFixtures(t, map[string]string{
		"en/hunspell/spelling.txt": "hello\n",
		"en/hunspell/en_US.dic":    "hello+Z\n",
		"en/hunspell/probibit.txt": "hallo\n",
	})
	cfg := languagetool.NewUserConfigWithWords([]string{"hallo"}, nil)
	r := NewSymSpellRule(enUSLang(), cfg)
	// user lookup of "hallo" exact
	user := r.getSpellerMatches("hallo", r.userDictSpeller)
	require.Equal(t, []string{"hallo"}, user)
	// filter would drop hallo from default candidates
	filtered := r.filterCandidates([]string{"hallo", "hello"})
	require.Equal(t, []string{"hello"}, filtered)
	// Match: first user candidate == word → no match (correct via user dict)
	sent := languagetool.AnalyzePlain("hallo")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms, "userDict exact first candidate equals word")
}

// Twin: getSpellerMatches(null) → empty list.
func TestSymSpellRule_GetSpellerMatches_Nil(t *testing.T) {
	withSymSpellFixtures(t, nil)
	r := NewSymSpellRule(enUSLang(), nil)
	require.Empty(t, r.getSpellerMatches("x", nil))
}

// Twin: initParameters — candidates Verbosity + editDistance from experiment.
func TestSymSpellRule_InitParameters_Experiment(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	sc := suggestions.InitSuggestionsChanges(&suggestions.SuggestionChangesTestConfig{
		ExperimentRuns: []suggestions.SuggestionChangesExperimentRuns{
			{Name: "exp", Parameters: map[string][]any{
				"candidates":   {"All"},
				"editDistance": {2},
			}},
		},
	})
	exps := sc.GetExperiments()
	require.NotEmpty(t, exps)
	sc.SetCurrentExperiment(exps[0])

	r := NewSymSpellRule(enUSLang(), nil)
	require.Equal(t, symimpl.VerbosityAll, r.Verbosity())
	require.Equal(t, 2, r.EditDistance())
}

// Twin: SymSpell+NewSuggestionsOrderer experiment installs SuggestionsOrdererFeatureExtractor.
func TestSymSpellRule_InitParameters_OrdererExperiment(t *testing.T) {
	withSymSpellFixtures(t, nil)
	sc := suggestions.InitSuggestionsChanges(&suggestions.SuggestionChangesTestConfig{
		ExperimentRuns: []suggestions.SuggestionChangesExperimentRuns{
			{Name: "SymSpell+NewSuggestionsOrderer"},
		},
	})
	exps := sc.GetExperiments()
	require.NotEmpty(t, exps)
	sc.SetCurrentExperiment(exps[0])
	require.True(t, suggestions.IsRunningExperiment("SymSpell+NewSuggestionsOrderer"))

	r := NewSymSpellRule(enUSLang(), nil)
	require.NotNil(t, r.Orderer())
	_, ok := r.Orderer().(*suggestions.SuggestionsOrdererFeatureExtractor)
	require.True(t, ok)
}

// Twin: constructor wires userDictSpeller from UserConfig.
func TestSymSpellRule_Constructor_UserDict(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	cfg := languagetool.NewUserConfigWithWords([]string{"customword"}, nil)
	r := NewSymSpellRule(enUSLang(), cfg)
	require.NotNil(t, r.UserDictSpeller())
	require.NotNil(t, r.DefaultDictSpeller())
	// customword known via user
	user := r.getSpellerMatches("customword", r.UserDictSpeller())
	require.Equal(t, []string{"customword"}, user)
}

// Twin: spellerCache reuses default dict for same language key.
func TestSymSpellRule_SpellerCache(t *testing.T) {
	withSymSpellFixtures(t, defaultEnUSResources())
	r1 := NewSymSpellRule(enUSLang(), nil)
	r2 := NewSymSpellRule(enUSLang(), nil)
	require.Same(t, r1.DefaultDictSpeller(), r2.DefaultDictSpeller())
}

// Twin: IgnoreWordsWithLength on base must not skip (Java match uses ignore set only).
func TestSymSpellRule_IgnoreSetOnlyNotLength(t *testing.T) {
	withSymSpellFixtures(t, nil)
	r := NewSymSpellRule(enUSLang(), nil)
	r.IgnoreWordsWithLength = 1
	sent := languagetool.AnalyzePlain("x")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, ms, "IgnoreWordsWithLength must not skip like ignore set")
	require.Equal(t, "Misspelling or unknown word!", ms[0].Message)
}

// Twin: frequency class A→1 … Z→26 via createDictionaryEntry counts (tiebreaker only).
func TestSymSpellRule_FreqDictParse(t *testing.T) {
	withSymSpellFixtures(t, map[string]string{
		"en/hunspell/en_US.dic": "rare+A\ncommon+Z\n",
	})
	sp := InitDefaultDictSpeller(enUSLang())
	// Both exact present
	require.NotEmpty(t, sp.LookupMax("rare", symimpl.VerbosityTop, 0))
	require.NotEmpty(t, sp.LookupMax("common", symimpl.VerbosityTop, 0))
	// Closest on a typo equidistant would prefer higher count — "comon" distance 1 to common
	got := sp.LookupMax("comon", symimpl.VerbosityClosest, 1)
	require.NotEmpty(t, got)
	require.Equal(t, "common", got[0].Term)
}
