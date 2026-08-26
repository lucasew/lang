package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEN_Confusion_ExamplePair(t *testing.T) {
	require.Equal(t, []string{"brakes"}, NewEnglishConfusionProbabilityRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
