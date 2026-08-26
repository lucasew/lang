package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleException(t *testing.T) {
	e := NewRuleException(RuleExceptionException)
	require.Equal(t, RuleExceptionException, e.Type)
	s := NewRuleExceptionSkip(3)
	require.Equal(t, RuleExceptionSkip, s.Type)
	require.Equal(t, 3, s.Skip)
}
