package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestOrdinalSuffixFilter(t *testing.T) {
	f := NewOrdinalSuffixFilter()
	require.Equal(t, "1st", f.Fix("1nd"))
	require.Equal(t, "2nd", f.Fix("2th"))
	require.Equal(t, "3rd", f.Fix("3st"))
	require.Equal(t, "4th", f.Fix("4nd"))
	require.Equal(t, "11th", f.Fix("11st"))
	require.Equal(t, "12th", f.Fix("12nd"))
	require.Equal(t, "13th", f.Fix("13rd"))
	require.Equal(t, "21st", f.Fix("21nd"))
	require.Equal(t, "22nd", f.Fix("22th"))
	require.Equal(t, "23rd", f.Fix("23th"))
}

func TestOrdinalSuffixFilter_AcceptRuleMatch(t *testing.T) {
	f := NewOrdinalSuffixFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"1nd"})
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"1st"}, out.GetSuggestedReplacements())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, 3, "msg")
	m2.SetSuggestedReplacements([]string{"22th"})
	out = f.AcceptRuleMatch(m2, nil, 0, nil, nil)
	require.Equal(t, []string{"22nd"}, out.GetSuggestedReplacements())
}

func TestOrdinalSuffixFilter_EmptySuggestionsPanics(t *testing.T) {
	f := NewOrdinalSuffixFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, nil, 0, nil, nil)
	})
}

func TestOrdinalSuffixFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.OrdinalSuffixFilter"))
}
