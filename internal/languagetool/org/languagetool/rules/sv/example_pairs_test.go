package sv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java WordCoherencyRule: multi-marker fixed; first correction = mejl
func TestSV_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"mejl"}, NewWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
