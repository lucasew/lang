package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseRuleExceptions(t *testing.T) {
	ex := CaseRuleExceptions()
	require.NotEmpty(t, ex)
	require.True(t, IsCaseRuleException("Absolut Vodka"))
	require.False(t, IsCaseRuleException("xyzzy not an exception"))
}
