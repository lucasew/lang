package en

import (
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
	} {
		require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
		require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(class), class)
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
