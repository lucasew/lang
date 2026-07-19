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

// Java Rule.addExamplePair: fixed marker span becomes incorrect example correction.
func TestBaseRule_AddExamplePair(t *testing.T) {
	r := &BaseRule{ID: "E"}
	r.AddExamplePair(
		Wrong("See <marker>err</marker> here."),
		Fixed("See <marker>fix</marker> here."),
	)
	require.Len(t, r.GetIncorrectExamples(), 1)
	require.Equal(t, []string{"fix"}, r.GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, "See <marker>fix</marker> here.", r.GetCorrectExamples()[0].GetExample())
	r.SetExamplePair(Wrong("<marker>a</marker>"), Fixed("<marker>b</marker>"))
	require.Len(t, r.GetIncorrectExamples(), 1)
	require.Equal(t, []string{"b"}, r.GetIncorrectExamples()[0].GetCorrections())
}
