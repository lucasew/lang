package ga

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java rule demo sentences (addExamplePair) — correction = fixed marker span.
func TestGA_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"botún"}, NewMorfologikIrishSpellerRule().GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Dámaicléas"}, NewPeopleRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"mBéal Feirste"}, NewIrishSpecificCaseRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"mí-úsáid"}, NewCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"súisí"}, NewEnglishHomophoneRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"mná"}, NewDativePluralStandardReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"baol"}, NewPrestandardReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"beirt"}, NewDhaNoBeirtRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"bhur"}, NewIrishReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"ullamh"}, NewIrishFGBEqReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
