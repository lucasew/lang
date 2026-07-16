package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestGetResolvedArguments_Literal(t *testing.T) {
	args := GetResolvedArguments("wordFrom:1 hasTypographicalApostrophe:true", nil, 0, nil)
	require.Equal(t, "1", args["wordFrom"])
	require.Equal(t, "true", args["hasTypographicalApostrophe"])
}

func TestGetResolvedArguments_Backref(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("alpha", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("beta", nil, nil)),
	}
	args := GetResolvedArguments(`wordFrom:\1 other:\2`, toks, 0, []int{1, 1})
	require.Equal(t, "alpha", args["wordFrom"])
	require.Equal(t, "beta", args["other"])
}

type ruleFilterAdapter func(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch

func (f ruleFilterAdapter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return f(match, arguments, patternTokenPos, patternTokens, tokenPositions)
}

func TestRuleFilterEvaluator_Run(t *testing.T) {
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
