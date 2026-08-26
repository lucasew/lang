package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRemoteRuleResult(t *testing.T) {
	s1 := &languagetool.AnalyzedSentence{}
	s2 := &languagetool.AnalyzedSentence{}
	m := NewRuleMatch(NewFakeRule("R"), s1, 0, 1, "msg")
	r := NewRemoteRuleResult(true, true, true, []*RuleMatch{m}, []*languagetool.AnalyzedSentence{s1, s2})
	require.True(t, r.IsRemote())
	require.Len(t, r.MatchesForSentence(s1), 1)
	require.NotNil(t, r.MatchesForSentence(s2))
	require.Empty(t, r.MatchesForSentence(s2))
	require.Nil(t, r.MatchesForSentence(&languagetool.AnalyzedSentence{}))
}
