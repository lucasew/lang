package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Twin of RuleFilterEvaluatorTest.testGetResolvedArguments
func TestRuleFilterEvaluator_GetResolvedArguments(t *testing.T) {
	pos := "pos"
	readingsList := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fake1", &pos, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fake2", &pos, nil)),
	}
	// Java: new RuleFilterEvaluator(null).getResolvedArguments(...)
	m := GetResolvedArguments(`year:\1 month:\2`, readingsList, -1, []int{1, 1})
	require.Equal(t, "fake1", m["year"])
	require.Equal(t, "fake2", m["month"])
	require.Len(t, m, 2)
}

// Twin of testGetResolvedArgumentsWithColon — value may contain ':'
func TestRuleFilterEvaluator_GetResolvedArgumentsWithColon(t *testing.T) {
	pos := "pos"
	readingsList := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fake1", &pos, nil)),
	}
	m := GetResolvedArguments("regex:(?:foo[xyz])bar", readingsList, -1, []int{1, 1})
	require.Equal(t, "(?:foo[xyz])bar", m["regex"])
	require.Len(t, m, 1)
}

// Twin of testDuplicateKey — two backrefs with same key → RuntimeException
func TestRuleFilterEvaluator_DuplicateKey(t *testing.T) {
	pos := "pos"
	readingsList := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fake1", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fake1", &pos, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fake2", &pos, nil)),
	}
	require.Panics(t, func() {
		GetResolvedArguments(`year:\1 year:\2`, readingsList, -1, []int{1, 2})
	})
}

// Twin of testNoBackReference
func TestRuleFilterEvaluator_NoBackReference(t *testing.T) {
	args := GetResolvedArguments("year:2 foo:bar", nil, -1, nil)
	require.Equal(t, map[string]string{"year": "2", "foo": "bar"}, args)
}

// Twin of testTooLargeBackRef — empty tokenPositions + \1…
func TestRuleFilterEvaluator_TooLargeBackRef(t *testing.T) {
	require.Panics(t, func() {
		GetResolvedArguments(`year:\1 month:\2 day:\3 weekDay:\4`, nil, -1, nil)
	})
}

func TestRuleFilterEvaluator_RunFilterPort(t *testing.T) {
	var got map[string]string
	ev := NewRuleFilterEvaluator(ruleFilterAdapter(func(m *rules.RuleMatch, a map[string]string, _ int,
		_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
		got = a
		return m
	}))
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	out := ev.RunFilter("k:v", m, nil, 0, nil)
	require.Equal(t, m, out)
	require.Equal(t, "v", got["k"])
}
