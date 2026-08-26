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

// Twin of Rule tags / defaultTempOff / priority / premium / url / ITS surface.
func TestBaseRule_RuleMetaFields(t *testing.T) {
	r := &BaseRule{ID: "META"}
	// tags
	require.Empty(t, r.GetTags())
	r.AddTags([]string{"picky", "PICKY", "academic"})
	require.Equal(t, []Tag{TagPicky, TagAcademic}, r.GetTags())
	require.True(t, r.HasTag(TagPicky))
	require.True(t, r.HasTag(TagAcademic))
	r.SetTags([]Tag{TagPicky})
	require.Equal(t, []Tag{TagPicky}, r.GetTags())
	r.SetTags(nil)
	require.Empty(t, r.GetTags())
	// defaultTempOff
	r.SetDefaultTempOff()
	require.True(t, r.IsDefaultOff())
	require.True(t, r.IsDefaultTempOff())
	r.SetDefaultOn()
	require.False(t, r.IsDefaultOff())
	// priority / premium / url / ITS
	r.SetPriority(7)
	require.Equal(t, 7, r.GetPriority())
	r.SetPremium(true)
	require.True(t, r.IsPremium())
	r.SetURL("https://example.com/r")
	require.Equal(t, "https://example.com/r", r.GetURL())
	r.SetLocQualityIssueType(ITSMisspelling)
	require.Equal(t, ITSMisspelling, r.GetLocQualityIssueType())
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
