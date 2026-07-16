package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnitConvert(t *testing.T) {
	v, ok := Convert(100, UnitMile, UnitKilometre)
	require.True(t, ok)
	require.InDelta(t, 160.9344, v, 0.01)

	f, ok := Convert(32, UnitFahrenheit, UnitCelsius)
	require.True(t, ok)
	require.InDelta(t, 0, f, 0.01)

	_, ok = Convert(1, UnitMile, UnitKilogram)
	require.False(t, ok)
}

func TestAbstractUnitConversionRule_Match(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	sent := languagetool.AnalyzePlain("The trail is 10 mi long.")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.NotEmpty(t, ms[0].GetSuggestedReplacements())
	require.Contains(t, ms[0].GetSuggestedReplacements()[0], "km")
}
