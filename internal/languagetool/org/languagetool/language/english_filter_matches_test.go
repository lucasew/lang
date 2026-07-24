package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFilterEnglishRuleMatches_ContractionApostropheSpace(t *testing.T) {
	// error "'s" → suggestion "is" becomes " is"
	in := []languagetool.LocalMatch{
		{RuleID: "X", OriginalErrorStr: "'s more", Suggestions: []string{"is more", "has more"}},
	}
	out := FilterEnglishRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, []string{" is more", " has more"}, out[0].Suggestions)
}

func TestFilterEnglishRuleMatches_ContractionNt(t *testing.T) {
	in := []languagetool.LocalMatch{
		{RuleID: "X", OriginalErrorStr: "n't known", Suggestions: []string{"not known"}},
	}
	out := FilterEnglishRuleMatches(in)
	require.Equal(t, []string{" not known"}, out[0].Suggestions)
}

func TestFilterEnglishRuleMatches_NoDoubleSpaceIfAlreadySpaced(t *testing.T) {
	in := []languagetool.LocalMatch{
		{RuleID: "X", OriginalErrorStr: "'s more", Suggestions: []string{" is more"}},
	}
	out := FilterEnglishRuleMatches(in)
	require.Equal(t, []string{" is more"}, out[0].Suggestions)
}

func TestFilterEnglishRuleMatches_SkipWhenErrorShort(t *testing.T) {
	// len(errorStr) > 2 required
	in := []languagetool.LocalMatch{
		{RuleID: "X", OriginalErrorStr: "'s", Suggestions: []string{"is"}},
	}
	out := FilterEnglishRuleMatches(in)
	require.Equal(t, []string{"is"}, out[0].Suggestions)
}

func TestFilterEnglishRuleMatches_DedupSuggestions(t *testing.T) {
	in := []languagetool.LocalMatch{
		{RuleID: "X", OriginalErrorStr: "'s more", Suggestions: []string{"is more", "is more"}},
	}
	out := FilterEnglishRuleMatches(in)
	require.Equal(t, []string{" is more"}, out[0].Suggestions)
}

func TestFilterEnglishRuleMatches_GrammeLocaleViolation(t *testing.T) {
	in := []languagetool.LocalMatch{
		{RuleID: "EN_SIMPLE_REPLACE_PROGRAMME", Suggestions: []string{"program"}},
		{RuleID: "EN_SIMPLE_REPLACE_PROGRAMMES", Suggestions: []string{"programs"}},
		{RuleID: "EN_SIMPLE_REPLACE_OTHER", Suggestions: []string{"x"}},
	}
	out := FilterEnglishRuleMatches(in)
	require.Equal(t, "locale-violation", out[0].IssueType)
	require.Equal(t, "locale-violation", out[1].IssueType)
	require.Empty(t, out[2].IssueType)
}

func TestFilterEnglishRuleMatches_NoSurfaceNoContractionHack(t *testing.T) {
	// fail-closed: empty OriginalSurface → no invent leading space
	in := []languagetool.LocalMatch{
		{RuleID: "X", Suggestions: []string{"is"}},
	}
	out := FilterEnglishRuleMatches(in)
	require.Equal(t, []string{"is"}, out[0].Suggestions)
}
