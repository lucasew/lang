package eval

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFMeasure(t *testing.T) {
	// classic F1
	f1 := GetFMeasure(0.5, 0.5, 1.0)
	require.InDelta(t, 0.5, f1, 1e-9)

	// F0.5 from RealWordCorpusEvaluator historical note: p=0.71 r=0.18
	f05 := GetWeightedFMeasure(0.71, 0.18)
	require.True(t, f05 > 0.4 && f05 < 0.5, "got %v", f05)

	require.Equal(t, 0.0, GetFMeasure(0, 0, 0.5))
	require.False(t, math.IsNaN(GetWeightedFMeasure(1, 0)))
}

func TestPrecisionRecall_F05(t *testing.T) {
	pr := NewPrecisionRecall(0.58, 0.14)
	require.InDelta(t, GetWeightedFMeasure(0.58, 0.14), pr.F05(), 1e-12)
}
