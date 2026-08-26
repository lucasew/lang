package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTestRemoteRule(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "TEST_REMOTE_RULE"
	cfg.Options["waitTime"] = "0"
	r := NewTestRemoteRule(cfg)
	require.Equal(t, "TEST_REMOTE_RULE", r.GetID())
	s := languagetool.AnalyzePlain("hi")
	res := r.Execute([]*languagetool.AnalyzedSentence{s})
	require.True(t, res.IsSuccess())
	require.Len(t, res.GetMatches(), 1)
	fb := r.Fallback([]*languagetool.AnalyzedSentence{s})
	require.False(t, fb.IsSuccess())
	require.Empty(t, fb.GetMatches())
}
