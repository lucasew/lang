package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java SimpleReplaceProfanityRule: STYLE, Style ITS, picky, wiktionary URL, no suggestions.
func TestSimpleReplaceProfanityRule_Metadata(t *testing.T) {
	rule := NewSimpleReplaceProfanityRule(nil)
	require.Equal(t, "PROFANITY", rule.GetID())
	require.Equal(t, "Profanity", rule.GetDescription())
	require.Contains(t, rule.GetURL(), "English_offensive_terms")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.True(t, rule.HasTag(rules.TagPicky))
	require.NotNil(t, rule.RuleHasSuggestions)
	require.False(t, *rule.RuleHasSuggestions)
}

func TestSimpleReplaceProfanityRule_MatchNoSuggestions(t *testing.T) {
	rule := NewSimpleReplaceProfanityRule(nil)
	// File entry "albo" is key-only (no replacements).
	ms := rule.Match(languagetool.AnalyzePlain("That albo was rude."))
	require.NotEmpty(t, ms)
	require.Empty(t, ms[0].GetSuggestedReplacements())
	require.Contains(t, ms[0].GetMessage(), "offensive")
}
