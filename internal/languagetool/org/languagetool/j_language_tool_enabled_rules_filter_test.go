package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnabledRulesForFilters_IncludesActiveRegistered(t *testing.T) {
	lt := NewJLanguageTool("fr")
	lt.AddRuleChecker("APOS_TYP", SimpleWordRepeatChecker("APOS_TYP"))
	lt.AddRuleChecker("OTHER", SimpleWordRepeatChecker("OTHER"))
	lt.DisableRule("OTHER")
	en := lt.enabledRulesForFilters()
	require.Contains(t, en, "APOS_TYP")
	require.NotContains(t, en, "OTHER")
	// explicit EnableRule also present
	lt.EnableRule("APOS_TYP")
	en2 := lt.enabledRulesForFilters()
	require.Contains(t, en2, "APOS_TYP")
}

func TestEnableRule_ClearsDefaultOff(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.MarkDefaultOff("X")
	require.True(t, lt.IsRuleDisabled("X"))
	require.Contains(t, lt.GetDefaultOffRuleIDs(), "X")
	lt.EnableRule("X")
	require.False(t, lt.IsRuleDisabled("X"))
	require.NotContains(t, lt.GetDefaultOffRuleIDs(), "X")
}
