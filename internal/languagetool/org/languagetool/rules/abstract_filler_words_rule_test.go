package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractFillerWordsRule_DefaultMinPercent(t *testing.T) {
	r := &AbstractFillerWordsRule{
		AbstractStatisticStyleRule: &AbstractStatisticStyleRule{},
		FillerWords:                map[string]struct{}{"really": {}},
		Message:                    "filler",
	}
	InitFillerWordsMeta(r, nil, false)
	require.Equal(t, FillerWordsDefaultMinPercent, r.GetMinPercent())
	require.True(t, r.IsDefaultOff())

	// Short sentence: 1 filler / few words → over 8%
	matches := r.Match(languagetool.AnalyzePlain("He really came."))
	require.Equal(t, 1, len(matches))

	// Many words, one filler under 8%
	long := "He came to the house with friends and family for dinner and really enjoyed the evening together after work."
	// count words roughly > 12 with 1 filler → under 8%
	matches = r.Match(languagetool.AnalyzePlain(long))
	require.Equal(t, 0, len(matches))

	// MinPercent 0: show all
	r.SetMinPercent(0)
	matches = r.Match(languagetool.AnalyzePlain(long))
	require.Equal(t, 1, len(matches))
}
