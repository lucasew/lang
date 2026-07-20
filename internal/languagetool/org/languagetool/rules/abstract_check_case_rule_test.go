package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAbstractCheckCaseRule(t *testing.T) {
	r := NewAbstractCheckCaseRule(nil, "CHECK_CASE", "case check")
	require.True(t, r.CheckingCase)
	require.Equal(t, "CHECK_CASE", r.ID)
	require.Equal(t, ITSTypographical, r.IssueType)
	require.NotNil(t, r.Category)
	require.Equal(t, "CASING", r.Category.GetID().String())
}
