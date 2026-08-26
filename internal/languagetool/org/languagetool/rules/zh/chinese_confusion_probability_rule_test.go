package zh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseConfusionProbabilityRule(t *testing.T) {
	r := NewChineseConfusionProbabilityRule(nil)
	require.NotNil(t, r)
	require.NotNil(t, r.ConfusionProbabilityRule)
}
