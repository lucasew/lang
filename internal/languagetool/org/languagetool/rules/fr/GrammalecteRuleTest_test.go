package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGrammalecteRule_IgnoredRuleIds(t *testing.T) {
	_, ok := GrammalecteIgnoreRules["tab_fin_ligne"]
	require.True(t, ok)
	_, ok = GrammalecteIgnoreRules["nbsp_avant_deux_points"]
	require.True(t, ok)
	r := NewGrammalecteRule("http://localhost:8080")
	r.Post = func(text string) ([]GrammalecteMatch, error) {
		return []GrammalecteMatch{
			{RuleID: "tab_fin_ligne", Start: 0, End: 1, Message: "ignored"},
			{RuleID: "real_typo", Start: 0, End: 3, Message: "err", Suggestions: []string{"ok"}},
		}, nil
	}
	ms, err := r.Match(languagetool.AnalyzePlain("foo"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
}
