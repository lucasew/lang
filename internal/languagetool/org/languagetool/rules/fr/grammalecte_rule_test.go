package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGrammalecteRule(t *testing.T) {
	r := NewGrammalecteRule("http://localhost:8080")
	r.Post = func(text string) ([]GrammalecteMatch, error) {
		return []GrammalecteMatch{
			{RuleID: "typo_foo", Start: 0, End: 3, Message: "err", Suggestions: []string{"bar"}},
			{RuleID: "tab_fin_ligne", Start: 0, End: 1, Message: "ignored"},
		}, nil
	}
	sent := languagetool.AnalyzePlain("foo")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, "bar", ms[0].GetSuggestedReplacements()[0])
}

func TestParseGrammalecteJSON(t *testing.T) {
	raw := `{"errors":[{"rule":"x","start":1,"end":2,"message":"m"}]}`
	ms, err := ParseGrammalecteJSON([]byte(raw))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, "x", ms[0].RuleID)
}
