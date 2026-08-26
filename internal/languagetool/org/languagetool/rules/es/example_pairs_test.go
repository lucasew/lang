package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestES_ExamplePairs(t *testing.T) {
	// Java CompoundRule
	require.Equal(t, []string{"Guinea-Conakri"}, NewCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
	// Java QuestionMarkRule
	require.Equal(t, []string{"¿Qué"}, NewQuestionMarkRule(nil).GetIncorrectExamples()[0].GetCorrections())
	// Java SpanishWordRepeatBeginningRule
	require.Contains(t, NewSpanishWordRepeatBeginningRule(nil).GetIncorrectExamples()[0].GetExample(), "Asimismo")
}
