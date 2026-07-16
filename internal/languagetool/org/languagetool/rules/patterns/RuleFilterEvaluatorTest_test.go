package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestRuleFilterEvaluator_TestDuplicateKey(t *testing.T) {
	// last key wins or panic on invalid — our impl overwrites same key
	args := GetResolvedArguments("k:v1 k:v2", nil, 0, nil)
	require.Equal(t, "v2", args["k"])
}

func TestRuleFilterEvaluator_TestTooLargeBackRef(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("alpha", nil, nil)),
	}
	require.Panics(t, func() {
		GetResolvedArguments(`wordFrom:\9`, toks, 0, []int{1})
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
