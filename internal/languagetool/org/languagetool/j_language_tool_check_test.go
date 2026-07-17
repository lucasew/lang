package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageTool_CheckWordRepeat(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddChecker(SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("This is fine."))
	m := lt.Check("This is is wrong.")
	require.Len(t, m, 1)
	require.Equal(t, "WORD_REPEAT_RULE", m[0].RuleID)
	require.Greater(t, m[0].ToPos, m[0].FromPos)
}

func TestJLanguageTool_CheckCancel(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddChecker(SimpleWordRepeatChecker(""))
	lt.Cancelled = func() bool { return true }
	require.Empty(t, lt.Check("is is"))
}

func TestJLanguageTool_UnknownWords(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.SetListUnknownWords(true)
	lt.IsKnownWord = KnownWordSet("This", "is", "a", "text")
	_ = lt.Check("This is a xyzzy text")
	require.Equal(t, []string{"xyzzy"}, lt.GetUnknownWords())
}

func TestCleanOverlappingLocalMatches(t *testing.T) {
	// non-overlap preserved
	require.Len(t, CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 2, Priority: 1},
		{FromPos: 3, ToPos: 5, Priority: 1},
	}), 2)
	// higher priority wins
	got := CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "a", Priority: 1},
		{FromPos: 1, ToPos: 3, RuleID: "b", Priority: 5},
	})
	require.Len(t, got, 1)
	require.Equal(t, "b", got[0].RuleID)
}
