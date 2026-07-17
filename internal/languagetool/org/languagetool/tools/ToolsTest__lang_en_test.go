package tools

// Twin of EN ToolsTest.testCorrect — CorrectTextFromMatches surface (JLT inject in languagetool package).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTools_lang_en_Correct(t *testing.T) {
	// "This is an test." → "This is a test." (EN_A_VS_AN)
	src := "This is an test."
	require.Equal(t, "an", src[8:10])
	got := CorrectTextFromMatches(src, []TextMatch{
		{FromPos: 8, ToPos: 10, SuggestedReplacements: []string{"a"}},
	})
	require.Equal(t, "This is a test.", got)

	// spelling style
	got2 := CorrectTextFromMatches("A speling error", []TextMatch{
		{FromPos: 2, ToPos: 9, SuggestedReplacements: []string{"spelling"}},
	})
	require.Equal(t, "A spelling error", got2)
}
