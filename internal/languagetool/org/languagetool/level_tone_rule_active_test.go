package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRuleActiveForLevelAndToneTags_Picky(t *testing.T) {
	picky := RuleLevelToneMeta{IsPicky: true}
	require.False(t, IsRuleActiveForLevelAndToneTags(picky, LevelDefault, nil))
	require.True(t, IsRuleActiveForLevelAndToneTags(picky, LevelPicky, nil))
}

func TestIsRuleActiveForLevelAndToneTags_NoTone(t *testing.T) {
	// rule without tone tags always active (aside from picky)
	plain := RuleLevelToneMeta{}
	require.True(t, IsRuleActiveForLevelAndToneTags(plain, LevelDefault, map[ToneTag]struct{}{ToneNoToneRule: {}}))
	// goal-specific rule disabled when ALL_WITHOUT_GOAL_SPECIFIC (default empty set)
	goal := RuleLevelToneMeta{GoalSpecific: true, ToneTags: []ToneTag{ToneClarity}}
	require.False(t, IsRuleActiveForLevelAndToneTags(goal, LevelDefault, nil))
	// non-goal tone rule with empty toneTags set uses ALL_WITHOUT_GOAL_SPECIFIC → keep non-goal
	clarity := RuleLevelToneMeta{ToneTags: []ToneTag{ToneClarity}, GoalSpecific: false}
	require.True(t, IsRuleActiveForLevelAndToneTags(clarity, LevelDefault, nil))
}

func TestIsRuleActiveForLevelAndToneTags_ExplicitTone(t *testing.T) {
	rule := RuleLevelToneMeta{ToneTags: []ToneTag{ToneFormal}}
	enabled := map[ToneTag]struct{}{ToneFormal: {}}
	require.True(t, IsRuleActiveForLevelAndToneTags(rule, LevelDefault, enabled))
	require.False(t, IsRuleActiveForLevelAndToneTags(rule, LevelDefault, map[ToneTag]struct{}{ToneInformal: {}}))
	// ALL_TONE_RULES enables real tags including formal
	require.True(t, IsRuleActiveForLevelAndToneTags(rule, LevelDefault, map[ToneTag]struct{}{ToneAllToneRules: {}}))
}

func TestFilterMatchesForLevelAndToneTags(t *testing.T) {
	ms := []LocalMatch{
		{RuleID: "A", IsPicky: true},
		{RuleID: "B", IsPicky: false},
	}
	out := FilterMatchesForLevelAndToneTags(ms, LevelDefault, nil)
	require.Len(t, out, 1)
	require.Equal(t, "B", out[0].RuleID)
	out = FilterMatchesForLevelAndToneTags(ms, LevelPicky, nil)
	require.Len(t, out, 2)
}
