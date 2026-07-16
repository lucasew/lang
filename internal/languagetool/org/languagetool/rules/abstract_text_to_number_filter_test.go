package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextToNumberFilter_Basic(t *testing.T) {
	f := &TextToNumberFilter{
		Numbers: map[string]float64{
			"dos": 2, "tres": 3, "cien": 100, "mil": 0, // mil is multiplier only
		},
		Multipliers: map[string]float64{
			"mil": 1000,
		},
		IsComma: func(s string) bool { return s == "coma" },
	}
	// Note: "mil" only in multipliers in real ES table.
	f.Numbers = map[string]float64{"dos": 2, "tres": 3, "cien": 100, "ciento": 100, "uno": 1}
	require.Equal(t, "2", f.ConvertTokens([]string{"dos"}))
	require.Equal(t, "100", f.ConvertTokens([]string{"cien"}))
	require.Equal(t, "2000", f.ConvertTokens([]string{"dos", "mil"}))
	require.Equal(t, "1000", f.ConvertTokens([]string{"mil"}))      // current=0 → 1 * mil
	require.Equal(t, "5", f.ConvertTokens([]string{"dos", "tres"})) // additive within segment
}

func TestTextToNumberFilter_DecimalAndPercent(t *testing.T) {
	f := &TextToNumberFilter{
		Numbers: map[string]float64{"dos": 2, "cinco": 5},
		IsComma: func(s string) bool { return s == "coma" },
		IsPercentage: func(tokens []string, i int) bool {
			return i > 0 && tokens[i] == "ciento" && tokens[i-1] == "por"
		},
	}
	require.Equal(t, "2.5", f.ConvertTokens([]string{"dos", "coma", "cinco"}))
	require.Equal(t, "2\u202F%", f.ConvertTokens([]string{"dos", "por", "ciento"}))
}

func TestTextToNumberFilter_FormatResult(t *testing.T) {
	f := &TextToNumberFilter{
		Numbers:      map[string]float64{"medio": 0.5},
		FormatResult: func(s string) string { return s + "x" },
	}
	require.Equal(t, "0.5x", f.ConvertTokens([]string{"medio"}))
}
