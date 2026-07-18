package en

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestENRuleFiltersRegistered(t *testing.T) {
	for _, class := range []string{
		"org.languagetool.rules.en.OrdinalSuffixFilter",
		"org.languagetool.rules.en.AdverbFilter",
		"org.languagetool.rules.en.FutureDateFilter",
		"org.languagetool.rules.en.DateCheckFilter",
		"org.languagetool.rules.en.NewYearDateFilter",
		"org.languagetool.rules.en.YMDNewYearDateFilter",
		"org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter",
		"org.languagetool.rules.en.EnglishNumberInWordFilter",
		"org.languagetool.rules.en.FindSuggestionsFilter",
		"org.languagetool.rules.patterns.RegexAntiPatternFilter",
	} {
		require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
		require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(class), class)
	}
}

func TestNewYearDateRuleFilter(t *testing.T) {
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.NewYearDateFilter")
	// rules.IsTest() forces January 2014 → year 2013 non-December should flag
	m := rules.NewRuleMatch(nil, nil, 0, 5, "Did you mean {realYear} instead of {year}?")
	out := f.AcceptRuleMatch(m, map[string]string{"year": "2013", "month": "March", "day": "1"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.Message, "2014")
	require.Contains(t, out.Message, "2013")
	// December suppressed
	out = f.AcceptRuleMatch(m, map[string]string{"year": "2013", "month": "December", "day": "1"}, 0, nil, nil)
	require.Nil(t, out)
}

func TestNumberInWordAndFindSuggestionsFailClosed(t *testing.T) {
	ClearEnglishFilterSpeller()
	ClearEnglishFilterTagger()
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.EnglishNumberInWordFilter")
	m := rules.NewRuleMatch(nil, nil, 0, 4, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "t0ken"}, 0, nil, nil))
	f2 := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.FindSuggestionsFilter")
	// No speller dict → fail-closed (do not invent suggestions).
	require.Nil(t, f2.AcceptRuleMatch(m, map[string]string{"wordFrom": "1", "desiredPostag": "VB"}, 0, nil, nil))
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", ".."))
}

func TestNumberInWordWithOfficialDict(t *testing.T) {
	root := repoRoot(t)
	dict := filepath.Join(root, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", "en_US.dict")
	if st, err := os.Stat(dict); err != nil || st.IsDir() {
		t.Skipf("en_US.dict missing: %s", dict)
	}
	require.True(t, WireEnglishFilterSpeller(dict))
	t.Cleanup(ClearEnglishFilterSpeller)

	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.EnglishNumberInWordFilter")
	m := rules.NewRuleMatch(nil, nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"word": "H0use"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "House")
}

func TestFindSuggestionsWithDictAndTagger(t *testing.T) {
	root := repoRoot(t)
	speller := filepath.Join(root, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", "en_US.dict")
	pos := filepath.Join(root, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict")
	if st, err := os.Stat(speller); err != nil || st.IsDir() {
		t.Skipf("en_US.dict missing: %s", speller)
	}
	if st, err := os.Stat(pos); err != nil || st.IsDir() {
		t.Skipf("english.dict missing: %s", pos)
	}
	require.True(t, WireEnglishFilterSpeller(speller))
	require.True(t, WireEnglishFilterTagger(pos))
	t.Cleanup(func() {
		ClearEnglishFilterSpeller()
		ClearEnglishFilterTagger()
	})

	// POS probe: known verb should match VB.*
	require.True(t, FilterSuggestionMatchesPostag("running", "VBG|NN"))
	require.False(t, FilterSuggestionMatchesPostag("running", "^DT$"))

	// Filter end-to-end: suggestion must pass desiredPostag
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.FindSuggestionsFilter")
	tok := atr("runing", 0) // misspelling of running
	m := rules.NewRuleMatch(nil, languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{tok}), 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*|NN",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, []int{1})
	// may or may not find "running" depending on SuggestEdits; if any sugs, all must tag-match
	if out != nil {
		for _, s := range out.GetSuggestedReplacements() {
			require.True(t, FilterSuggestionMatchesPostag(s, "VB.*|NN"), "suggestion %q must match desiredPostag", s)
		}
	}
}

func TestOrdinalSuffixRuleFilter(t *testing.T) {
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.OrdinalSuffixFilter")
	m := rules.NewRuleMatch(nil, nil, 0, 3, "msg")
	m.SetSuggestedReplacement("1nd")
	out := f.AcceptRuleMatch(m, map[string]string{"ignored": "ignored"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"1st"}, out.GetSuggestedReplacements())
}

func TestAdverbRuleFilter(t *testing.T) {
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.AdverbFilter")
	m := rules.NewRuleMatch(nil, nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"adverb": "quickly", "noun": "car"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"quick car"}, out.GetSuggestedReplacements())
}

func TestFutureDateRuleFilter(t *testing.T) {
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.FutureDateFilter")
	// IsTest() may force 2014-01-01 as today for rules.FutureDateFilterCore
	m := rules.NewRuleMatch(nil, nil, 0, 5, "msg")
	// Past relative to any reasonable now
	out := f.AcceptRuleMatch(m, map[string]string{"year": "2000", "month": "1", "day": "1"}, 0, nil, nil)
	require.Nil(t, out, "past date dropped")
	// Far future
	out = f.AcceptRuleMatch(m, map[string]string{"year": "2099", "month": "December", "day": "31"}, 0, nil, nil)
	require.NotNil(t, out, "future date kept")
}

func TestDateCheckRuleFilter_WrongWeekday(t *testing.T) {
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.DateCheckFilter")
	// Monday 2014-10-07 was a Tuesday (actual) — claimed Monday should fire
	// pattern tokens: weekDay, day, month, year at 1-based 1,2,3,4
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("Monday", 0),
		atr("7", 7),
		atr("October", 9),
		atr("2014", 17),
	}
	// tokenPositions: one each
	pos := []int{1, 1, 1, 1}
	m := rules.NewRuleMatch(nil, languagetool.NewAnalyzedSentence(toks), 0, 21, "The date is not a {day}, but a {realDay}.")
	out := f.AcceptRuleMatch(m, map[string]string{
		"weekDay": "1", "day": "2", "month": "3", "year": "4",
	}, 0, toks, pos)
	require.NotNil(t, out, "weekday mismatch should keep match")
	require.Contains(t, out.Message, "Tuesday")
}

func atr(tok string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(tok, nil, nil), start)
}

func TestPatternRuleLoader_ENFiltersLoad(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="ORD" name="ord">
      <pattern><token>1</token><token>nd</token></pattern>
      <filter class="org.languagetool.rules.en.OrdinalSuffixFilter" args="ignored:ignored"/>
      <message>m</message>
      <suggestion>1nd</suggestion>
    </rule>
    <rule id="ADV" name="adv">
      <pattern><token>a</token><token>quickly</token><token>car</token></pattern>
      <filter class="org.languagetool.rules.en.AdverbFilter" args="adverb:\2 noun:\3"/>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := patterns.NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 2)
	require.NotNil(t, ars[0].Filter)
	require.NotNil(t, ars[1].Filter)
}
