package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishPatternRule_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("en").Analyze("x"))
}

func TestEnglishPatternRule_L2Languages(t *testing.T) {
	// L2 false-friend rule metadata constructors
	require.Equal(t, "EN_FOR_DE_SPEAKERS_FALSE_FRIENDS", NewEnglishForGermansFalseFriendRule().GetID())
}

func TestEnglishPatternRule_Bug(t *testing.T) {
	// Historical bug regression: analyze still works on edge strings
	require.NotEmpty(t, languagetool.NewJLanguageTool("en").Analyze("a"))
}
