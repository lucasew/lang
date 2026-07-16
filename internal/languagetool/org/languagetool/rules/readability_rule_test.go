package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFleschAndSyllables(t *testing.T) {
	require.Equal(t, 1, CountSyllablesEN("cat"))
	require.GreaterOrEqual(t, CountSyllablesEN("beautiful"), 2)
	f := FleschReadingEase(1, 10, 10)
	require.Greater(t, f, 50.0)
	require.Equal(t, 0, ReadabilityLevel(95))
	require.Equal(t, 6, ReadabilityLevel(10))
}

func TestReadabilityRule_Evaluate(t *testing.T) {
	r := NewReadabilityRule(false, 3)
	words := strings.Fields("The cat sat on the mat and looked at the dog carefully nearby more words")
	require.GreaterOrEqual(t, len(words), 10)
	_, _, _ = r.EvaluateParagraph(1, words)
	require.Equal(t, "READABILITY_RULE_DIFFICULT", r.GetID())
	require.Equal(t, "READABILITY_RULE_SIMPLE", NewReadabilityRule(true, 0).GetID())
}
