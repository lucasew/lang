package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

func TestTextToNumberFilter_AcceptRuleMatch(t *testing.T) {
	f := &TextToNumberFilter{
		Numbers:     map[string]float64{"dos": 2},
		Multipliers: map[string]float64{"mil": 1000},
	}
	tok1 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("dos", nil, nil), 0)
	tok2 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("mil", nil, nil), 4)
	m := NewRuleMatch(nil, nil, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{tok1, tok2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"2000"}, out.GetSuggestedReplacements())
	// appends to existing suggestions
	m2 := NewRuleMatch(nil, nil, 0, 3, "msg")
	m2.SetSuggestedReplacements([]string{"prev"})
	out2 := f.AcceptRuleMatch(m2, nil, 0, []*languagetool.AnalyzedTokenReadings{tok1}, nil)
	require.Equal(t, []string{"prev", "2"}, out2.GetSuggestedReplacements())
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
}
