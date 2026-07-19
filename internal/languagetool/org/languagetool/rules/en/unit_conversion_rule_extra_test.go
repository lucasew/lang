package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRule_IDsAndImperialUS(t *testing.T) {
	gen := NewUnitConversionRule(nil)
	require.Equal(t, "METRIC_UNITS_EN_GENERAL", gen.GetID())
	// Java UnitConversionRule.setTags(Tag.picky)
	require.True(t, gen.HasTag(rules.TagPicky))
	imp := NewUnitConversionRuleImperial(nil)
	us := NewUnitConversionRuleUS(nil)
	require.Equal(t, "METRIC_UNITS_EN_IMPERIAL", imp.GetID())
	require.Equal(t, "METRIC_UNITS_EN_US", us.GetID())
	require.True(t, imp.HasTag(rules.TagPicky))
	require.True(t, us.HasTag(rules.TagPicky))

	// Imperial pints still suggest metric
	ms := imp.Match(languagetool.AnalyzePlain("I just drank 3 pints."))
	require.NotEmpty(t, ms)
	// US gallons
	ms2 := us.Match(languagetool.AnalyzePlain("The tank holds 10 gallons."))
	require.NotEmpty(t, ms2)
	// cubic metres metric path smoke
	_ = imp.Match(languagetool.AnalyzePlain("The volume is 2 cubic metres."))
	// ounces mass (general) still fires
	ms3 := NewUnitConversionRule(nil).Match(languagetool.AnalyzePlain("Weighs 8 ounces."))
	require.NotEmpty(t, ms3)
	joined := strings.Join(ms3[0].GetSuggestedReplacements(), " ")
	// should suggest grams/kilograms not invent empty
	require.NotEmpty(t, joined)
}
