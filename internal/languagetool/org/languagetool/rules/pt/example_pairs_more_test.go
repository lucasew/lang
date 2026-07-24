package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Remaining Java addExamplePair demos not covered by earlier PT example commits.
func TestPT_ExamplePairs_More(t *testing.T) {
	require.Equal(t, []string{"XYZ"}, NewPortugueseWeaselWordsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"currículo"}, NewPortugueseBarbarismsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Raramente acontece"}, NewPortugueseWordinessRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"quente"}, NewPortugueseClicheRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Foi"}, NewPortugueseWordRepeatBeginningRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
