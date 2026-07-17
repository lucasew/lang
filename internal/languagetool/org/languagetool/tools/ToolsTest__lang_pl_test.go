package tools

// Twin of PL ToolsTest — full Polish JLT deferred; CorrectTextFromMatches surface.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTools_lang_pl_Check(t *testing.T) {
	// soft: full Polish check deferred — CorrectTextFromMatches stands in for post-check fixup
	// (Java Tools.check + correct); surface asserts correction API on PL sample.
	src := "To jest jest problem."
	got := CorrectTextFromMatches(src, []TextMatch{
		{FromPos: 8, ToPos: 13, SuggestedReplacements: []string{""}},
	})
	require.Equal(t, "To jest problem.", got)
	require.NotEqual(t, src, got)
}

func TestTools_lang_pl_Correct(t *testing.T) {
	// Ports Tools.correctText suggestion application without full PL rules.
	got := CorrectTextFromMatches("To jest jest problem.", []TextMatch{
		{FromPos: 8, ToPos: 13, SuggestedReplacements: []string{""}},
	})
	require.Equal(t, "To jest problem.", got)

	// Two successive repeats in one text (Java multi-sentence style).
	src := "To jest jest problem. Ale to juz juz nie."
	// second "jest "
	require.Equal(t, "jest ", src[8:13])
	// second "juz"
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
