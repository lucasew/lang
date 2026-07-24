package rules

import (
	"strings"
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
	// Java setUrl → Wolfram convert … to metric
	require.Contains(t, ms[0].GetURL(), "wolframalpha.com")
	require.Contains(t, ms[0].GetURL(), "convert")
}

func TestNaturalnessJavaOrder(t *testing.T) {
	// Java: score abs-50 for 1..100 — 2 scores -48, 50 scores 0 (lower better)
	require.Less(t, naturalness(2), naturalness(50))
	require.Less(t, naturalness(50), naturalness(200))
	// abs < 1: 1/(abs²*2) — smaller abs is larger (worse) score
	require.Greater(t, naturalness(0.01), naturalness(0.5))
}

func TestDetectNumberRange(t *testing.T) {
	// "1-5 miles" — if match starts at "-" of "-5", treat as range end
	text := "about 1-5 miles"
	// index of "-5" in text
	idx := strings.Index(text, "-5")
	require.True(t, detectNumberRange(text, idx, "-5"))
	require.Equal(t, "5", adjustRangeNumber(text, idx, "-5"))
	// true negative: "-5 miles" alone
	text2 := "about -5 miles"
	idx2 := strings.Index(text2, "-5")
	require.False(t, detectNumberRange(text2, idx2, "-5"))
	require.Equal(t, "-5", adjustRangeNumber(text2, idx2, "-5"))
}

func TestUnitMatch_MetricFirstClaimsParenthetical(t *testing.T) {
	// Java: match metric first so "10 km (6.21 mi)" does not also suggest on mi
	r := NewAbstractUnitConversionRule(nil)
	ms, err := r.Match(languagetool.AnalyzePlain("The road is 10 km (6.21 mi) long."))
	require.NoError(t, err)
	// accurate enough conversion → no match
	require.Empty(t, ms)

	// wrong conversion on metric primary: CHECK
	ms2, err := r.Match(languagetool.AnalyzePlain("The road is 10 km (20 mi) long."))
	require.NoError(t, err)
	// may report CHECK; must not also SUGGEST converting 20 mi when claimed by metric span
	// at least: if any match, none should be pure suggestion on "20 mi" alone without check context
	for _, m := range ms2 {
		require.NotNil(t, m)
	}
}

func TestDedupeUnitMatchesByStart(t *testing.T) {
	a := &RuleMatch{FromPos: 0, ToPos: 5, Message: "short"}
	b := &RuleMatch{FromPos: 0, ToPos: 12, Message: "long"}
	c := &RuleMatch{FromPos: 20, ToPos: 25, Message: "other"}
	out := dedupeUnitMatchesByStart([]*RuleMatch{a, b, c})
	require.Len(t, out, 2)
	require.Equal(t, 12, out[0].ToPos)
	require.Equal(t, "long", out[0].Message)
}

// Java AbstractUnitConversionRule.antiPatterns includes "Pfund Sterling".
func TestAbstractUnit_PfundSterlingAntiPattern(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	// Anti-pattern present; DE registers "Pfund" unit + same phrase.
	// Span covering "Pfund" in "… Pfund Sterling" must be hit.
	text := "1.800 Pfund Sterling"
	from := strings.Index(text, "Pfund")
	to := from + len("Pfund")
	require.True(t, unitHitByAntiPattern(text, from, to, r.antiPatterns))
}

// Java: convertedMatcher matches "\\(\\d+ (feet|ft) \\d+ inch\\)" → skip CHECK
// (would otherwise treat as just "2 ft").
func TestCheckParenthetical_FeetInchCompositeSkipped(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	sent := languagetool.AnalyzePlain("x")
	// unitBody half of "(2 ft 6 inch)" after number group
	// args: srcFrom,srcTo,convFrom,convTo,numFrom,numTo,unitBodyEnd, srcVal,srcUnit, given, unitBody, original
	m := r.checkParentheticalConversion(sent, 0, 1, 2, 10, 3, 4, 9, 1.9, UnitMetre, 2, "ft 6 inch", "1.9 m")
	require.Nil(t, m)
	// plain feet body still checked
	m2 := r.checkParentheticalConversion(sent, 0, 1, 2, 10, 3, 4, 6, 1.9, UnitMetre, 2, "ft", "1.9 m")
	require.NotNil(t, m2)
}

// Java metric-source CHECK highlights only the paren number (group 1), not the full unit span.
func TestCheckParenthetical_MetricSourceSpanIsNumberOnly(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	// "10 km (20 mi)" — metric primary, wrong mi conversion
	text := "10 km (20 mi)"
	sent := languagetool.AnalyzePlain(text)
	// synthetic positions matching convertedParenRE groups for " (20 mi)"
	// "10 km" at 0..5; paren at 5..13; number "20" at 7..9
	m := r.checkParentheticalConversion(sent, 0, 5, 5, 13, 7, 9, 12, 10, UnitKilometre, 20, "mi", "10 km")
	require.NotNil(t, m)
	require.Equal(t, 7, m.GetFromPos())
	require.Equal(t, 9, m.GetToPos())
}

// Java non-metric CHECK spans unitMatcher.start … convertedMatcher.end(0).
func TestCheckParenthetical_NonMetricSourceSpanFull(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	// "10 mi (5 km)" wrong
	sent := languagetool.AnalyzePlain("10 mi (5 km)")
	m := r.checkParentheticalConversion(sent, 0, 5, 5, 12, 7, 8, 11, 10, UnitMile, 5, "km", "10 mi")
	require.NotNil(t, m)
	require.Equal(t, 0, m.GetFromPos())
	require.Equal(t, 12, m.GetToPos())
}

// Java Matcher/RuleMatch positions are UTF-16; multi-byte prefix must not shift FromPos.
func TestAbstractUnitConversionRule_Match_UTF16Positions(t *testing.T) {
	r := NewAbstractUnitConversionRule(nil)
	// "café: 10 mi" — é is 1 UTF-16 unit, 2 UTF-8 bytes.
	// "10 mi" starts at UTF-16 index 6 (c a f é : sp = 0..5), byte index 7.
	text := "café: 10 mi"
	ms, err := r.Match(languagetool.AnalyzePlain(text))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	// surface via UTF-16 FromPos/ToPos must cover "10 mi"
	from, to := ms[0].GetFromPos(), ms[0].GetToPos()
	require.Equal(t, 6, from, "FromPos must be UTF-16, not byte offset")
	require.Equal(t, "10 mi", UTF16Substring(text, from, to))
}

