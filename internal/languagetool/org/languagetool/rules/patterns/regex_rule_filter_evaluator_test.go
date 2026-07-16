package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestResolveFilterArguments(t *testing.T) {
	m := ResolveFilterArguments("antipatterns:foo|bar other:1")
	require.Equal(t, "foo|bar", m["antipatterns"])
	require.Equal(t, "1", m["other"])
	require.Panics(t, func() { ResolveFilterArguments("bad") })
}

func TestRegexRuleFilterEvaluator(t *testing.T) {
	var got map[string]string
	f := RegexRuleFilterFunc(func(match *rules.RuleMatch, arguments map[string]string,
		_ *languagetool.AnalyzedSentence, _ string) *rules.RuleMatch {
		got = arguments
		return match
	})
	ev := NewRegexRuleFilterEvaluator(f)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	out := ev.RunFilter("a:1 b:2", m, nil, "pat")
	require.Equal(t, m, out)
	require.Equal(t, "1", got["a"])
	require.Equal(t, "2", got["b"])
}
