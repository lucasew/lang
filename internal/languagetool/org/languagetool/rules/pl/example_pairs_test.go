package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java rule demo sentences (addExamplePair) — correction = fixed marker span.
func TestPL_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"Rabce-Zdroju"}, NewCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"się"}, NewSimpleReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"powoli"}, NewPolishWordRepeatRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Grapefruit"}, NewWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"„"}, NewPolishUnpairedBracketsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"błędem"}, NewMorfologikPolishSpellerRule().GetIncorrectExamples()[0].GetCorrections())
}
