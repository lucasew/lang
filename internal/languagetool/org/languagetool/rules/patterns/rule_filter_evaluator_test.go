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

// Java Pattern.compile("\\s+") does not treat NBSP as whitespace.
func TestGetResolvedArguments_NBSPNotSeparator(t *testing.T) {
	// key\u00a0with:value is one arg with key "key\u00a0with"
	args := GetResolvedArguments("key\u00a0with:val other:x", nil, 0, nil)
	require.Equal(t, "val", args["key\u00a0with"])
	require.Equal(t, "x", args["other"])
	// multi ASCII spaces still split
	args2 := GetResolvedArguments("a:1   b:2", nil, 0, nil)
	require.Equal(t, "1", args2["a"])
	require.Equal(t, "2", args2["b"])
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
