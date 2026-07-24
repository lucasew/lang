package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPT_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"duradouro"}, NewPortugueseWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"coupé"}, NewPortugueseDiacriticsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"duna"}, NewPortugueseRedundancyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"abstrato"}, NewPortugueseAgreementReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"infligiu"}, NewPortugueseWrongWordInContextRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"hospedeira de bordo"}, NewPortugalPortugueseReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"aeromoça"}, NewBrazilianPortugueseReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"é"}, NewPortugueseWordRepeatRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"what's"}, NewEnglishContractionSpellingRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
