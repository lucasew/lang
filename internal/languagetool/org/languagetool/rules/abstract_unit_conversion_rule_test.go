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

func TestAbstractUnitConversionRule_FormatRounded(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	// default Java formatRounded prefix
	require.Equal(t, "ca. 2 m", r.formatRounded("2 m"))
	// near-integer (within ROUNDING_DELTA 0.05) emits rounded + exact forms
	forms := r.formatConversionSuggestion(1.98, "m")
	require.Contains(t, forms, "ca. 2 m")
	require.Contains(t, forms, "1.98 m")
}

func TestAbstractUnitConversionRule_ShortMessage(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	require.Equal(t, "Add metric equivalent?", r.GetShortMessage(UnitMsgSuggestion))
	require.Equal(t, "Incorrect unit conversion. Correct it?", r.GetShortMessage(UnitMsgCheck))
	// Match sets short message on RuleMatch
	ms, err := r.Match(languagetool.AnalyzePlain("The trail is 10 mi long."))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Equal(t, "Add metric equivalent?", ms[0].GetShortMessage())
}
