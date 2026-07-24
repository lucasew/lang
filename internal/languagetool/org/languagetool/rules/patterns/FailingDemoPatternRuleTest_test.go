package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFailingDemoPatternRule_RulesWithErrors1(t *testing.T) {
	// Invalid XML should error
	_, err := NewPatternRuleLoader().GetRulesFromString(`<rules><broken`, "bad.xml", "xx")
	require.Error(t, err)
}

func TestFailingDemoPatternRule_RulesWithErrors2(t *testing.T) {
	// Empty rules document may parse with zero rules
	rules, err := NewPatternRuleLoader().GetRulesFromString(`<?xml version="1.0"?><rules></rules>`, "empty.xml", "xx")
	require.NoError(t, err)
	require.Empty(t, rules)
}
