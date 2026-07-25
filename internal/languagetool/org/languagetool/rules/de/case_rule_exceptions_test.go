package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseRuleExceptions(t *testing.T) {
	ex := CaseRuleExceptions()
	require.NotEmpty(t, ex)
	// case_rule_exceptions.txt
	require.True(t, IsCaseRuleException("Absolut Vodka"))
	// eigennamen_gross.txt (Java loadExceptions first path)
	require.True(t, IsCaseRuleException("Europäische Zentralbank"))
	require.True(t, IsCaseRuleException("Heilige Vater"))
	require.False(t, IsCaseRuleException("xyzzy not an exception"))
}
