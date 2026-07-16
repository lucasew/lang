package tools

// Twin of EN ToolsTest.testCorrect — full JLT deferred; CorrectTextFromMatches surface
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTools_lang_en_Correct(t *testing.T) {
	// "This is an test." → "This is a test." (EN_A_VS_AN)
	// "an" at positions 8-10
	src := "This is an test."
	require.Equal(t, "an", src[8:10])
	got := CorrectTextFromMatches(src, []TextMatch{
		{FromPos: 8, ToPos: 10, SuggestedReplacements: []string{"a"}},
	})
	require.Equal(t, "This is a test.", got)
}
