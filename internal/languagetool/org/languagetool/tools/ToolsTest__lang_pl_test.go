package tools

// Twin of PL ToolsTest — CorrectTextFromMatches surface (no languagetool import: cycle).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTools_lang_pl_Check(t *testing.T) {
	// soft: correct path after check would produce this TextMatch
	src := "To jest jest problem."
	got := CorrectTextFromMatches(src, []TextMatch{
		{FromPos: 8, ToPos: 13, SuggestedReplacements: []string{""}},
	})
	require.Equal(t, "To jest problem.", got)
	require.NotEqual(t, src, got)
}

func TestTools_lang_pl_Correct(t *testing.T) {
	got := CorrectTextFromMatches("To jest jest problem.", []TextMatch{
		{FromPos: 8, ToPos: 13, SuggestedReplacements: []string{""}},
	})
	require.Equal(t, "To jest problem.", got)

	src := "To jest jest problem. Ale to juz juz nie."
	require.Equal(t, "jest ", src[8:13])
	firstJuz := strings.Index(src, "juz")
	secondJuz := strings.Index(src[firstJuz+3:], "juz")
	require.GreaterOrEqual(t, firstJuz, 0)
	require.GreaterOrEqual(t, secondJuz, 0)
	secondJuz += firstJuz + 3
	end := secondJuz + 3
	if end < len(src) && src[end] == ' ' {
		end++
	}
	got = CorrectTextFromMatches(src, []TextMatch{
		{FromPos: 8, ToPos: 13, SuggestedReplacements: []string{""}},
		{FromPos: secondJuz, ToPos: end, SuggestedReplacements: []string{""}},
	})
	require.Equal(t, "To jest problem. Ale to juz nie.", got)
}
