package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmallLangConstructors(t *testing.T) {
	require.Equal(t, "km", NewKhmer().ShortCode)
	require.Equal(t, "ml", NewMalayalam().ShortCode)
	require.Equal(t, "tl", NewTagalog().ShortCode)
	require.Equal(t, "ta", Tamil.ShortCode)
	require.Equal(t, "lt", Lithuanian.ShortCode)
	require.Equal(t, "is", Icelandic.ShortCode)
	all := AllExtendedSmallLangs()
	require.GreaterOrEqual(t, len(all), 10)
	seen := map[string]bool{}
	// Java modules without createDefaultSpellingRule / no speller in getRelevantRules
	noDefaultSpeller := map[string]bool{"ja": true, "zh": true, "fa": true, "ta": true}
	for _, l := range all {
		require.False(t, seen[l.ShortCode], l.ShortCode)
		seen[l.ShortCode] = true
		require.NotEmpty(t, l.Name)
		if noDefaultSpeller[l.ShortCode] {
			require.Empty(t, l.SpellerRuleID, l.ShortCode)
		} else {
			require.NotEmpty(t, l.SpellerRuleID, l.ShortCode)
		}
	}
}
