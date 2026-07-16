package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAbstractCheckCaseRule(t *testing.T) {
	r := NewAbstractCheckCaseRule("CHECK_CASE", "case check")
	require.True(t, r.CheckingCase)
	require.Equal(t, "CHECK_CASE", r.ID)
}
