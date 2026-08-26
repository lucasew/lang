package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRomanNumeralFilter(t *testing.T) {
	f := NewRomanNumeralFilter()
	require.Equal(t, "I", f.Suggest("1"))
	require.Equal(t, "IV", f.Suggest("4"))
	require.Equal(t, "IX", f.Suggest("9"))
	require.Equal(t, "XLII", f.Suggest("42"))
	require.Equal(t, "MCMXCIX", f.Suggest("1999"))
	require.Equal(t, "MMXXIV", f.Suggest("2024"))
	require.Equal(t, "", f.Suggest("0"))
	require.Equal(t, "", f.Suggest("abc"))
}

func TestRomanNumeralFilter_AcceptRuleMatch(t *testing.T) {
	f := NewRomanNumeralFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"arabicSource": "2024"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"MMXXIV"}, out.GetSuggestedReplacements())
}

func TestPTRomanNumeralFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.RomanNumeralFilter"))
}
