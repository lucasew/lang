package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLinguServices_Defaults(t *testing.T) {
	l := NewLinguServices()
	syn := l.GetSynonyms("word", "en")
	require.NotNil(t, syn) // Java: new ArrayList, never null
	require.Empty(t, syn)
	require.False(t, l.IsCorrectSpell("word", "en"))
	require.Equal(t, 0, l.GetNumberOfSyllables("word", "en"))
	l.SetThesaurusRelevantRule("RULE")
	require.Equal(t, "RULE", l.ThesaurusRelevantRule)
}
