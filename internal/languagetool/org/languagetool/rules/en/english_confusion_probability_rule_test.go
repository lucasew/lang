package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishConfusionProbabilityRule(t *testing.T) {
	r := NewEnglishConfusionProbabilityRule(nil)
	require.Equal(t, EnglishConfusionRuleID, r.GetID())
}
