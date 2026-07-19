package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishVariants(t *testing.T) {
	require.Equal(t, "English (US)", AmericanEnglish.GetName())
	require.Equal(t, []string{"US"}, AmericanEnglish.GetCountries())
	v, ok := EnglishVariantByCode("en-gb")
	require.True(t, ok)
	require.Equal(t, "en-GB", v.ShortCode)
	require.Len(t, AllEnglishVariants(), 6)
}

func TestEnglishRelevantRuleIDs(t *testing.T) {
	ids := EnglishRelevantRuleIDs()
	require.Contains(t, ids, "EN_A_VS_AN")
	require.Contains(t, ids, "EN_COMPOUNDS")
	require.Contains(t, ids, "READABILITY_RULE_DIFFICULT")
	require.Contains(t, ids, "READABILITY_RULE_SIMPLE")
	require.Contains(t, ids, "EN_REPEATEDWORDS")
	// OpenNMTRule commented out in Java
	require.NotContains(t, ids, "OPENNMT")
	us := AmericanEnglish.GetRelevantRuleIDs()
	require.Contains(t, us, "EN_US_SIMPLE_REPLACE")
	require.Contains(t, us, "METRIC_UNITS_EN_US")
	require.Equal(t, len(ids)+2, len(us))
	gb := BritishEnglish.GetRelevantRuleIDs()
	require.Contains(t, gb, "EN_GB_SIMPLE_REPLACE")
	require.Contains(t, gb, "METRIC_UNITS_EN_IMPERIAL")
	require.Equal(t, len(ids)+2, len(gb))
	// Canadian / Australian: Imperial unit conversion only
	ca := CanadianEnglish.GetRelevantRuleIDs()
	require.Contains(t, ca, "METRIC_UNITS_EN_IMPERIAL")
	require.Equal(t, len(ids)+1, len(ca))
	// NZ: replace + Imperial
	nz := NewZealandEnglish.GetRelevantRuleIDs()
	require.Contains(t, nz, "EN_NZ_SIMPLE_REPLACE")
	require.Contains(t, nz, "METRIC_UNITS_EN_IMPERIAL")
	require.Equal(t, len(ids)+2, len(nz))
	// ZA: super only
	require.Equal(t, ids, SouthAfricanEnglish.GetRelevantRuleIDs())
}
