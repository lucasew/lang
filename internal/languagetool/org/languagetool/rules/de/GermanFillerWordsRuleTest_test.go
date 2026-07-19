package de

// Twin of GermanFillerWordsRuleTest (statistic % mode + minPercent=0).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanFillerWordsRule_Rule(t *testing.T) {
	// Java default minPercent=8: short sentences with fillers fire.
	rule := NewGermanFillerWordsRule(nil)
	require.Equal(t, "FILLER_WORDS_DE", rule.GetID())
	require.Equal(t, 8, rule.MinPercent)

	matches := rule.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich ein Füllwort."))
	require.Equal(t, 1, len(matches))
	matches = rule.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich relativ viele Füllwörter."))
	require.Equal(t, 2, len(matches))
	// longer sentence: share of fillers drops under 8%
	matches = rule.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich ein Füllwort, aber es sind nicht genug um angezeigt zu werden."))
	require.Equal(t, 0, len(matches))
	// direct speech excluded by default
	matches = rule.Match(languagetool.AnalyzePlain("»Der Satz enthält augenscheinlich ein Füllwort«"))
	require.Equal(t, 0, len(matches))

	// minPercent=0: show all fillers (including direct speech when MinPercent==0 path)
	rule0 := NewGermanFillerWordsRuleWithMinPercent(nil, 0)
	matches = rule0.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich ein Füllwort, aber es sind nicht genug um angezeigt zu werden."))
	require.Equal(t, 1, len(matches))
	// sentence-start token (index 1) is exception
	matches = rule0.Match(languagetool.AnalyzePlain("Allerdings war es kalt."))
	require.Equal(t, 0, len(matches))
	matches = rule0.Match(languagetool.AnalyzePlain("Es war allerdings kalt."))
	require.Equal(t, 1, len(matches))
	// Two-word exception: immer wieder
	matches = rule0.Match(languagetool.AnalyzePlain("Das passiert immer wieder hier."))
	require.Equal(t, 0, len(matches))
	require.Equal(t, "Statistische Stilanalyse: Füllwörter", rule.GetDescription())
}
