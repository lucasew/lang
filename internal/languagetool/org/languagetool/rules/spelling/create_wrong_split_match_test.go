package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Twin of SpellingCheckRule.createWrongSplitMatch control flow + messages.
func TestCreateWrongSplitMatch_MessagesAndSpan(t *testing.T) {
	sent := languagetool.AnalyzePlain("foo bar")
	rule := rules.NewFakeRule("SPELL")
	// prev match at prevPos=0 covering "foo"
	prev := rules.NewRuleMatch(rule, sent, 0, 3, "old")
	matches := []*rules.RuleMatch{prev}
	// merge when last.fromPos == prevPos
	m := CreateWrongSplitMatch(rule, sent, &matches, 4, "bar", "foo", "bar", 0)
	require.Empty(t, matches, "previous match at prevPos removed")
	require.Equal(t, SpellingMessage, m.Message)
	require.Equal(t, DescSpellingShort, m.GetShortMessage())
	require.Equal(t, rules.RuleMatchTypeUnknownWord, m.GetType())
	require.Equal(t, 0, m.GetFromPos())
	// pos 4 + len("bar")=3 → 7
	require.Equal(t, 7, m.GetToPos())
	require.Equal(t, []string{"foo bar"}, m.GetSuggestedReplacements())
}

func TestCreateWrongSplitMatch_NoRemoveWhenPrevDiffers(t *testing.T) {
	sent := languagetool.AnalyzePlain("aa bb cc")
	rule := rules.NewFakeRule("SPELL")
	prev := rules.NewRuleMatch(rule, sent, 0, 2, "old")
	matches := []*rules.RuleMatch{prev}
	m := CreateWrongSplitMatch(rule, sent, &matches, 3, "bb", "a", "bb", 99)
	require.Len(t, matches, 1, "prev fromPos != prevPos keeps list")
	require.Equal(t, SpellingMessage, m.Message)
	require.Equal(t, []string{"a bb"}, m.GetSuggestedReplacements())
}

func TestFilterDupes(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, FilterDupes([]string{"a", "b", "a", "b"}))
	require.Equal(t, []int{1, 2}, FilterDupes([]int{1, 2, 1}))
}

func TestNewSpellingRuleMatch(t *testing.T) {
	sent := languagetool.AnalyzePlain("x")
	m := NewSpellingRuleMatch(rules.NewFakeRule("S"), sent, 0, 1)
	require.Equal(t, SpellingMessage, m.Message)
	require.Equal(t, DescSpellingShort, m.ShortMessage)
	require.Equal(t, rules.RuleMatchTypeUnknownWord, m.GetType())
}

// Message constants match MessagesBundle_en.properties keys.
func TestSpellingMessageConstants_JavaBundle(t *testing.T) {
	require.Equal(t, "Possible spelling mistake found.", SpellingMessage)
	require.Equal(t, "Possible spelling mistake", DescSpelling)
	require.Equal(t, "Spelling mistake", DescSpellingShort)
	require.Equal(t, "Possible spelling mistake (without suggestions)", DescSpellingNoSuggestions)
}
