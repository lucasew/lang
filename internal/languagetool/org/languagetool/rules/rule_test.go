package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCategoryIds(t *testing.T) {
	require.True(t, CategoryIds.Grammar.Equals(CategoryGrammar))
	require.Equal(t, "TYPOS", CategoryIds.Typos.String())
}

func TestBaseRule(t *testing.T) {
	r := &BaseRule{ID: "X", Description: "desc", DefaultOff: true}
	require.Equal(t, "X", r.GetID())
	require.Equal(t, "desc", r.GetDescription())
	require.True(t, r.IsDefaultOff())
	r.SetCategory(NewCategory(CategoryGrammar, "Grammar"))
	require.Equal(t, "Grammar", r.GetCategory().GetName())
}
