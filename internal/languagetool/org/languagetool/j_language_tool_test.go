package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageToolSurface(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	require.Equal(t, "en-US", lt.GetLanguageCode())
	require.Equal(t, ModeAll, lt.GetMode())
	lt.SetMode(ModeTextLevelOnly)
	require.Equal(t, ModeTextLevelOnly, lt.GetMode())
	require.Equal(t, SentenceStartTagName, "SENT_START")
	require.Equal(t, PatternFile, "grammar.xml")
}
