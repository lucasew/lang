package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLinguServices_Defaults(t *testing.T) {
	l := NewLinguServices()
	require.Empty(t, l.GetSynonyms("word", "en"))
	require.False(t, l.IsCorrectSpell("word", "en"))
	require.Equal(t, 0, l.GetNumberOfSyllables("word", "en"))
	l.SetThesaurusRelevantRule("RULE")
	require.Equal(t, "RULE", l.ThesaurusRelevantRule)
}
